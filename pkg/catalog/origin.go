package catalog

import (
	"fmt"
	"log"
	"time"

	"github.com/facette/facette/pkg/config"
	"github.com/facette/facette/pkg/connector"
)

// Origin represents an origin of source sets (e.g. a Collectd or Graphite instance).
type Origin struct {
	Name            string
	Connector       connector.Connector
	Sources         map[string]*Source
	RefreshInterval int
	LastRefresh     time.Time
	Catalog         *Catalog
	controlChan     chan OriginWorkerCmd
	connectorChan   chan [2]string
}

// OriginWorkerCmd represents a command to an origin worker
type OriginWorkerCmd struct {
	Command int
	Err     chan error
}

// NewOrigin creates a new origin instance.
func NewOrigin(name string, config *config.OriginConfig) (*Origin, error) {
	connectorType, ok := config.Connector["type"].(string)
	if !ok {
		return nil, fmt.Errorf("connector type should be a string")
	} else if connectorType == "" {
		return nil, fmt.Errorf("missing connector type")
	} else if _, ok := connector.Connectors[connectorType]; !ok {
		return nil, fmt.Errorf("unknown connector type `%s'", connectorType)
	}

	origin := &Origin{
		Name:            name,
		Sources:         make(map[string]*Source),
		RefreshInterval: config.RefreshInterval,
		controlChan:     make(chan OriginWorkerCmd),
	}

	originConnector, err := connector.Connectors[connectorType](&origin.connectorChan, config.Connector)
	if err != nil {
		return nil, err
	}

	origin.Connector = originConnector.(connector.Connector)

	// Start origin worker goroutine in background
	go originWorker(origin)

	return origin, nil
}

// Refresh updates the current origin by querying its connector for sources and metrics.
func (origin *Origin) Refresh() error {
	if origin.Connector == nil {
		return fmt.Errorf("connector for origin `%s' is not initialized", origin.Name)
	}

	if origin.Catalog.debugLevel > 0 {
		log.Printf("DEBUG: refreshing origin `%s'", origin.Name)
	}

	origin.Sources = make(map[string]*Source)

	// Origin input channel
	origin.connectorChan = make(chan [2]string)

	// Channel to be notified in case of connector refresh error
	connectorErrChan := make(chan error)

	go origin.Connector.Refresh(connectorErrChan)

	for {
		select {
		case err := <-connectorErrChan:
			// An error occurred while connector refreshed orgin
			return err

		case entry, ok := <-origin.connectorChan:
			// Channel is closed: connector is done refreshing origin
			if !ok {
				goto done
			}

			originalSource, originalMetric := entry[0], entry[1]

			for _, filter := range origin.Catalog.Config.Origins[origin.Name].Filters {
				if filter.Target != "source" && filter.Target != "metric" && filter.Target != "" {
					log.Printf("ERROR: unknown `%s' filter target", filter.Target)
					continue
				}

				if (filter.Target == "source" || filter.Target == "") && filter.PatternRegexp.MatchString(entry[0]) {
					if filter.Discard {
						goto nextEntry
					}

					entry[0] = filter.PatternRegexp.ReplaceAllString(entry[0], filter.Rewrite)
				}

				if (filter.Target == "metric" || filter.Target == "") && filter.PatternRegexp.MatchString(entry[1]) {
					if filter.Discard {
						goto nextEntry
					}

					entry[1] = filter.PatternRegexp.ReplaceAllString(entry[1], filter.Rewrite)
				}
			}

			if _, ok := origin.Sources[entry[0]]; !ok {
				origin.Sources[entry[0]] = NewSource(entry[0], originalSource, origin)
			}

			if origin.Catalog.debugLevel > 3 {
				log.Printf("DEBUG: appending metric `%s' to source `%s'", entry[1], entry[0])
			}

			origin.Sources[entry[0]].Metrics[entry[1]] = NewMetric(entry[1], originalMetric, origin.Sources[entry[0]])
		}

	nextEntry:
	}

done:
	origin.LastRefresh = time.Now()

	return nil
}

// originWorker starts a new origin worker goroutine
func originWorker(origin *Origin) {
	var (
		selfRefreshTimeticker *time.Ticker
		selfRefreshTimerChan  <-chan time.Time
	)

	defer close(origin.controlChan)

	// If origin "self refresh" has been configured, set up a time ticker
	if origin.RefreshInterval > 0 {
		selfRefreshTimeticker = time.NewTicker(time.Duration(origin.RefreshInterval) * time.Second)
		selfRefreshTimerChan = selfRefreshTimeticker.C
	}

	for {
		select {
		case _ = <-selfRefreshTimerChan:
			// Periodic origin refresh triggered
			if origin.Catalog.debugLevel > 1 {
				log.Printf("DEBUG: periodic refresh of origin `%s' triggered", origin.Name)
			}

			if err := origin.Refresh(); err != nil {
				log.Printf("ERROR: unable to refresh origin `%s': %s", origin.Name, err)
			}

			origin.LastRefresh = time.Now()

		case cmd := <-origin.controlChan:
			// Control command received
			switch cmd.Command {
			case OriginCmdRefresh:
				// Explicit origin refresh triggered
				if err := origin.Refresh(); err != nil {
					cmd.Err <- fmt.Errorf("unable to refresh origin `%s': %s", origin.Name, err)
					continue
				}

				origin.LastRefresh = time.Now()
				cmd.Err <- nil

			case OriginCmdShutdown:
				// Global shutdown triggered
				if selfRefreshTimerChan != nil {
					// Stop "self refresh" time ticker
					selfRefreshTimeticker.Stop()
				}

				cmd.Err <- nil
				return

			default:
				// Unsupported command
				cmd.Err <- fmt.Errorf("unsupported command received")
				continue
			}
		}
	}
}

// SendOriginWorkerCmd sends a command to an origin worker goroutine
func SendOriginWorkerCmd(origin *Origin, cmd int) error {
	errChan := make(chan error)

	origin.controlChan <- OriginWorkerCmd{
		Command: cmd,
		Err:     errChan,
	}

	return <-errChan
}

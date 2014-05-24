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
	Name          string
	Connector     connector.Connector
	Sources       map[string]*Source
	SelfRefresh   int
	LastRefresh   time.Time
	Catalog       *Catalog
	connectorChan chan [2]string
}

// NewOrigin creates a new origin instance.
func NewOrigin(name string, config *config.OriginConfig) (*Origin, error) {
	connectorType := config.ConnectorSettings["type"]

	if _, ok := config.ConnectorSettings["type"]; !ok {
		return nil, fmt.Errorf("missing connector type")
	} else if _, ok := connector.Connectors[connectorType]; !ok {
		return nil, fmt.Errorf("unknown connector type `%s'", connectorType)
	}

	origin := &Origin{
		Name:        name,
		Sources:     make(map[string]*Source),
		SelfRefresh: config.SelfRefresh,
	}

	originConnector, err := connector.Connectors[connectorType](
		&origin.connectorChan,
		config.ConnectorSettings)
	if err != nil {
		return nil, err
	}

	origin.Connector = originConnector.(connector.Connector)

	return origin, nil
}

// Refresh updates the current origin by querying its connector for sources and metrics.
func (origin *Origin) Refresh() error {
	if origin.Connector == nil {
		return fmt.Errorf("connector for origin `%s' is not initialized", origin.Name)
	}

	if origin.Catalog.debugLevel > 0 {
		log.Printf("DEBUG: refreshing origin `%s'...\n", origin.Name)
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
				log.Printf("DEBUG: appending `%s' metric for `%s' source...\n", entry[1], entry[0])
			}

			origin.Sources[entry[0]].Metrics[entry[1]] = NewMetric(entry[1], originalMetric, origin.Sources[entry[0]])
		}

	nextEntry:
	}

done:
	origin.LastRefresh = time.Now()

	return nil
}

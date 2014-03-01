package catalog

import (
	"fmt"
	"log"
	"sync"

	"github.com/facette/facette/pkg/connector"
)

// Origin represents an origin of source sets (e.g. a Collectd or Graphite instance).
type Origin struct {
	Name      string
	Connector connector.Connector
	Sources   map[string]*Source
	Catalog   *Catalog
	inputChan chan [2]string
}

// Refresh updates the current origin by querying its connector for sources and metrics.
func (origin *Origin) Refresh(wait *sync.WaitGroup) error {
	if origin.Connector == nil {
		return fmt.Errorf("connector for `%s' origin is not initialized", origin.Name)
	}

	if origin.Catalog.debugLevel > 1 {
		log.Printf("DEBUG: updating origin `%s'...\n", origin.Name)
	}

	origin.Sources = make(map[string]*Source)

	// Create update channel
	origin.inputChan = make(chan [2]string)

	wait.Add(1)

	go func() {
		defer wait.Done()

		for entry := range origin.inputChan {
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

			if origin.Catalog.debugLevel > 2 {
				log.Printf("DEBUG: appending `%s' metric for `%s' source...\n", entry[1], entry[0])
			}

			origin.Sources[entry[0]].Metrics[entry[1]] = NewMetric(entry[1], originalMetric, origin.Sources[entry[0]])

		nextEntry:
		}
	}()

	return origin.Connector.Refresh()
}

// NewOrigin creates a new origin instance.
func NewOrigin(name string, config map[string]string, catalog *Catalog) (*Origin, error) {
	if _, ok := config["type"]; !ok {
		return nil, fmt.Errorf("missing connector type")
	} else if _, ok := connector.Connectors[config["type"]]; !ok {
		return nil, fmt.Errorf("unknown `%s' connector type", config["type"])
	}

	origin := &Origin{
		Name:    name,
		Sources: make(map[string]*Source),
		Catalog: catalog,
	}

	handler, err := connector.Connectors[config["type"]](&origin.inputChan, config)
	if err != nil {
		return nil, err
	}

	origin.Connector = handler.(connector.Connector)

	catalog.Origins[name] = origin

	return origin, nil
}

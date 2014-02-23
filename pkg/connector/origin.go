package connector

import (
	"fmt"
	"log"
)

// A Origin represents an origin entry.
type Origin struct {
	Name      string
	Connector ConnectorHandler
	Sources   map[string]*Source
	catalog   *Catalog
	inputChan chan [2]string
}

// AppendSource adds a new Source entry into the Origin instance.
func (origin *Origin) AppendSource(name, origName string) *Source {
	if origin.catalog.debugLevel > 2 {
		log.Printf("DEBUG: appending `%s' source into origin...\n", name)
	}

	// Append new source instance into origin
	source := &Source{Name: name, OriginalName: origName, Metrics: make(map[string]*Metric), origin: origin}
	origin.Sources[name] = source

	return source
}

// Update updates the current Origin by parsing the filesystem for sources or metrics.
func (origin *Origin) Update() error {
	if origin.catalog.debugLevel > 1 {
		log.Printf("DEBUG: updating origin `%s'...\n", origin.Name)
	}

	origin.Sources = make(map[string]*Source)

	if origin.Connector == nil {
		return fmt.Errorf("connector for `%s' origin is not initialized", origin.Name)
	}

	// Create update channel
	origin.inputChan = make(chan [2]string)

	go func() {
		for entry := range origin.inputChan {
			originalSource, originalMetric := entry[0], entry[1]

			for _, filter := range origin.catalog.Config.Origins[origin.Name].Filters {
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
				origin.AppendSource(entry[0], originalSource)
			}

			origin.Sources[entry[0]].AppendMetric(entry[1], originalMetric)

		nextEntry:
		}
	}()

	return origin.Connector.Update()
}

package backend

import (
	"facette/common"
	"fmt"
	"log"
	"time"
)

// Catalog represents the main structure of running Facette's instance (e.g. origins, sources, metrics).
type Catalog struct {
	Config     *common.Config
	Origins    map[string]*Origin
	Updated    time.Time
	debugLevel int
}

// AddOrigin adds a new Origin entry into the Catalog instance.
func (catalog *Catalog) AddOrigin(name string, config map[string]string) (*Origin, error) {
	var (
		err    error
		origin *Origin
	)

	if _, ok := config["type"]; !ok {
		return nil, fmt.Errorf("missing backend type")
	} else if _, ok := BackendHandlers[config["type"]]; !ok {
		return nil, fmt.Errorf("unknown `%s' backend type", config["type"])
	}

	origin = &Origin{Name: name, Sources: make(map[string]*Source), catalog: catalog}
	origin.inputChan = make(chan [2]string)

	go func() {
		var (
			originalSource string
			originalMetric string
		)

		for entry := range origin.inputChan {
			originalSource, originalMetric = entry[0], entry[1]

			for _, filter := range catalog.Config.Origins[name].Filters {
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

	if err = BackendHandlers[config["type"]](origin, config); err != nil {
		return nil, err
	}

	catalog.Origins[name] = origin

	return origin, nil
}

// GetMetric returns an existing Metric entry based on its origin, source and name.
func (catalog *Catalog) GetMetric(origin, source, name string) *Metric {
	if !catalog.MetricExists(origin, source, name) {
		return nil
	}

	return catalog.Origins[origin].Sources[source].Metrics[name]
}

// MetricExists returns whether a metric exists or not.
func (catalog *Catalog) MetricExists(origin, source, name string) bool {
	if _, ok := catalog.Origins[origin]; ok {
		if _, ok := catalog.Origins[origin].Sources[source]; ok {
			if _, ok := catalog.Origins[origin].Sources[source].Metrics[name]; ok {
				return true
			}
		}
	}

	return false
}

// Update updates the current Catalog by updating its origins.
func (catalog *Catalog) Update() error {
	var (
		err     error
		success bool
	)

	success = true

	log.Println("INFO: catalog update started")

	// Update catalog origins
	for _, origin := range catalog.Origins {
		if err = origin.Update(); err != nil {
			log.Println("ERROR: " + err.Error())
			success = false
		}
	}

	// Handle output information
	if !success {
		log.Println("INFO: catalog update failed")
		return err
	}

	catalog.Updated = time.Now()

	log.Println("INFO: catalog update completed")
	return nil
}

// NewCatalog creates a new instance of Catalog.
func NewCatalog(config *common.Config, debugLevel int) *Catalog {
	// Create new Catalog instance
	return &Catalog{Config: config, Origins: make(map[string]*Origin), debugLevel: debugLevel}
}

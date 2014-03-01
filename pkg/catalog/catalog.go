// Package catalog implements the service catalog handling immutable data: origins, sources and metrics.
package catalog

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/facette/facette/pkg/config"
)

// Catalog represents the main structure of a catalog instance.
type Catalog struct {
	Config     *config.Config
	Origins    map[string]*Origin
	Updated    time.Time
	debugLevel int
}

// GetMetric returns an existing Metric entry based on its origin, source and name.
func (catalog *Catalog) GetMetric(origin, source, name string) *Metric {
	if _, ok := catalog.Origins[origin]; !ok {
		return nil
	} else if _, ok := catalog.Origins[origin].Sources[source]; !ok {
		return nil
	} else if _, ok := catalog.Origins[origin].Sources[source].Metrics[name]; !ok {
		return nil
	}

	return catalog.Origins[origin].Sources[source].Metrics[name]
}

// Refresh updates the current Catalog by refreshing its origins.
func (catalog *Catalog) Refresh() error {
	success := true

	log.Println("INFO: catalog refresh started")

	// Get origins from configuration
	catalog.Origins = make(map[string]*Origin)

	for originName, originConfig := range catalog.Config.Origins {
		origin, err := NewOrigin(originName, originConfig.Connector, catalog)
		if err != nil {
			log.Printf("ERROR: %s\n", err.Error())
		}

		catalog.Origins[originName] = origin
	}

	// Update catalog origins
	wait := &sync.WaitGroup{}

	for _, origin := range catalog.Origins {
		if err := origin.Refresh(wait); err != nil {
			log.Println("ERROR: " + err.Error())
			success = false
		}
	}

	wait.Wait()

	// Handle output information
	if !success {
		log.Println("INFO: catalog refresh failed")
		return fmt.Errorf("unable to refresh catalog")
	}

	catalog.Updated = time.Now()

	log.Println("INFO: catalog refresh completed")

	return nil
}

// NewCatalog creates a new instance of Catalog.
func NewCatalog(config *config.Config, debugLevel int) *Catalog {
	return &Catalog{
		Config:     config,
		Origins:    make(map[string]*Origin),
		debugLevel: debugLevel,
	}
}

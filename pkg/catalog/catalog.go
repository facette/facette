// Package catalog implements the service catalog handling immutable data: origins, sources and metrics.
package catalog

import (
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

const (
	_ = iota
	// OriginCmdRefresh represents an origin refresh command
	OriginCmdRefresh
	// OriginCmdShutdown represents an origin shutdown command
	OriginCmdShutdown
)

// NewCatalog creates a new instance of catalog.
func NewCatalog(config *config.Config, debugLevel int) *Catalog {
	return &Catalog{
		Config:     config,
		Origins:    make(map[string]*Origin),
		debugLevel: debugLevel,
	}
}

// GetMetric returns an existing metric entry based on its origin, source and name.
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

// Refresh updates the current catalog by refreshing its origins.
func (catalog *Catalog) Refresh() error {
	var (
		origin *Origin
		err    error
	)

	log.Println("INFO: catalog refresh started")

	// Get origins from configuration
	catalog.Origins = make(map[string]*Origin)

	for originName, originConfig := range catalog.Config.Origins {
		origin, err = NewOrigin(originName, originConfig)
		if err != nil {
			log.Printf("ERROR: in origin `%s', %s", originName, err.Error())
			log.Printf("WARNING: discarding origin `%s'", originName)
			continue
		}

		origin.Catalog = catalog

		catalog.Origins[originName] = origin
	}

	wait := &sync.WaitGroup{}

	// Update catalog origins concurrently
	for _, origin = range catalog.Origins {
		wait.Add(1)

		go func(wg *sync.WaitGroup, origin *Origin) {
			defer wg.Done()

			if err = SendOriginWorkerCmd(origin, OriginCmdRefresh); err != nil {
				log.Println("ERROR: " + err.Error())
			}
		}(wait, origin)
	}

	// Wait for all origins to be refreshed
	wait.Wait()

	catalog.Updated = time.Now()

	log.Println("INFO: catalog refresh completed")

	return nil
}

// Close terminates all origin workers and performs catalog clean-up
func (catalog *Catalog) Close() error {
	var err error

	// Shutdown catalog origin workers
	for _, origin := range catalog.Origins {
		if err = SendOriginWorkerCmd(origin, OriginCmdShutdown); err != nil {
			log.Printf("ERROR: unable to shut down origin `%s' worker: %s", origin.Name, err)
		}
	}

	log.Println("INFO: catalog closed")

	return nil
}

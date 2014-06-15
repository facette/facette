// Package catalog implements the service catalog handling immutable data: origins, sources and metrics.
package catalog

import (
	"fmt"
	"log"

	"github.com/facette/facette/pkg/config"
)

// Catalog represents the main structure of a catalog instance.
type Catalog struct {
	Config     *config.Config
	Origins    map[string]*Origin
	RecordChan chan *CatalogRecord
	debugLevel int // TODO: remove this
}

// CatalogRecord represents a catalog record.
type CatalogRecord struct {
	Origin    string
	Source    string
	Metric    string
	Connector interface{}
}

func (r CatalogRecord) String() string {
	return fmt.Sprintf("{Origin: \"%s\", Source: \"%s\", Metric: \"%s\"}", r.Origin, r.Source, r.Metric)
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
		RecordChan: make(chan *CatalogRecord),
		debugLevel: debugLevel,
	}
}

// Insert inserts a new record in the catalog.
func (catalog *Catalog) Insert(record *CatalogRecord) {
	if catalog.debugLevel > 3 {
		log.Printf("DEBUG: appending metric `%s' to source `%s' via origin `%s'", record.Metric, record.Source,
			record.Origin)
	}

	if _, ok := catalog.Origins[record.Origin]; !ok {
		catalog.Origins[record.Origin] = NewOrigin(
			record.Origin,
			nil,
			catalog,
		)
	}

	if _, ok := catalog.Origins[record.Origin].Sources[record.Source]; !ok {
		catalog.Origins[record.Origin].Sources[record.Source] = NewSource(
			record.Source,
			catalog.Origins[record.Origin],
		)
	}

	if _, ok := catalog.Origins[record.Origin].Sources[record.Source].Metrics[record.Metric]; !ok {
		catalog.Origins[record.Origin].Sources[record.Source].Metrics[record.Metric] = NewMetric(
			record.Metric,
			catalog.Origins[record.Origin].Sources[record.Source],
			record.Connector,
		)
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

// Close terminates all origin workers and performs catalog clean-up
func (catalog *Catalog) Close() error {
	close(catalog.RecordChan)

	return nil
}

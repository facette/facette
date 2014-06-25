// Package catalog implements the service catalog handling immutable data: origins, sources and metrics.
package catalog

import (
	"fmt"

	"github.com/facette/facette/pkg/logger"
)

const (
	_ = iota
	// OriginCmdRefresh represents an origin refresh command
	OriginCmdRefresh
	// OriginCmdShutdown represents an origin shutdown command
	OriginCmdShutdown
)

// Catalog represents the main structure of a catalog instance.
type Catalog struct {
	Origins    map[string]*Origin
	RecordChan chan *CatalogRecord
}

// CatalogRecord represents a catalog record.
type CatalogRecord struct {
	Origin         string
	Source         string
	Metric         string
	OriginalOrigin string
	OriginalSource string
	OriginalMetric string
	Connector      interface{}
}

func (r CatalogRecord) String() string {
	return fmt.Sprintf("{Origin: \"%s\", Source: \"%s\", Metric: \"%s\"}", r.Origin, r.Source, r.Metric)
}

// NewCatalog creates a new instance of catalog.
func NewCatalog() *Catalog {
	return &Catalog{
		Origins:    make(map[string]*Origin),
		RecordChan: make(chan *CatalogRecord),
	}
}

// Insert inserts a new record in the catalog.
func (catalog *Catalog) Insert(record *CatalogRecord) {
	logger.Log(
		logger.LevelDebug,
		"catalog",
		"appending metric `%s' to source `%s' via origin `%s'",
		record.Metric,
		record.Source,
		record.Origin,
	)

	if _, ok := catalog.Origins[record.Origin]; !ok {
		catalog.Origins[record.Origin] = NewOrigin(
			record.Origin,
			record.OriginalOrigin,
			catalog,
		)
	}

	if _, ok := catalog.Origins[record.Origin].Sources[record.Source]; !ok {
		catalog.Origins[record.Origin].Sources[record.Source] = NewSource(
			record.Source,
			record.OriginalSource,
			catalog.Origins[record.Origin],
		)
	}

	if _, ok := catalog.Origins[record.Origin].Sources[record.Source].Metrics[record.Metric]; !ok {
		catalog.Origins[record.Origin].Sources[record.Source].Metrics[record.Metric] = NewMetric(
			record.Metric,
			record.OriginalMetric,
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

// Close closes a catalog instance.
func (catalog *Catalog) Close() error {
	close(catalog.RecordChan)

	return nil
}

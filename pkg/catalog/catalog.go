// Package catalog implements the service catalog handling immutable data: origins, sources and metrics.
package catalog

import (
	"fmt"
	"sync"

	"github.com/facette/facette/pkg/logger"
)

// Catalog represents the main structure of a catalog instance.
type Catalog struct {
	RecordChan chan *Record
	origins    map[string]*Origin
	sync.RWMutex
}

// Record represents a catalog record.
type Record struct {
	Origin         string
	Source         string
	Metric         string
	OriginalOrigin string
	OriginalSource string
	OriginalMetric string
	Connector      interface{}
}

func (r Record) String() string {
	return fmt.Sprintf("{Origin: %q, Source: %q, Metric: %q}", r.Origin, r.Source, r.Metric)
}

// NewCatalog creates a new instance of catalog.
func NewCatalog() *Catalog {
	return &Catalog{
		RecordChan: make(chan *Record),
		origins:    make(map[string]*Origin),
	}
}

// Insert inserts a new record in the catalog.
func (c *Catalog) Insert(record *Record) {
	c.Lock()
	defer c.Unlock()

	logger.Log(
		logger.LevelDebug,
		"catalog",
		"appending metric `%s' to source `%s' via origin `%s'",
		record.Metric,
		record.Source,
		record.Origin,
	)

	if _, ok := c.origins[record.Origin]; !ok {
		c.origins[record.Origin] = NewOrigin(
			record.Origin,
			record.OriginalOrigin,
			c,
		)
	}

	if _, ok := c.origins[record.Origin].sources[record.Source]; !ok {
		c.origins[record.Origin].sources[record.Source] = NewSource(
			record.Source,
			record.OriginalSource,
			c.origins[record.Origin],
		)
	}

	if _, ok := c.origins[record.Origin].sources[record.Source].metrics[record.Metric]; !ok {
		c.origins[record.Origin].sources[record.Source].metrics[record.Metric] = NewMetric(
			record.Metric,
			record.OriginalMetric,
			c.origins[record.Origin].sources[record.Source],
			record.Connector,
		)
	}
}

// Close closes a catalog instance.
func (c *Catalog) Close() error {
	close(c.RecordChan)

	return nil
}

// OriginExists returns whether an origin exists for the catalog based on its name.
func (c *Catalog) OriginExists(name string) bool {
	c.RLock()
	defer c.RUnlock()

	_, ok := c.origins[name]
	return ok
}

// GetOrigin returns an existing origin entry based on its name.
func (c *Catalog) GetOrigin(name string) (*Origin, error) {
	c.RLock()
	defer c.RUnlock()

	if !c.OriginExists(name) {
		return nil, fmt.Errorf("unknown origin `%s'", name)
	}

	return c.origins[name], nil
}

// GetOrigins returns a slice of origins.
func (c *Catalog) GetOrigins() []*Origin {
	c.RLock()
	defer c.RUnlock()

	items := make([]*Origin, 0)
	for _, o := range c.origins {
		items = append(items, o)
	}

	return items
}

// GetSource returns an existing source entry based on its origin and name.
func (c *Catalog) GetSource(origin, name string) (*Source, error) {
	c.RLock()
	defer c.RUnlock()

	o, err := c.GetOrigin(origin)
	if err != nil {
		return nil, err
	}

	return o.GetSource(name)
}

// GetMetric returns an existing metric entry based on its origin, source and name.
func (c *Catalog) GetMetric(origin, source, name string) (*Metric, error) {
	c.RLock()
	defer c.RUnlock()

	s, err := c.GetSource(origin, source)
	if err != nil {
		return nil, err
	}

	return s.GetMetric(name)
}

package catalog

import (
	"fmt"
)

// Origin represents an origin of source sets (e.g. a Collectd or Graphite instance).
type Origin struct {
	Name         string
	OriginalName string
	sources      map[string]*Source
	catalog      *Catalog
}

// NewOrigin creates a new origin instance.
func NewOrigin(name, origName string, catalog *Catalog) *Origin {
	return &Origin{
		Name:         name,
		OriginalName: origName,
		sources:      make(map[string]*Source),
		catalog:      catalog,
	}
}

// SourceExists returns whether a source exists for the origin based on its name.
func (o *Origin) SourceExists(name string) bool {
	o.catalog.RLock()
	defer o.catalog.RUnlock()

	_, ok := o.sources[name]
	return ok
}

// GetSource returns an existing origin source entry based on its name.
func (o *Origin) GetSource(name string) (*Source, error) {
	o.catalog.RLock()
	defer o.catalog.RUnlock()

	s, ok := o.sources[name]
	if !ok {
		return nil, fmt.Errorf("unknown source `%s' for origin `%s'", name, o.Name)
	}

	return s, nil
}

// GetSources returns a slice of sources for the origin.
func (o *Origin) GetSources() []*Source {
	o.catalog.RLock()
	defer o.catalog.RUnlock()

	items := make([]*Source, 0)
	for _, s := range o.sources {
		items = append(items, s)
	}

	return items
}

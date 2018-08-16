package catalog

import "facette.io/maputil"

// Metric represents a catalog metric instance.
type Metric struct {
	Name       string
	Attributes *maputil.Map
	source     *Source
}

// Catalog returns the parent catalog of the origin.
func (m *Metric) Catalog() *Catalog {
	return m.source.origin.catalog
}

// Origin returns the parent origin of the source.
func (m *Metric) Origin() *Origin {
	return m.source.origin
}

// Source returns the parent origin of the source.
func (m *Metric) Source() *Source {
	return m.source
}

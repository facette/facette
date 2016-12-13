package catalog

// Metric represents a catalog metric instance.
type Metric struct {
	Name         string
	OriginalName string
	source       *Source
	connector    interface{}
}

// Source returns the parent source from the catalog metric.
func (m *Metric) Source() *Source {
	m.source.origin.catalog.RLock()
	defer m.source.origin.catalog.RUnlock()

	return m.source
}

// Connector returns the connector handler associated to the catalog metric.
func (m *Metric) Connector() interface{} {
	m.source.origin.catalog.RLock()
	defer m.source.origin.catalog.RUnlock()

	return m.connector
}

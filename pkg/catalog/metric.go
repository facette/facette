package catalog

// Metric represents a metric entry.
type Metric struct {
	Name         string
	OriginalName string
	source       *Source
	connector    interface{}
}

// NewMetric creates a new metric instance.
func NewMetric(name, origName string, source *Source, connector interface{}) *Metric {
	return &Metric{
		Name:         name,
		OriginalName: origName,
		source:       source,
		connector:    connector,
	}
}

// GetSource returns the parent source of the metric.
func (m *Metric) GetSource() *Source {
	m.source.origin.catalog.RLock()
	defer m.source.origin.catalog.RUnlock()

	return m.source
}

// GetConnector returns the connector associated with the metric.
func (m *Metric) GetConnector() interface{} {
	m.source.origin.catalog.RLock()
	defer m.source.origin.catalog.RUnlock()

	return m.connector
}

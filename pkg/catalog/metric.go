package catalog

// Metric represents a metric entry.
type Metric struct {
	Name      string
	Source    *Source
	Connector interface{}
}

// NewMetric creates a new metric instance.
func NewMetric(name string, source *Source, connector interface{}) *Metric {
	return &Metric{
		Name:      name,
		Source:    source,
		Connector: connector,
	}
}

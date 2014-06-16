package catalog

// Metric represents a metric entry.
type Metric struct {
	Name         string
	OriginalName string
	Source       *Source
	Connector    interface{}
}

// NewMetric creates a new metric instance.
func NewMetric(name, originalName string, source *Source, connector interface{}) *Metric {
	return &Metric{
		Name:         name,
		OriginalName: originalName,
		Source:       source,
		Connector:    connector,
	}
}

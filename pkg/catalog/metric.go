package catalog

// A Metric represents a metric entry.
type Metric struct {
	Name         string
	OriginalName string
	Source       *Source
}

// NewMetric creates a new Metric instances.
func NewMetric(name, originalName string, source *Source) *Metric {
	return &Metric{
		Name:         name,
		OriginalName: originalName,
		Source:       source,
	}
}

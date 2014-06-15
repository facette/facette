package catalog

// Metric represents a metric entry.
type Metric struct {
	Name   string
	Source *Source
}

// NewMetric creates a new metric instance.
func NewMetric(name string, source *Source) *Metric {
	return &Metric{
		Name:   name,
		Source: source,
	}
}

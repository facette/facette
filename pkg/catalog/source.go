package catalog

// Source represents the source of a set of metric entries (e.g. an host name).
type Source struct {
	Name         string
	OriginalName string
	Metrics      map[string]*Metric
	Origin       *Origin
}

// NewSource creates a new source instance.
func NewSource(name, originalName string, origin *Origin) *Source {
	return &Source{
		Name:         name,
		OriginalName: originalName,
		Metrics:      make(map[string]*Metric),
		Origin:       origin,
	}
}

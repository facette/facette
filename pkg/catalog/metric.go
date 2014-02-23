package catalog

// A Metric represents a metric entry.
type Metric struct {
	Name         string
	OriginalName string
	source       *Source
}

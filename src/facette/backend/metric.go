package backend

// A Metric represents a metric entry.
type Metric struct {
	Name     string
	Dataset  string
	FilePath string
	source   *Source
}

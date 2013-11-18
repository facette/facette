package backend

// A Metric represents a RRD metric entry.
type Metric struct {
	Name     string
	Dataset  string
	FilePath string
	source   *Source
}

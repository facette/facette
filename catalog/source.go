package catalog

// Source represents a catalog source instance.
type Source struct {
	Name    string
	Metrics map[string]*Metric
	origin  *Origin
}

// Catalog returns the parent catalog of the origin.
func (s *Source) Catalog() *Catalog {
	return s.origin.catalog
}

// Origin returns the parent origin of the source.
func (s *Source) Origin() *Origin {
	return s.origin
}

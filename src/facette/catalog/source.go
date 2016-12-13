package catalog

// Source represents a catalog source instance.
type Source struct {
	Name         string
	OriginalName string
	metrics      map[string]*Metric
	origin       *Origin
}

// Metric returns a metric from the catalog source.
func (s *Source) Metric(name string) (*Metric, error) {
	s.origin.catalog.RLock()
	defer s.origin.catalog.RUnlock()

	m, ok := s.metrics[name]
	if !ok {
		return nil, ErrUnknownMetric
	}

	return m, nil
}

// Metrics returns a slice of metrics from the catalog source.
func (s *Source) Metrics() []*Metric {
	s.origin.catalog.RLock()
	defer s.origin.catalog.RUnlock()

	items := []*Metric{}
	for _, m := range s.metrics {
		items = append(items, m)
	}

	return items
}

// Origin returns the parent origin from the catalog source.
func (s *Source) Origin() *Origin {
	s.origin.catalog.RLock()
	defer s.origin.catalog.RUnlock()

	return s.origin
}

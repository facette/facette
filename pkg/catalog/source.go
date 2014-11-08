package catalog

import (
	"fmt"
)

// Source represents the source of a set of metric entries (e.g. an host name).
type Source struct {
	Name         string
	OriginalName string
	metrics      map[string]*Metric
	origin       *Origin
}

// NewSource creates a new source instance.
func NewSource(name, origName string, origin *Origin) *Source {
	return &Source{
		Name:         name,
		OriginalName: origName,
		metrics:      make(map[string]*Metric),
		origin:       origin,
	}
}

// MetricExists returns whether a source exists for the origin based on its name.
func (s *Source) MetricExists(name string) bool {
	s.origin.catalog.RLock()
	defer s.origin.catalog.RUnlock()

	_, ok := s.metrics[name]
	return ok
}

// GetMetric returns an existing source metric entry based on its name.
func (s *Source) GetMetric(name string) (*Metric, error) {
	s.origin.catalog.RLock()
	defer s.origin.catalog.RUnlock()

	m, ok := s.metrics[name]
	if !ok {
		return nil, fmt.Errorf("unknown metric `%s' for source `%s' in origin `%s'", name, s.Name, s.origin.Name)
	}

	return m, nil
}

// GetMetrics returns a slice of metrics for the source.
func (s *Source) GetMetrics() []*Metric {
	s.origin.catalog.RLock()
	defer s.origin.catalog.RUnlock()

	items := make([]*Metric, 0)
	for _, m := range s.metrics {
		items = append(items, m)
	}

	return items
}

// GetOrigin returns the parent origin of the source.
func (s *Source) GetOrigin() *Origin {
	s.origin.catalog.RLock()
	defer s.origin.catalog.RUnlock()

	return s.origin
}

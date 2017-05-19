package catalog

import "sort"

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

	metrics := metricList{}
	for _, m := range s.metrics {
		metrics = append(metrics, m)
	}
	sort.Sort(metrics)

	return metrics
}

// Origin returns the parent origin from the catalog source.
func (s *Source) Origin() *Origin {
	s.origin.catalog.RLock()
	defer s.origin.catalog.RUnlock()

	return s.origin
}

type sourceList []*Source

func (l sourceList) Len() int {
	return len(l)
}

func (l sourceList) Less(i, j int) bool {
	return l[i].Name < l[j].Name
}

func (l sourceList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

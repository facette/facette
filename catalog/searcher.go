package catalog

import (
	"sort"
	"sync"

	"facette.io/sliceutil"
)

// Searcher represents a catalog searcher instance.
type Searcher struct {
	sync.RWMutex
	catalogs []*Catalog
}

// NewSearcher creates a new catalog search instance.
func NewSearcher() *Searcher {
	return &Searcher{}
}

// Register registers a new catalog in the catalog searcher.
func (s *Searcher) Register(c *Catalog) {
	s.Lock()
	defer s.Unlock()

	s.catalogs = append(s.catalogs, c)
}

// Unregister unregisters a catalog from the catalog searcher.
func (s *Searcher) Unregister(c *Catalog) {
	s.Lock()
	defer s.Unlock()

	idx := sliceutil.IndexOf(s.catalogs, c)
	if idx == -1 {
		return
	}
	s.catalogs = append(s.catalogs[:idx], s.catalogs[idx+1:]...)
}

// Origins returns a slice of origins from the catalog searcher.
func (s *Searcher) Origins(originName string) []*Origin {
	var result []*Origin

	s.RLock()
	defer s.RUnlock()

	for _, c := range s.catalogs {
		for _, o := range c.Origins {
			if originName != "" && o.Name != originName {
				continue
			}
			result = append(result, o)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Name < result[j].Name {
			return true
		} else if result[i].Name > result[j].Name {
			return false
		}
		return result[i].Catalog().Priority > result[j].Catalog().Priority
	})

	return result
}

// Sources returns a slice of sources from the catalog searcher.
func (s *Searcher) Sources(originName, sourceName string) []*Source {
	var result []*Source

	s.RLock()
	defer s.RUnlock()

	for _, o := range s.Origins(originName) {
		for _, s := range o.Sources {
			if sourceName != "" && s.Name != sourceName {
				continue
			}
			result = append(result, s)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Name < result[j].Name {
			return true
		} else if result[i].Name > result[j].Name {
			return false
		} else if result[i].Origin().Name < result[j].Origin().Name {
			return true
		} else if result[i].Origin().Name > result[j].Origin().Name {
			return false
		}
		return result[i].Catalog().Priority > result[j].Catalog().Priority
	})

	return result
}

// Metrics returns a slice of metrics from the catalog searcher.
func (s *Searcher) Metrics(originName, sourceName, metricName string) []*Metric {
	var result []*Metric

	s.RLock()
	defer s.RUnlock()

	for _, s := range s.Sources(originName, sourceName) {
		for _, m := range s.Metrics {
			if metricName != "" && m.Name != metricName {
				continue
			}
			result = append(result, m)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Name < result[j].Name {
			return true
		} else if result[i].Name > result[j].Name {
			return false
		}
		return result[i].Catalog().Priority > result[j].Catalog().Priority
	})

	return result
}

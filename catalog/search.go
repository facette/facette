package catalog

import (
	"sort"
	"sync"

	"facette.io/sliceutil"
)

// Searcher represents a catalgo searcher instance.
type Searcher struct {
	catalogs catalogList
	sync.RWMutex
}

// NewSearcher creates a new catalog search instance.
func NewSearcher() *Searcher {
	return &Searcher{
		catalogs: catalogList{},
	}
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

// ApplyPriorities reorders the catalogs instances according to the set priorities.
func (s *Searcher) ApplyPriorities() {
	sort.Sort(s.catalogs)
}

// Origins returns a slice of origins from the catalog searcher.
func (s *Searcher) Origins(name string, limit int) []*Origin {
	s.RLock()
	defer s.RUnlock()

	result := []*Origin{}
	for _, c := range s.catalogs {
		for _, o := range c.Origins() {
			if name != "" && o.Name != name {
				continue
			} else if limit > -1 && len(result) >= limit {
				return result
			}
			result = append(result, o)
		}
	}

	return result
}

// Sources returns a slice of sources from the catalog searcher.
func (s *Searcher) Sources(origin, name string, limit int) []*Source {
	s.RLock()
	defer s.RUnlock()

	result := []*Source{}
	for _, o := range s.Origins(origin, -1) {
		for _, s := range o.Sources() {
			if name != "" && s.Name != name {
				continue
			} else if limit > -1 && len(result) >= limit {
				return result
			}
			result = append(result, s)
		}
	}

	return result
}

// Metrics returns a slice of metrics from the catalog searcher.
func (s *Searcher) Metrics(origin, source, name string, limit int) []*Metric {
	s.RLock()
	defer s.RUnlock()

	result := []*Metric{}
	for _, s := range s.Sources(origin, source, -1) {
		for _, m := range s.Metrics() {
			if name != "" && m.Name != name {
				continue
			} else if limit > -1 && len(result) >= limit {
				return result
			}
			result = append(result, m)
		}
	}

	return result
}

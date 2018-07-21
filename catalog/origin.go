package catalog

import "sort"

// Origin represents a catalog origin instance.
type Origin struct {
	Name         string
	OriginalName string
	sources      map[string]*Source
	catalog      *Catalog
}

// Catalog returns the parent catalog from the catalog origin.
func (o *Origin) Catalog() *Catalog {
	o.catalog.RLock()
	defer o.catalog.RUnlock()

	return o.catalog
}

// Source returns a source from the catalog origin.
func (o *Origin) Source(name string) (*Source, error) {
	o.catalog.RLock()
	defer o.catalog.RUnlock()

	s, ok := o.sources[name]
	if !ok {
		return nil, ErrUnknownSource
	}

	return s, nil
}

// Sources returns a slice of sources from the catalog origin.
func (o *Origin) Sources() []*Source {
	o.catalog.RLock()
	defer o.catalog.RUnlock()

	sources := sourceList{}
	for _, s := range o.sources {
		sources = append(sources, s)
	}
	sort.Sort(sources)

	return sources
}

type originList []*Origin

func (l originList) Len() int {
	return len(l)
}

func (l originList) Less(i, j int) bool {
	return l[i].Name < l[j].Name
}

func (l originList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

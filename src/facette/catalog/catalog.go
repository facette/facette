package catalog

import "sync"

// Catalog represents a catalog instance.
type Catalog struct {
	name     string
	origins  map[string]*Origin
	priority int
	sync.RWMutex
}

// NewCatalog creates a new catalog instance.
func NewCatalog(name string) *Catalog {
	return &Catalog{
		name:     name,
		origins:  make(map[string]*Origin),
		priority: 100,
	}
}

// Name returns the name of the catalog.
func (c *Catalog) Name() string {
	return c.name
}

// SetPriority sets the catalog priority value used for metrics conflicts.
func (c *Catalog) SetPriority(priority int) {
	c.priority = priority
}

// Insert registers a new record in the catalog.
func (c *Catalog) Insert(r *Record) {
	var (
		origin *Origin
		source *Source
		ok     bool
	)

	c.Lock()
	defer c.Unlock()

	origin, ok = c.origins[r.Origin]
	if !ok {
		c.origins[r.Origin] = &Origin{
			Name:         r.Origin,
			OriginalName: r.OriginalOrigin,
			sources:      make(map[string]*Source),
			catalog:      c,
		}
		origin = c.origins[r.Origin]
	}

	source, ok = origin.sources[r.Source]
	if !ok {
		origin.sources[r.Source] = &Source{
			Name:         r.Source,
			OriginalName: r.OriginalSource,
			metrics:      make(map[string]*Metric),
			origin:       origin,
		}
		source = origin.sources[r.Source]
	}

	_, ok = source.metrics[r.Metric]
	if !ok {
		source.metrics[r.Metric] = &Metric{
			Name:         r.Metric,
			OriginalName: r.OriginalMetric,
			source:       source,
			connector:    r.Connector,
		}
	}
}

// Origin returns an origin from the catalog.
func (c *Catalog) Origin(name string) (*Origin, error) {
	c.RLock()
	defer c.RUnlock()

	if _, ok := c.origins[name]; !ok {
		return nil, ErrUnknownOrigin
	}

	return c.origins[name], nil
}

// Origins returns a slice of origins from the catalog.
func (c *Catalog) Origins() []*Origin {
	c.RLock()
	defer c.RUnlock()

	items := []*Origin{}
	for _, o := range c.origins {
		items = append(items, o)
	}

	return items
}

// Source returns a source for a specific origin from the catalog.
func (c *Catalog) Source(origin, name string) (*Source, error) {
	c.RLock()
	defer c.RUnlock()

	o, err := c.Origin(origin)
	if err != nil {
		return nil, err
	}

	return o.Source(name)
}

// Metric returns a metric for specifics origin and source from the catalog.
func (c *Catalog) Metric(origin, source, name string) (*Metric, error) {
	c.RLock()
	defer c.RUnlock()

	s, err := c.Source(origin, source)
	if err != nil {
		return nil, err
	}

	return s.Metric(name)
}

// catalogList represents a list of catalog instances.
type catalogList []*Catalog

func (l catalogList) Len() int {
	return len(l)
}

func (l catalogList) Less(i, j int) bool {
	return l[i].priority < l[j].priority
}

func (l catalogList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

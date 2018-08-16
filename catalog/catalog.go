package catalog

// Catalog represents a catalog instance.
type Catalog struct {
	Name      string
	Priority  int
	Origins   map[string]*Origin
	Connector interface{}
}

// New creates a new catalog instance.
func New(name string, connector interface{}) *Catalog {
	return &Catalog{
		Name:      name,
		Origins:   make(map[string]*Origin),
		Connector: connector,
	}
}

// Insert inserts a new record into the catalog.
func (c *Catalog) Insert(r *Record) {
	origin, ok := c.Origins[r.Origin]
	if !ok {
		c.Origins[r.Origin] = &Origin{
			Name:    r.Origin,
			Sources: make(map[string]*Source),
			catalog: c,
		}
		origin = c.Origins[r.Origin]
	}

	source, ok := origin.Sources[r.Source]
	if !ok {
		origin.Sources[r.Source] = &Source{
			Name:    r.Source,
			Metrics: make(map[string]*Metric),
			origin:  origin,
		}
		source = origin.Sources[r.Source]
	}

	_, ok = source.Metrics[r.Metric]
	if !ok {
		source.Metrics[r.Metric] = &Metric{
			Name:       r.Metric,
			Attributes: r.Attributes,
			source:     source,
		}
	}
}

// Origin returns an existing origin from the catalog or an error if the origin doesn't exist.
func (c *Catalog) Origin(originName string) (*Origin, error) {
	origin, ok := c.Origins[originName]
	if !ok {
		return nil, ErrUnknownOrigin
	}

	return origin, nil
}

// Metric returns an existing metric from the catalog or an error if the metric doesn't exist.
func (c *Catalog) Metric(originName, sourceName, metricName string) (*Metric, error) {
	source, err := c.Source(originName, sourceName)
	if err != nil {
		return nil, err
	}

	metric, ok := source.Metrics[metricName]
	if !ok {
		return nil, ErrUnknownMetric
	}

	return metric, nil
}

// Source returns an existing source from the catalog or an error if the source doesn't exist.
func (c *Catalog) Source(originName, sourceName string) (*Source, error) {
	origin, err := c.Origin(originName)
	if err != nil {
		return nil, err
	}

	source, ok := origin.Sources[sourceName]
	if !ok {
		return nil, ErrUnknownSource
	}

	return source, nil
}

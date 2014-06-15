package catalog

// Origin represents an origin of source sets (e.g. a Collectd or Graphite instance).
type Origin struct {
	Name    string
	Sources map[string]*Source
	Catalog *Catalog
}

// NewOrigin creates a new origin instance.
func NewOrigin(name string, catalog *Catalog) *Origin {
	return &Origin{
		Name:    name,
		Sources: make(map[string]*Source),
		Catalog: catalog,
	}
}

package catalog

// Origin represents an origin of source sets (e.g. a Collectd or Graphite instance).
type Origin struct {
	Name         string
	OriginalName string
	Sources      map[string]*Source
	Catalog      *Catalog
}

// NewOrigin creates a new origin instance.
func NewOrigin(name, originalName string, catalog *Catalog) *Origin {
	return &Origin{
		Name:         name,
		OriginalName: originalName,
		Sources:      make(map[string]*Source),
		Catalog:      catalog,
	}
}

package catalog

// Origin represents a catalog origin instance.
type Origin struct {
	Name    string
	Sources map[string]*Source
	catalog *Catalog
}

// Catalog returns the parent catalog of the origin.
func (o *Origin) Catalog() *Catalog {
	return o.catalog
}

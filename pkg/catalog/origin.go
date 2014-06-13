package catalog

import (
	"time"

	"github.com/facette/facette/pkg/config"
)

// Origin represents an origin of source sets (e.g. a Collectd or Graphite instance).
type Origin struct {
	Name        string
	Config      *config.OriginConfig
	Sources     map[string]*Source
	Filters     filterChain
	Catalog     *Catalog
	LastRefresh time.Time
}

// NewOrigin creates a new origin instance.
func NewOrigin(name string, config *config.OriginConfig, catalog *Catalog) *Origin {
	return &Origin{
		Name:    name,
		Config:  config,
		Sources: make(map[string]*Source),
		Filters: newFilterChain(config.Filters, catalog.RecordChan),
		Catalog: catalog,
	}
}

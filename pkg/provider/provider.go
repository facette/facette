package provider

import (
	"time"

	"github.com/facette/facette/pkg/catalog"
	"github.com/facette/facette/pkg/config"
	"github.com/facette/facette/pkg/connector"
)

// Provider represents a provider instance.
type Provider struct {
	Name        string
	Config      *config.ProviderConfig
	Catalog     *catalog.Catalog
	Connector   connector.Connector
	Filters     filterChain
	LastRefresh time.Time
}

// NewProvider creates a new provider instance.
func NewProvider(name string, config *config.ProviderConfig, catalog *catalog.Catalog) *Provider {
	return &Provider{
		Name:    name,
		Config:  config,
		Catalog: catalog,
		Filters: newFilterChain(config.Filters, catalog.RecordChan),
	}
}

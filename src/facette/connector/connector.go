package connector

import (
	"sort"

	"facette/catalog"
	"facette/mapper"
	"facette/plot"
)

const connectorDefaultTimeout int64 = 10

var (
	version string

	connectors = make(map[string]func(string, mapper.Map) (Connector, error))
)

// Connector represents a connector handler interface.
type Connector interface {
	Name() string
	Refresh(chan<- *catalog.Record) chan error
	Plots(*plot.Query) ([]plot.Series, error)
}

// NewConnector creates a new instance of a connector handler.
func NewConnector(typ string, name string, settings mapper.Map) (Connector, error) {
	// Check for existing connector handler
	if _, ok := connectors[typ]; !ok {
		return nil, ErrUnsupportedConnector
	}

	// Return new connector handler instance
	return connectors[typ](name, settings)
}

// Connectors returns the list of supported connectors.
func Connectors() []string {
	list := []string{}
	for name := range connectors {
		list = append(list, name)
	}

	sort.Strings(list)

	return list
}

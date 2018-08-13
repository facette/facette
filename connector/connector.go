package connector

import (
	"sort"

	"facette.io/facette/catalog"
	"facette.io/facette/series"
	"facette.io/logger"
	"facette.io/maputil"
)

const defaultTimeout = 10

var connectors = make(map[string]func(string, *maputil.Map, *logger.Logger) (Connector, error))

// Connector represents a connector handler interface.
type Connector interface {
	Name() string
	Points(*series.Query) ([]series.Series, error)
	Refresh(chan<- *catalog.Record) error
}

// New creates a new instance of a connector handler.
func New(typ, name string, settings *maputil.Map, logger *logger.Logger) (Connector, error) {
	// Check for existing connector handler
	if _, ok := connectors[typ]; !ok {
		return nil, ErrUnsupportedConnector
	}

	// Return new connector handler instance
	return connectors[typ](name, settings, logger)
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

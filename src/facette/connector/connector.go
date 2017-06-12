package connector

import (
	"sort"

	"facette/catalog"
	"facette/series"

	"github.com/facette/logger"
	"github.com/facette/maputil"
)

const connectorDefaultTimeout int = 10

var (
	version string

	connectors = make(map[string]func(string, *maputil.Map, *logger.Logger) (Connector, error))
)

// Connector represents a connector handler interface.
type Connector interface {
	Name() string
	Refresh(chan<- *catalog.Record) error
	Points(*series.Query) ([]series.Series, error)
}

// NewConnector creates a new instance of a connector handler.
func NewConnector(typ string, name string, settings *maputil.Map, log *logger.Logger) (Connector, error) {
	// Check for existing connector handler
	if _, ok := connectors[typ]; !ok {
		return nil, ErrUnsupportedConnector
	}

	// Return new connector handler instance
	return connectors[typ](name, settings, log)
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

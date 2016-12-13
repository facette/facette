package backend

import (
	"sort"

	"github.com/brettlangdon/forge"
)

var drivers = make(map[string]func() sqlDriver)

// sqlDriver represents the backend database driver interface.
type sqlDriver interface {
	name() string
	buildDSN(*forge.Section) (string, error)
	whereClause(string, interface{}) (string, interface{})
}

// Drivers returns the list of supported backend drivers.
func Drivers() []string {
	list := []string{}
	for name := range drivers {
		list = append(list, name)
	}

	sort.Strings(list)

	return list
}

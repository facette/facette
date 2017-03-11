package backend

import (
	"sort"

	"facette/mapper"
)

var drivers = make(map[string]func(*mapper.Map) (sqlDriver, error))

// sqlDriver represents the backend database driver interface.
type sqlDriver interface {
	name() string
	DSN() string
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

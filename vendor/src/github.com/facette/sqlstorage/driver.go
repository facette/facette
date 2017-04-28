package sqlstorage

import (
	"database/sql"
	"sort"

	"github.com/facette/maputil"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

type sqlDriver interface {
	Name() string
	Open() (*sql.DB, error)
	Init(*gorm.DB) error
	NormalizeError(error) error
	WhereClause(string, interface{}) (string, interface{})
}

func newSQLDriver(name string, settings *maputil.Map) (sqlDriver, error) {
	driver, _ := settings.GetString("driver", "")
	if driver == "" {
		return nil, errors.Wrap(ErrUnsupportedDriver, "empty driver setting")
	}

	if v, ok := drivers[driver]; ok {
		return v(name, settings)
	}

	return nil, ErrUnsupportedDriver
}

var drivers = map[string]func(string, *maputil.Map) (sqlDriver, error){}

// Drivers returns the list of supported storage drivers.
func Drivers() []string {
	list := []string{}
	for name := range drivers {
		list = append(list, name)
	}

	sort.Strings(list)

	return list
}

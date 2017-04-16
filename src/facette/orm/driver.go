package orm

import (
	"fmt"
	"reflect"
	"sort"
	"time"

	"facette/mapper"

	"github.com/pkg/errors"
)

// SQLDriver represents the database driver interface.
type SQLDriver interface {
	DSN() string
	QuoteName(string) string
	BooleanValue(string) string
	NowCall() string
	WhereClause(string, interface{}) (string, interface{})
	LimitClause(int, int) string

	setDB(db *DB)
	init() error
	name() string
	hasTable(string) bool
	hasColumn(string, string) bool
	hasIndex(string, string) bool
	bindVar(int) string
	typeOf(reflect.Value, bool) (string, error)
	returningSuffix(string) string
	adaptValue(reflect.Value) reflect.Value
	scanValue(reflect.Value, reflect.Value) error
	normalizeError(error) error
}

func newDriver(settings *mapper.Map) (SQLDriver, error) {
	driver, _ := settings.GetString("driver", "")
	if driver == "" {
		return nil, errors.Wrap(ErrUnsupportedDriver, "empty driver setting")
	}

	if v, ok := drivers[driver]; ok {
		return v(settings)
	}

	return nil, ErrUnsupportedDriver
}

// commonDriver implements the common database driver interface methods.
type commonDriver struct {
	db *DB
}

func (d commonDriver) QuoteName(name string) string {
	return fmt.Sprintf("%q", name)
}

func (d commonDriver) BooleanValue(value string) string {
	return value
}

func (d commonDriver) NowCall() string {
	return "now()"
}

func (d commonDriver) WhereClause(column string, v interface{}) (string, interface{}) {
	return column + " = ?", v
}

func (d commonDriver) LimitClause(offset, limit int) string {
	return fmt.Sprintf("LIMIT %d, %d", offset, limit)
}

func (d *commonDriver) setDB(db *DB) {
	d.db = db
}

func (d commonDriver) init() error {
	return nil
}

func (d commonDriver) bindVar(count int) string {
	return "?"
}

func (d commonDriver) returningSuffix(columnName string) string {
	return ""
}

func (d commonDriver) adaptValue(value reflect.Value) reflect.Value {
	switch value.Kind() {
	case reflect.Struct:
		if v, ok := value.Interface().(time.Time); ok {
			return reflect.ValueOf(v.UTC().Format(TimeFormat))
		}
	}

	return value
}

func (d commonDriver) scanValue(dst, src reflect.Value) error {
	return nil
}

var drivers = map[string]func(settings *mapper.Map) (SQLDriver, error){}

// Drivers returns the list of supported backend drivers.
func Drivers() []string {
	list := []string{}
	for name := range drivers {
		list = append(list, name)
	}

	sort.Strings(list)

	return list
}

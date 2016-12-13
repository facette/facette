package orm

import (
	"fmt"
	"reflect"
	"time"
)

var drivers = map[string]SQLDriver{}

// SQLDriver represents the database driver interface.
type SQLDriver interface {
	QuoteName(string) string
	BooleanValue(string) string
	NowCall() string
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

func newDriver(name string) SQLDriver {
	if value, ok := drivers[name]; ok {
		return reflect.New(reflect.TypeOf(value).Elem()).Interface().(SQLDriver)
	}
	return nil
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

// registerDriver registers a new driver intance.
func registerDriver(name string, driver SQLDriver) {
	drivers[name] = driver
}

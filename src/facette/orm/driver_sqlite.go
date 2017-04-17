// +build !disable_backend_sqlite

package orm

import (
	"database/sql"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"facette/mapper"

	"github.com/facette/sliceutil"
	sqlite3 "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

const (
	defaultSqliteDriverPath = "data.db"
)

// sqliteDriver implements the database driver interface for SQLite 3.
type sqliteDriver struct {
	commonDriver

	path string
}

func (d sqliteDriver) DSN() string {
	return d.path
}

func (d sqliteDriver) BooleanValue(value string) string {
	if value == "true" {
		return "1"
	}

	return "0"
}

func (d sqliteDriver) NowCall() string {
	return "CURRENT_TIMESTAMP"
}

func (d sqliteDriver) WhereClause(column string, v interface{}) (string, interface{}) {
	operator := "="
	extra := ""

	switch v.(type) {
	case GlobModifier:
		operator = "LIKE"
		extra = " ESCAPE '\\'"

		s := string(v.(GlobModifier))
		s = strings.Replace(s, "%", "\\%", -1)
		s = strings.Replace(s, "_", "\\_", -1)
		s = strings.Replace(s, "*", "%", -1)

		v = strings.Replace(s, "?", "_", -1)

	case RegexpModifier:
		operator = "REGEXP"
		v = "(?i)" + string(v.(RegexpModifier))

	case string:
		if s := v.(string); s == "null" {
			operator = "IS"
			v = nil
		}
	}

	return fmt.Sprintf("%s %s ?%s", column, operator, extra), v
}

func (d sqliteDriver) init() error {
	// Enable "foreign_key" pragma
	_, err := d.db.Raw("PRAGMA foreign_keys=ON").quiet().Result()
	return err
}

func (d sqliteDriver) name() string {
	return "sqlite3_ext"
}

func (d sqliteDriver) hasTable(tableName string) bool {
	var count int

	d.db.From("sqlite_master").
		Where("type = ?", "table").
		Where("name = ?", tableName).
		quiet().
		Count(&count)

	return count > 0
}

func (d sqliteDriver) hasColumn(tableName, columnName string) bool {
	rows, err := d.db.Raw(fmt.Sprintf("PRAGMA table_info(%s)", d.QuoteName(tableName))).quiet().Rows()
	if err != nil {
		return false
	}
	defer rows.Close()

	for rows.Next() {
		values := make([]interface{}, 6)
		for i := range values {
			values[i] = new(*string)
		}

		if err := rows.Scan(values...); err != nil {
			continue
		}

		if **values[1].(**string) == columnName {
			return true
		}
	}

	return false
}

func (d sqliteDriver) hasIndex(tableName, indexName string) bool {
	var count int

	d.db.From("sqlite_master").
		Where("type = ?", "index").
		Where("name = ?", indexName).
		Where("tbl_name = ?", tableName).
		quiet().
		Count(&count)

	return count > 0
}

func (d sqliteDriver) typeOf(rv reflect.Value, autoIncrement bool) (string, error) {
	switch rv.Kind() {
	case reflect.Bool, reflect.Int8, reflect.Int16, reflect.Uint8, reflect.Int, reflect.Int32, reflect.Uint,
		reflect.Uint16, reflect.Uint32, reflect.Int64, reflect.Uint64:
		return "integer", nil

	case reflect.Float32, reflect.Float64:
		return "real", nil

	case reflect.String:
		return "text", nil

	case reflect.Struct:
		if _, ok := rv.Interface().(time.Time); ok {
			return "text", nil
		}

	default:
		if _, ok := rv.Interface().([]byte); ok {
			return "blob", nil
		}
	}

	return "", ErrUnsupportedType
}

func (d sqliteDriver) scanValue(dst, src reflect.Value) error {
	if dst.Kind() == reflect.Ptr {
		dst.Set(reflect.New(dst.Type().Elem()))
		dst = dst.Elem()
	}

	switch dst.Kind() {
	case reflect.Bool:
		dst.SetBool(src.Interface().(int64) != 0)

	case reflect.Struct:
		if _, ok := dst.Interface().(time.Time); ok {
			t, err := time.Parse(TimeFormat, string(src.Interface().([]byte)))
			if err != nil {
				return err
			}

			dst.Set(reflect.ValueOf(t))
		}

	default:
		return ErrNotScanable
	}

	return nil
}

func (d sqliteDriver) normalizeError(err error) error {
	if _, ok := err.(sqlite3.Error); !ok {
		return err
	}

	switch err.(sqlite3.Error).ExtendedCode {
	case sqlite3.ErrConstraintPrimaryKey, sqlite3.ErrConstraintUnique:
		return ErrConstraintUnique

	case sqlite3.ErrConstraintForeignKey:
		return ErrConstraintForeignKey

	case sqlite3.ErrConstraintNotNull:
		return ErrConstraintNotNull
	}

	return err
}

func init() {
	if !sliceutil.Has(sql.Drivers(), "sqlite3_ext") {
		sql.Register("sqlite3_ext", &sqlite3.SQLiteDriver{
			ConnectHook: func(c *sqlite3.SQLiteConn) error {
				return c.RegisterFunc("regexp", regexp.MatchString, true)
			},
		})
	}

	drivers["sqlite"] = func(settings *mapper.Map) (SQLDriver, error) {
		var err error

		d := &sqliteDriver{}

		if d.path, err = settings.GetString("path", defaultSqliteDriverPath); err != nil {
			return nil, errors.Wrap(err, "invalid \"path\" setting")
		}

		return d, nil
	}
}

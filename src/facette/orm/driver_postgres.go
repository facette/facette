// +build !disable_postgres

package orm

import (
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/lib/pq"
)

// postgresDriver implements the database driver interface for PostgreSQL.
type postgresDriver struct {
	commonDriver
	dbName string
}

func (d postgresDriver) LimitClause(offset, limit int) string {
	return fmt.Sprintf("OFFSET %d LIMIT %d", offset, limit)
}

func (d *postgresDriver) init() error {
	// Get current database name
	return d.db.Select("current_database()").quiet().Row().Scan(&d.dbName)
}

func (d postgresDriver) name() string {
	return "postgres"
}

func (d postgresDriver) hasTable(tableName string) bool {
	var count int

	d.db.From("information_schema.tables").
		Where("table_catalog = $1", d.dbName).
		Where("table_name = $2", tableName).
		Where("table_type = $3", "BASE TABLE").
		quiet().
		Count(&count)

	return count > 0
}

func (d postgresDriver) hasColumn(tableName, columnName string) bool {
	var count int

	d.db.From("information_schema.columns").
		Where("table_catalog = $1", d.dbName).
		Where("table_name = $2", tableName).
		Where("column_name = $3", columnName).
		quiet().
		Count(&count)

	return count > 0
}

func (d postgresDriver) hasIndex(tableName, indexName string) bool {
	var count int

	d.db.From("pg_indexes").
		Where("tablename = $1", tableName).
		Where("indexname = $2", indexName).
		quiet().
		Count(&count)

	return count > 0
}

func (d postgresDriver) bindVar(count int) string {
	return fmt.Sprintf("$%d", count)
}

func (d postgresDriver) typeOf(rv reflect.Value, autoIncrement bool) (string, error) {
	switch rv.Kind() {
	case reflect.Bool:
		return "boolean", nil

	case reflect.Float32, reflect.Float64:
		return "numeric", nil

	case reflect.Int8, reflect.Int16, reflect.Uint8:
		if autoIncrement {
			return "serial", nil
		}
		return "smallint", nil

	case reflect.Int, reflect.Int32, reflect.Uint, reflect.Uint16, reflect.Uint32:
		if autoIncrement {
			return "serial", nil
		}
		return "integer", nil

	case reflect.Int64, reflect.Uint64:
		if autoIncrement {
			return "bigserial", nil
		}
		return "bigint", nil

	case reflect.String:
		return "text", nil

	case reflect.Struct:
		if _, ok := rv.Interface().(time.Time); ok {
			return "timestamp without time zone", nil
		}

	default:
		if _, ok := rv.Interface().([]byte); ok {
			return "bytea", nil
		}
	}

	return "", ErrUnsupportedType
}

func (d postgresDriver) returningSuffix(columnName string) string {
	return " RETURNING " + d.QuoteName(columnName)
}

func (d postgresDriver) scanValue(dst, src reflect.Value) error {
	if dst.Kind() == reflect.Ptr {
		dst.Set(reflect.New(dst.Type().Elem()))
		dst = dst.Elem()
	}

	switch dst.Kind() {
	case reflect.Float32, reflect.Float64:
		bitSize := 64
		if dst.Kind() == reflect.Float32 {
			bitSize = 32
		}

		v, err := strconv.ParseFloat(string(src.Interface().([]byte)), bitSize)
		if err != nil {
			return err
		}

		dst.Set(reflect.ValueOf(v).Convert(dst.Type()))

	case reflect.Struct:
		if _, ok := dst.Interface().(time.Time); ok {
			t, ok := src.Interface().(time.Time)
			if !ok {
				return ErrNotScanable
			}

			dst.Set(reflect.ValueOf(t.UTC()))
		}

	default:
		return ErrNotScanable
	}

	return nil
}

func (d postgresDriver) normalizeError(err error) error {
	if _, ok := err.(*pq.Error); !ok {
		return err
	}

	switch err.(*pq.Error).Code.Name() {
	case "unique_violation":
		return ErrConstraintUnique

	case "foreign_key_violation":
		return ErrConstraintForeignKey

	case "not_null_violation":
		return ErrConstraintNotNull
	}

	return err
}

func init() {
	registerDriver("postgres", &postgresDriver{})
}

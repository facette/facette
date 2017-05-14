// +build !disable_driver_pgsql

package sqlstorage

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/facette/maputil"
	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

const (
	defaultPgsqlDriverHost    = "localhost"
	defaultPgsqlDriverPort    = 5432
	defaultPgsqlDriverSSLMode = "disable"
)

// pgsqlDriver implements the database driver interface for PostgreSQL.
type pgsqlDriver struct {
	host     string
	port     int
	user     string
	password string
	dbName   string
	sslMode  string
}

func (d pgsqlDriver) Name() string {
	return "postgres"
}

func (d pgsqlDriver) Open() (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"%s=%v %s=%v %s=%v %s=%v %s=%v %s=%v",
		"host", d.host,
		"port", d.port,
		"user", d.user,
		"password", d.password,
		"dbname", d.dbName,
		"sslmode", d.sslMode,
	)

	return sql.Open("postgres", dsn)
}

func (d pgsqlDriver) Init(db *gorm.DB) error {
	return nil
}

func (d pgsqlDriver) NormalizeError(err error) error {
	pgsqlErr, ok := err.(*pq.Error)
	if !ok {
		return err
	}

	switch pgsqlErr.Code.Name() {
	case "unique_violation":
		return ErrItemConflict

	case "foreign_key_violation":
		return ErrUnknownReference

	case "not_null_violation":
		return ErrMissingField
	}

	return err
}

func (d pgsqlDriver) WhereClause(column string, v interface{}) (string, interface{}) {
	operator := "="
	switch v.(type) {
	case GlobModifier:
		operator = "ILIKE"

		s := string(v.(GlobModifier))
		s = strings.Replace(s, "%", "\\%", -1)
		s = strings.Replace(s, "_", "\\_", -1)
		s = strings.Replace(s, "*", "%", -1)

		v = strings.Replace(s, "?", "_", -1)

	case RegexpModifier:
		operator = "~"
		v = "(?i)" + string(v.(RegexpModifier))

	case string:
		if s := v.(string); s == "null" {
			operator = "IS"
			v = nil
		}
	}

	return fmt.Sprintf("%s %s ?", column, operator), v
}

func init() {
	drivers["pgsql"] = func(name string, settings *maputil.Map) (sqlDriver, error) {
		var err error

		d := &pgsqlDriver{}

		if d.host, err = settings.GetString("host", defaultPgsqlDriverHost); err != nil {
			return nil, errors.Wrap(err, "invalid \"host\" setting")
		}

		if d.port, err = settings.GetInt("port", defaultPgsqlDriverPort); err != nil {
			return nil, errors.Wrap(err, "invalid \"port\" setting")
		}

		if d.user, err = settings.GetString("user", name); err != nil {
			return nil, errors.Wrap(err, "invalid \"user\" setting")
		}

		if d.password, err = settings.GetString("password", ""); err != nil || d.password == "" {
			return nil, errors.Wrap(err, "invalid \"password\" setting")
		}

		if d.dbName, err = settings.GetString("dbname", name); err != nil {
			return nil, errors.Wrap(err, "invalid \"dbname\" setting")
		}

		if d.sslMode, err = settings.GetString("sslmode", defaultPgsqlDriverSSLMode); err != nil {
			return nil, errors.Wrap(err, "invalid \"sslmode\" setting")
		}

		return d, nil
	}
}

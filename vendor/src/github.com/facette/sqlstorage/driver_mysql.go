// +build !disable_driver_mysql

package sqlstorage

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/facette/maputil"
	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

const (
	defaultMysqlDriverHost = "localhost"
	defaultMysqlDriverPort = 3306
)

// mysqlDriver implements the database driver interface for MySQL.
type mysqlDriver struct {
	host     string
	port     int
	user     string
	password string
	dbName   string
	charset  string
}

func (d mysqlDriver) Name() string {
	return "mysql"
}

func (d mysqlDriver) Open() (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=true",
		d.user,
		d.password,
		d.host,
		d.port,
		d.dbName,
		d.charset,
	)

	return sql.Open("mysql", dsn)
}

func (d mysqlDriver) Init(db *gorm.DB) error {
	return nil
}

func (d mysqlDriver) NormalizeError(err error) error {
	mysqlErr, ok := err.(*mysql.MySQLError)
	if !ok {
		return err
	}

	switch mysqlErr.Number {
	case 1062:
		return ErrItemConflict

	case 1216, 1217, 1451, 1452:
		return ErrUnknownReference

	case 1048, 1364:
		return ErrMissingField
	}

	return err
}

func (d mysqlDriver) WhereClause(column string, v interface{}) (string, interface{}) {
	operator := "="

	switch v.(type) {
	case GlobModifier:
		operator = "LIKE"

		s := string(v.(GlobModifier))
		s = strings.Replace(s, "%", "\\%", -1)
		s = strings.Replace(s, "_", "\\_", -1)
		s = strings.Replace(s, "*", "%", -1)

		v = strings.Replace(s, "?", "_", -1)

	case RegexpModifier:
		operator = "REGEXP"

	case string:
		if s := v.(string); s == "null" {
			operator = "IS"
			v = nil
		}
	}

	return fmt.Sprintf("%s %s ?", column, operator), v
}

func init() {
	drivers["mysql"] = func(name string, settings *maputil.Map) (sqlDriver, error) {
		var err error

		d := &mysqlDriver{}

		if d.host, err = settings.GetString("host", defaultMysqlDriverHost); err != nil {
			return nil, errors.Wrap(err, "invalid \"host\" setting")
		}

		if d.port, err = settings.GetInt("port", defaultMysqlDriverPort); err != nil {
			return nil, errors.Wrap(err, "invalid \"port\" setting")
		}

		if d.user, err = settings.GetString("user", name); err != nil {
			return nil, errors.Wrap(err, "invalid \"user\" setting")
		}

		if d.password, err = settings.GetString("password", ""); err != nil {
			return nil, errors.Wrap(err, "invalid \"password\" setting")
		}

		if d.dbName, err = settings.GetString("dbname", name); err != nil {
			return nil, errors.Wrap(err, "invalid \"dbname\" setting")
		}

		if d.charset, err = settings.GetString("charset", "utf8"); err != nil {
			return nil, errors.Wrap(err, "invalid \"charset\" setting")
		}

		return d, nil
	}
}

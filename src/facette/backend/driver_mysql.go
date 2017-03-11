// +build !disable_mysql

package backend

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"facette/mapper"
)

const (
	defaultMysqlDriverHostName = "localhost"
	defaultMysqlDriverHostPort = 3306
	defaultMysqlDriverUser     = "facette"
	defaultMysqlDriverDatabase = "facette"
)

// mysqlDriver implements the backend database driver interface for MySQL.
type mysqlDriver struct {
	hostName string
	hostPort int
	user     string
	password string
	database string
}

func (d mysqlDriver) name() string {
	return "mysql"
}

func (d mysqlDriver) DSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?interpolateParams=true",
		d.user,
		d.password,
		d.hostName,
		d.hostPort,
		d.database,
	)
}

func (d mysqlDriver) whereClause(column string, v interface{}) (string, interface{}) {
	operator := "="

	switch v.(type) {
	case string:
		s := v.(string)
		if strings.HasPrefix(s, FilterGlobPrefix) {
			operator = "LIKE"
			s = strings.TrimPrefix(s, FilterGlobPrefix)
			s = strings.Replace(s, "%", "\\%", -1)
			s = strings.Replace(s, "_", "\\_", -1)
			s = strings.Replace(s, "*", "%", -1)
			v = strings.Replace(s, "?", "_", -1)
		} else if strings.HasPrefix(s, FilterRegexpPrefix) {
			operator = "REGEXP"
			v = strings.TrimPrefix(s, FilterRegexpPrefix)
		} else if s == "null" {
			operator = "IS"
			v = nil
		}
	}

	return fmt.Sprintf("%s %s ?", column, operator), v
}

func init() {
	drivers["mysql"] = func(settings *mapper.Map) (sqlDriver, error) {
		var (
			d   = mysqlDriver{}
			err error
		)

		if d.hostName, err = settings.GetString("host", defaultMysqlDriverHostName); err != nil {
			return nil, errors.Wrap(err, "mysql setting `host'")
		}

		if d.hostPort, err = settings.GetInt("port", defaultMysqlDriverHostPort); err != nil {
			return nil, errors.Wrap(err, "mysql setting `port'")
		}

		if d.user, err = settings.GetString("user", defaultMysqlDriverUser); err != nil {
			return nil, errors.Wrap(err, "mysql setting `user'")
		}

		if d.password, err = settings.GetString("password", ""); err != nil || d.password == "" {
			return nil, errors.Wrap(err, "mysql setting `password'")
		}

		if d.database, err = settings.GetString("dbname", defaultMysqlDriverDatabase); err != nil {
			return nil, errors.Wrap(err, "mysql setting `dbname'")
		}

		return d, nil
	}
}

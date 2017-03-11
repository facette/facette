// +build !disable_pgsql

package backend

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"facette/mapper"
)

const (
	defaultPgsqlDriverHostName = "localhost"
	defaultPgsqlDriverHostPort = 5432
	defaultPgsqlDriverUser     = "facette"
	defaultPgsqlDriverDatabase = "facette"
)

// pgsqlDriver implements the backend database driver interface for PostgreSQL.
type pgsqlDriver struct {
	hostName string
	hostPort int
	user     string
	password string
	database string
}

func (d pgsqlDriver) name() string {
	return "pgsql"
}

func (d pgsqlDriver) DSN() string {
	return fmt.Sprintf(
		"%s=%v %s=%v %s=%v %s=%v %s=%v",
		"host", d.hostName,
		"port", d.hostPort,
		"user", d.user,
		"password", d.password,
		"dbname", d.database,
	)
}

func (d pgsqlDriver) whereClause(column string, v interface{}) (string, interface{}) {
	operator := "="

	switch v.(type) {
	case string:
		s := v.(string)
		if strings.HasPrefix(s, FilterGlobPrefix) {
			operator = "ILIKE"
			s = strings.TrimPrefix(s, FilterGlobPrefix)
			s = strings.Replace(s, "%", "\\%", -1)
			s = strings.Replace(s, "_", "\\_", -1)
			s = strings.Replace(s, "*", "%", -1)
			v = strings.Replace(s, "?", "_", -1)
		} else if strings.HasPrefix(s, FilterRegexpPrefix) {
			operator = "~"
			v = "(?i)" + strings.TrimPrefix(s, FilterRegexpPrefix)
		} else if s == "null" {
			operator = "IS"
			v = nil
		}
	}

	return fmt.Sprintf("%s %s ?", column, operator), v
}

func init() {
	drivers["pgsql"] = func(settings *mapper.Map) (sqlDriver, error) {
		var (
			d   = pgsqlDriver{}
			err error
		)

		if d.hostName, err = settings.GetString("host", defaultPgsqlDriverHostName); err != nil {
			return nil, errors.Wrap(err, "pgsql setting `host'")
		}

		if d.hostPort, err = settings.GetInt("port", defaultPgsqlDriverHostPort); err != nil {
			return nil, errors.Wrap(err, "pgsql setting `port'")
		}

		if d.user, err = settings.GetString("user", defaultPgsqlDriverUser); err != nil {
			return nil, errors.Wrap(err, "pgsql setting `user'")
		}

		if d.password, err = settings.GetString("password", ""); err != nil || d.password == "" {
			return nil, errors.Wrap(err, "pgsql setting `password'")
		}

		if d.database, err = settings.GetString("dbname", defaultPgsqlDriverDatabase); err != nil {
			return nil, errors.Wrap(err, "pgsql setting `dbname'")
		}

		return d, nil
	}
}

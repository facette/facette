// +build !disable_mysql

package backend

import (
	"fmt"
	"strings"

	"github.com/brettlangdon/forge"
)

// mysqlDriver implements the backend database driver interface for MySQL.
type mysqlDriver struct{}

func (d mysqlDriver) name() string {
	return "mysql"
}

func (d mysqlDriver) buildDSN(config *forge.Section) (string, error) {
	database, err := config.GetString("dbname")
	if err != nil {
		return "", err
	}

	host, err := config.GetString("host")
	if err != nil {
		return "", err
	}

	port, err := config.GetInteger("port")
	if err != nil {
		return "", err
	}

	user, err := config.GetString("user")
	if err != nil {
		return "", err
	}

	password, err := config.GetString("password")
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?interpolateParams=true", user, password, host, port, database), nil
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
	drivers["mysql"] = func() sqlDriver {
		return mysqlDriver{}
	}
}

// +build !disable_sqlite

package backend

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"facette/mapper"
)

const (
	defaultSqliteDriverPath = "data.db"
)

// sqliteDriver implements the backend database driver interface for SQLite 3.
type sqliteDriver struct {
	path string
}

func (d sqliteDriver) name() string {
	return "sqlite3"
}

func (d sqliteDriver) DSN() string {
	return d.path
}

func (d sqliteDriver) whereClause(column string, v interface{}) (string, interface{}) {
	operator := "="
	extra := ""

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
			extra = " ESCAPE '\\'"
		} else if strings.HasPrefix(s, FilterRegexpPrefix) {
			operator = "REGEXP"
			v = "(?i)" + strings.TrimPrefix(s, FilterRegexpPrefix)
		} else if s == "null" {
			operator = "IS"
			v = nil
		}
	}

	return fmt.Sprintf("%s %s ?%s", column, operator, extra), v
}

func init() {
	drivers["sqlite"] = func(settings *mapper.Map) (sqlDriver, error) {
		var (
			d   = sqliteDriver{}
			err error
		)

		if d.path, err = settings.GetString("path", defaultSqliteDriverPath); err != nil {
			return nil, errors.Wrap(err, "sqlite setting 'path'")
		}

		return d, nil
	}
}

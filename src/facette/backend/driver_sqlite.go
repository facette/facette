// +build !disable_sqlite

package backend

import (
	"fmt"
	"strings"

	"github.com/brettlangdon/forge"
)

// sqliteDriver implements the backend database driver interface for SQLite 3.
type sqliteDriver struct{}

func (d sqliteDriver) name() string {
	return "sqlite3"
}

func (d sqliteDriver) buildDSN(config *forge.Section) (string, error) {
	path, err := config.GetString("path")
	if err != nil {
		return "", err
	}

	return path, nil
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
		}
	}

	return fmt.Sprintf("%s %s ?%s", column, operator, extra), v
}

func init() {
	drivers["sqlite"] = func() sqlDriver {
		return sqliteDriver{}
	}
}

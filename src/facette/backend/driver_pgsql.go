// +build !disable_pgsql

package backend

import (
	"fmt"
	"strings"

	"github.com/brettlangdon/forge"
)

// pgsqlDriver implements the backend database driver interface for PostgreSQL.
type pgsqlDriver struct{}

func (d pgsqlDriver) name() string {
	return "postgres"
}

func (d pgsqlDriver) buildDSN(config *forge.Section) (string, error) {
	parts := []string{}
	for key, value := range config.ToMap() {
		parts = append(parts, fmt.Sprintf("%s=%v", key, value))
	}

	return strings.Join(parts, " "), nil
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
	drivers["pgsql"] = func() sqlDriver {
		return pgsqlDriver{}
	}
}

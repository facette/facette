// +build !disable_driver_sqlite

package sqlstorage

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"github.com/facette/maputil"
	"github.com/facette/sliceutil"
	"github.com/jinzhu/gorm"
	"github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

const (
	defaultSqliteDriverPath = "data.db"
)

// sqliteDriver implements the database driver interface for SQLite 3.
type sqliteDriver struct {
	path string
}

func (d sqliteDriver) Name() string {
	return "sqlite3"
}

func (d sqliteDriver) Open() (*sql.DB, error) {
	return sql.Open("sqlite3_ext", d.path)
}

func (d sqliteDriver) Init(db *gorm.DB) error {
	// Enable 'foreign_key' pragma
	return db.Raw("PRAGMA foreign_keys = ON").Error
}

func (d sqliteDriver) NormalizeError(err error) error {
	sqliteErr, ok := err.(sqlite3.Error)
	if !ok {
		return err
	}

	switch sqliteErr.ExtendedCode {
	case sqlite3.ErrConstraintPrimaryKey, sqlite3.ErrConstraintUnique:
		return ErrItemConflict

	case sqlite3.ErrConstraintForeignKey:
		return ErrUnknownReference

	case sqlite3.ErrConstraintNotNull:
		return ErrMissingField
	}

	return err
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

func init() {
	if !sliceutil.Has(sql.Drivers(), "sqlite3_ext") {
		sql.Register("sqlite3_ext", &sqlite3.SQLiteDriver{
			ConnectHook: func(c *sqlite3.SQLiteConn) error {
				return c.RegisterFunc("regexp", regexp.MatchString, true)
			},
		})
	}

	drivers["sqlite"] = func(name string, settings *maputil.Map) (sqlDriver, error) {
		var err error

		d := &sqliteDriver{}

		if d.path, err = settings.GetString("path", defaultSqliteDriverPath); err != nil {
			return nil, errors.Wrap(err, "invalid \"path\" setting")
		}

		return d, nil
	}
}

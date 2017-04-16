// +build !disable_sqlite

package orm

import (
	"os"
	"testing"

	"facette/mapper"
)

var sqliteSettings *mapper.Map

func Test_DriverSqlite3(t *testing.T) {
	testORM(sqliteSettings, t, false)
}

func Test_DriverSqlite3_Tx(t *testing.T) {
	testORM(sqliteSettings, t, true)
}

func init() {
	sqliteSettings = &mapper.Map{
		"path": ":memory:",
	}

	if v := os.Getenv("TEST_SQLITE3_PATH"); v != "" {
		sqliteSettings.Set("path", v)
	}
}

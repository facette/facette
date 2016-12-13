// +build !disable_sqlite3

package orm

import (
	"os"
	"testing"
)

var sqlite3DSN string

func Test_DriverSqlite3(t *testing.T) {
	testORM("sqlite3", sqlite3DSN, t, false)
}

func Test_DriverSqlite3_Tx(t *testing.T) {
	testORM("sqlite3", sqlite3DSN, t, true)
}

func init() {
	sqlite3DSN = ":memory:"
	if v := os.Getenv("TEST_SQLITE3_PATH"); v != "" {
		sqlite3DSN = v
	}
}

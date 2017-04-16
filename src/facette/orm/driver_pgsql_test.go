// +build !disable_pgsql

package orm

import (
	"os"
	"testing"

	"facette/mapper"
)

var pgsqlSettings *mapper.Map

func Test_DriverPostgres(t *testing.T) {
	testORM(pgsqlSettings, t, false)
}

func Test_DriverPostgres_Tx(t *testing.T) {
	testORM(pgsqlSettings, t, true)
}

func init() {
	pgsqlSettings = &mapper.Map{
		"dbname": "orm_test",
	}

	if v := os.Getenv("TEST_POSTGRES_HOST"); v != "" {
		pgsqlSettings.Set("host", v)
	}
	if v := os.Getenv("TEST_POSTGRES_PORT"); v != "" {
		pgsqlSettings.Set("port", v)
	}
	if v := os.Getenv("TEST_POSTGRES_USER"); v != "" {
		pgsqlSettings.Set("user", v)
	}
	if v := os.Getenv("TEST_POSTGRES_PASSWORD"); v != "" {
		pgsqlSettings.Set("password", v)
	}
	if v := os.Getenv("TEST_POSTGRES_DBNAME"); v != "" {
		pgsqlSettings.Set("dbname", v)
	}
}

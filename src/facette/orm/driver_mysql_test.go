// +build !disable_backend_mysql

package orm

import (
	"os"
	"strconv"
	"testing"

	"facette/mapper"
)

var mysqlSettings *mapper.Map

func Test_DriverMySQL(t *testing.T) {
	testORM(mysqlSettings, t, false)
}

func Test_DriverMySQL_Tx(t *testing.T) {
	testORM(mysqlSettings, t, true)
}

func init() {
	mysqlSettings = &mapper.Map{
		"driver":   "mysql",
		"host":     "localhost",
		"port":     3306,
		"user":     "root",
		"password": "",
		"dbname":   "orm_test",
	}

	if v := os.Getenv("TEST_MYSQL_HOST"); v != "" {
		mysqlSettings.Set("host", v)
	}
	if v, err := strconv.Atoi(os.Getenv("TEST_MYSQL_PORT")); err != nil && v != 0 {
		mysqlSettings.Set("port", v)
	}
	if v := os.Getenv("TEST_MYSQL_USER"); v != "" {
		mysqlSettings.Set("user", v)
	}
	if v := os.Getenv("TEST_MYSQL_PASSWORD"); v != "" {
		mysqlSettings.Set("password", v)
	}
	if v := os.Getenv("TEST_MYSQL_DBNAME"); v != "" {
		mysqlSettings.Set("dbname", v)
	}
}

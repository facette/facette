// +build !disable_mysql

package orm

import (
	"fmt"
	"os"
	"strconv"
	"testing"
)

var mysqlDSN string

func Test_DriverMySQL(t *testing.T) {
	testORM("mysql", mysqlDSN, t, false)
}

func Test_DriverMySQL_Tx(t *testing.T) {
	testORM("mysql", mysqlDSN, t, true)
}

func init() {
	host := "localhost"
	port := 3306
	user := "root"
	password := ""
	dbname := "orm_test"

	if v := os.Getenv("TEST_MYSQL_HOST"); v != "" {
		host = v
	}
	if v, err := strconv.Atoi(os.Getenv("TEST_MYSQL_PORT")); err != nil && v != 0 {
		port = v
	}
	if v := os.Getenv("TEST_MYSQL_USER"); v != "" {
		user = v
	}
	if v := os.Getenv("TEST_MYSQL_PASSWORD"); v != "" {
		password = v
	}
	if v := os.Getenv("TEST_MYSQL_DBNAME"); v != "" {
		dbname = v
	}

	mysqlDSN = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?interpolateParams=true", user, password, host, port, dbname)
}

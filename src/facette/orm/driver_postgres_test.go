// +build !disable_postgres

package orm

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

var postgresDSN string

func Test_DriverPostgres(t *testing.T) {
	testORM("postgres", postgresDSN, t, false)
}

func Test_DriverPostgres_Tx(t *testing.T) {
	testORM("postgres", postgresDSN, t, true)
}

func init() {
	config := map[string]interface{}{
		"dbname":  "orm_test",
		"sslmode": "disable",
	}

	if v := os.Getenv("TEST_POSTGRES_HOST"); v != "" {
		config["host"] = v
	}
	if v := os.Getenv("TEST_POSTGRES_PORT"); v != "" {
		config["port"] = v
	}
	if v := os.Getenv("TEST_POSTGRES_USER"); v != "" {
		config["user"] = v
	}
	if v := os.Getenv("TEST_POSTGRES_PASSWORD"); v != "" {
		config["password"] = v
	}
	if v := os.Getenv("TEST_POSTGRES_DBNAME"); v != "" {
		config["dbname"] = v
	}

	parts := []string{}
	for key, value := range config {
		parts = append(parts, fmt.Sprintf("%s=%v", key, value))
	}

	postgresDSN = strings.Join(parts, " ")
}

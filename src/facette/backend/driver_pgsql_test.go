// +build !disable_pgsql

package backend

import "testing"

func Test_PostgreSQL_Provider(t *testing.T) {
	execTestProvider(&pgsqlConfig, t)
}

func Test_PostgreSQL_Collection(t *testing.T) {
	execTestCollection(&pgsqlConfig, t)
}

func Test_PostgreSQL_Graph(t *testing.T) {
	execTestGraph(&pgsqlConfig, t)
}

func Test_PostgreSQL_SourceGroup(t *testing.T) {
	execTestSourceGroup(&pgsqlConfig, t)
}

func Test_PostgreSQL_MetricGroup(t *testing.T) {
	execTestMetricGroup(&pgsqlConfig, t)
}

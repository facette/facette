// +build !disable_mysql

package backend

import "testing"

func Test_MySQL_Provider(t *testing.T) {
	execTestProvider(&mysqlConfig, t)
}

func Test_MySQL_Collection(t *testing.T) {
	execTestCollection(&mysqlConfig, t)
}

func Test_MySQL_Graph(t *testing.T) {
	execTestGraph(&mysqlConfig, t)
}

func Test_MySQL_SourceGroup(t *testing.T) {
	execTestSourceGroup(&mysqlConfig, t)
}

func Test_MySQL_MetricGroup(t *testing.T) {
	execTestMetricGroup(&mysqlConfig, t)
}

func Test_MySQL_Scale(t *testing.T) {
	execTestScale(&mysqlConfig, t)
}

func Test_MySQL_Unit(t *testing.T) {
	execTestUnit(&mysqlConfig, t)
}

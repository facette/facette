// +build !disable_sqlite

package backend

import "testing"

func Test_SQLite_Provider(t *testing.T) {
	execTestProvider(&sqliteConfig, t)
}

func Test_SQLite_Collection(t *testing.T) {
	execTestCollection(&sqliteConfig, t)
}

func Test_SQLite_Graph(t *testing.T) {
	execTestGraph(&sqliteConfig, t)
}

func Test_SQLite_SourceGroup(t *testing.T) {
	execTestSourceGroup(&sqliteConfig, t)
}

func Test_SQLite_MetricGroup(t *testing.T) {
	execTestMetricGroup(&sqliteConfig, t)
}

func Test_SQLite_Scale(t *testing.T) {
	execTestScale(&sqliteConfig, t)
}

func Test_SQLite_Unit(t *testing.T) {
	execTestUnit(&sqliteConfig, t)
}

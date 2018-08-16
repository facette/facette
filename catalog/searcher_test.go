package catalog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var testSearcher = NewSearcher()

func Test_Searcher_Register(t *testing.T) {
	for _, c := range testCatalogs {
		testSearcher.Register(c)
	}
	assert.Equal(t, []*Catalog{
		testCatalogs[0],
		testCatalogs[1],
	}, testSearcher.catalogs)
}

func Test_Searcher_Origins(t *testing.T) {
	expected := make([]*Origin, 2)
	expected[0], _ = testCatalogs[1].Origin("origin2")
	expected[1], _ = testCatalogs[0].Origin("origin2")
	assert.Equal(t, expected, testSearcher.Origins("origin2"))
}

func Test_Searcher_Origins_All(t *testing.T) {
	expected := make([]*Origin, 3)
	expected[0], _ = testCatalogs[0].Origin("origin1")
	expected[1], _ = testCatalogs[1].Origin("origin2")
	expected[2], _ = testCatalogs[0].Origin("origin2")
	assert.Equal(t, expected, testSearcher.Origins(""))
}

func Test_Searcher_Origins_Unknown(t *testing.T) {
	assert.Nil(t, testSearcher.Origins("unknown"))
}

func Test_Searcher_Sources(t *testing.T) {
	expected := make([]*Source, 1)
	expected[0], _ = testCatalogs[0].Source("origin1", "source2")
	assert.Equal(t, expected, testSearcher.Sources("origin1", "source2"))
}

func Test_Searcher_Sources_All(t *testing.T) {
	expected := make([]*Source, 4)
	expected[0], _ = testCatalogs[0].Source("origin1", "source1")
	expected[1], _ = testCatalogs[0].Source("origin1", "source2")
	expected[2], _ = testCatalogs[1].Source("origin2", "source2")
	expected[3], _ = testCatalogs[0].Source("origin2", "source2")
	assert.Equal(t, expected, testSearcher.Sources("", ""))
}

func Test_Searcher_Sources_ByOrigin(t *testing.T) {
	expected := make([]*Source, 2)
	expected[0], _ = testCatalogs[1].Source("origin2", "source2")
	expected[1], _ = testCatalogs[0].Source("origin2", "source2")
	assert.Equal(t, expected, testSearcher.Sources("origin2", ""))
}

func Test_Searcher_Sources_NoOrigin(t *testing.T) {
	expected := make([]*Source, 3)
	expected[0], _ = testCatalogs[0].Source("origin1", "source2")
	expected[1], _ = testCatalogs[1].Source("origin2", "source2")
	expected[2], _ = testCatalogs[0].Source("origin2", "source2")
	assert.Equal(t, expected, testSearcher.Sources("", "source2"))
}

func Test_Searcher_Metrics(t *testing.T) {
	expected := make([]*Metric, 1)
	expected[0], _ = testCatalogs[0].Metric("origin1", "source2", "metric2")
	assert.Equal(t, expected, testSearcher.Metrics("origin1", "source2", "metric2"))
}

func Test_Searcher_Metrics_All(t *testing.T) {
	expected := make([]*Metric, 4)
	expected[0], _ = testCatalogs[0].Metric("origin1", "source1", "metric1")
	expected[1], _ = testCatalogs[0].Metric("origin1", "source2", "metric2")
	expected[2], _ = testCatalogs[1].Metric("origin2", "source2", "metric3")
	expected[3], _ = testCatalogs[0].Metric("origin2", "source2", "metric3")
	assert.Equal(t, expected, testSearcher.Metrics("", "", ""))
}

func Test_Searcher_Metrics_ByOrigin(t *testing.T) {
	expected := make([]*Metric, 2)
	expected[0], _ = testCatalogs[1].Metric("origin2", "source2", "metric3")
	expected[1], _ = testCatalogs[0].Metric("origin2", "source2", "metric3")
	assert.Equal(t, expected, testSearcher.Metrics("origin2", "", ""))
}

func Test_Searcher_Metrics_BySource(t *testing.T) {
	expected := make([]*Metric, 3)
	expected[0], _ = testCatalogs[0].Metric("origin1", "source2", "metric2")
	expected[1], _ = testCatalogs[1].Metric("origin2", "source2", "metric3")
	expected[2], _ = testCatalogs[0].Metric("origin2", "source2", "metric3")
	assert.Equal(t, expected, testSearcher.Metrics("", "source2", ""))
}

func Test_Searcher_Metrics_ByOriginSource(t *testing.T) {
	expected := make([]*Metric, 1)
	expected[0], _ = testCatalogs[0].Metric("origin1", "source2", "metric2")
	assert.Equal(t, expected, testSearcher.Metrics("origin1", "source2", ""))
}

func Test_Searcher_Metrics_NoOrigin(t *testing.T) {
	expected := make([]*Metric, 2)
	expected[0], _ = testCatalogs[1].Metric("origin2", "source2", "metric3")
	expected[1], _ = testCatalogs[0].Metric("origin2", "source2", "metric3")
	assert.Equal(t, expected, testSearcher.Metrics("", "source2", "metric3"))
}

func Test_Searcher_Metrics_NoSource(t *testing.T) {
	expected := make([]*Metric, 1)
	expected[0], _ = testCatalogs[0].Metric("origin1", "source2", "metric2")
	assert.Equal(t, expected, testSearcher.Metrics("origin1", "", "metric2"))
}

func Test_Searcher_Metrics_NoOriginSource(t *testing.T) {
	expected := make([]*Metric, 2)
	expected[0], _ = testCatalogs[1].Metric("origin2", "source2", "metric3")
	expected[1], _ = testCatalogs[0].Metric("origin2", "source2", "metric3")
	assert.Equal(t, expected, testSearcher.Metrics("", "", "metric3"))
}

func Test_Searcher_Unregister(t *testing.T) {
	for _, c := range testCatalogs {
		testSearcher.Unregister(c)
	}
	assert.Equal(t, []*Catalog{}, testSearcher.catalogs)
}

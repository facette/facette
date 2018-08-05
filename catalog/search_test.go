package catalog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testSearcher *Searcher
)

func init() {
	testSearcher = NewSearcher()
}

func Test_Search_Register(t *testing.T) {
	expected := catalogList{testCatalogs[0], testCatalogs[1]}
	for _, c := range testCatalogs {
		testSearcher.Register(c)
	}
	assert.Equal(t, expected, testSearcher.catalogs)
}

func Test_Search_ApplyPriorities(t *testing.T) {
	expected := catalogList{testCatalogs[1], testCatalogs[0]}
	testSearcher.ApplyPriorities()
	assert.Equal(t, expected, testSearcher.catalogs)
}

func Test_Search_Origins(t *testing.T) {
	expected := make([]*Origin, 3)
	expected[0], _ = testCatalogs[1].Origin("origin2")
	expected[1], _ = testCatalogs[0].Origin("origin1")
	expected[2], _ = testCatalogs[0].Origin("origin2")
	assert.Equal(t, expected, testSearcher.Origins("", -1))
}

func Test_Search_Origins_Limit(t *testing.T) {
	expected := make([]*Origin, 1)
	expected[0], _ = testCatalogs[1].Origin("origin2")
	assert.Equal(t, expected, testSearcher.Origins("", 1))
}

func Test_Search_Origins_Name(t *testing.T) {
	expected := make([]*Origin, 2)
	expected[0], _ = testCatalogs[1].Origin("origin2")
	expected[1], _ = testCatalogs[0].Origin("origin2")
	assert.Equal(t, expected, testSearcher.Origins("origin2", -1))
}

func Test_Search_Origins_Name_Limit(t *testing.T) {
	expected := make([]*Origin, 1)
	expected[0], _ = testCatalogs[1].Origin("origin2")
	assert.Equal(t, expected, testSearcher.Origins("origin2", 1))
}

func Test_Search_Sources(t *testing.T) {
	expected := make([]*Source, 3)
	expected[0], _ = testCatalogs[1].Source("origin2", "source2")
	expected[1], _ = testCatalogs[0].Source("origin1", "source1")
	expected[2], _ = testCatalogs[0].Source("origin2", "source2")
	assert.Equal(t, expected, testSearcher.Sources("", "", -1))
}

func Test_Search_Sources_Limit(t *testing.T) {
	expected := make([]*Source, 1)
	expected[0], _ = testCatalogs[1].Source("origin2", "source2")
	assert.Equal(t, expected, testSearcher.Sources("", "", 1))
}

func Test_Search_Sources_Origin(t *testing.T) {
	expected := make([]*Source, 2)
	expected[0], _ = testCatalogs[1].Source("origin2", "source2")
	expected[1], _ = testCatalogs[0].Source("origin2", "source2")
	assert.Equal(t, expected, testSearcher.Sources("origin2", "", -1))
}

func Test_Search_Sources_Origin_Limit(t *testing.T) {
	expected := make([]*Source, 1)
	expected[0], _ = testCatalogs[1].Source("origin2", "source2")
	assert.Equal(t, expected, testSearcher.Sources("origin2", "", 1))
}

func Test_Search_Sources_OriginName(t *testing.T) {
	expected := make([]*Source, 2)
	expected[0], _ = testCatalogs[1].Source("origin2", "source2")
	expected[1], _ = testCatalogs[0].Source("origin2", "source2")
	assert.Equal(t, expected, testSearcher.Sources("origin2", "source2", -1))
}

func Test_Search_Sources_OriginName_Limit(t *testing.T) {
	expected := make([]*Source, 1)
	expected[0], _ = testCatalogs[1].Source("origin2", "source2")
	assert.Equal(t, expected, testSearcher.Sources("origin2", "source2", 1))
}

func Test_Search_Sources_Name(t *testing.T) {
	expected := make([]*Source, 2)
	expected[0], _ = testCatalogs[1].Source("origin2", "source2")
	expected[1], _ = testCatalogs[0].Source("origin2", "source2")
	assert.Equal(t, expected, testSearcher.Sources("", "source2", -1))
}

func Test_Search_Sources_Name_Limit(t *testing.T) {
	expected := make([]*Source, 1)
	expected[0], _ = testCatalogs[1].Source("origin2", "source2")
	assert.Equal(t, expected, testSearcher.Sources("", "source2", 1))
}

func Test_Search_Metrics(t *testing.T) {
	expected := make([]*Metric, 4)
	expected[0], _ = testCatalogs[1].Metric("origin2", "source2", "metric3")
	expected[1], _ = testCatalogs[0].Metric("origin1", "source1", "metric1")
	expected[2], _ = testCatalogs[0].Metric("origin1", "source1", "metric2")
	expected[3], _ = testCatalogs[0].Metric("origin2", "source2", "metric3")
	assert.Equal(t, expected, testSearcher.Metrics("", "", "", -1))
}

func Test_Search_Metrics_Limit(t *testing.T) {
	expected := make([]*Metric, 1)
	expected[0], _ = testCatalogs[1].Metric("origin2", "source2", "metric3")
	assert.Equal(t, expected, testSearcher.Metrics("", "", "", 1))
}

func Test_Search_Metrics_Origin(t *testing.T) {
	expected := make([]*Metric, 2)
	expected[0], _ = testCatalogs[0].Metric("origin1", "source1", "metric1")
	expected[1], _ = testCatalogs[0].Metric("origin1", "source1", "metric2")
	assert.Equal(t, expected, testSearcher.Metrics("origin1", "", "", -1))
}

func Test_Search_Metrics_Origin_Limit(t *testing.T) {
	expected := make([]*Metric, 1)
	expected[0], _ = testCatalogs[0].Metric("origin1", "source1", "metric1")
	assert.Equal(t, expected, testSearcher.Metrics("origin1", "", "", 1))
}

func Test_Search_Metrics_OriginSource(t *testing.T) {
	expected := make([]*Metric, 2)
	expected[0], _ = testCatalogs[0].Metric("origin1", "source1", "metric1")
	expected[1], _ = testCatalogs[0].Metric("origin1", "source1", "metric2")
	assert.Equal(t, expected, testSearcher.Metrics("origin1", "source1", "", -1))
}

func Test_Search_Metrics_OriginSource_Limit(t *testing.T) {
	expected := make([]*Metric, 1)
	expected[0], _ = testCatalogs[0].Metric("origin1", "source1", "metric1")
	assert.Equal(t, expected, testSearcher.Metrics("origin1", "source1", "", 1))
}

func Test_Search_Metrics_OriginSourceName(t *testing.T) {
	expected := make([]*Metric, 2)
	expected[0], _ = testCatalogs[1].Metric("origin2", "source2", "metric3")
	expected[1], _ = testCatalogs[0].Metric("origin2", "source2", "metric3")
	assert.Equal(t, expected, testSearcher.Metrics("origin2", "source2", "metric3", -1))
}

func Test_Search_Metrics_OriginSourceName_Limit(t *testing.T) {
	expected := make([]*Metric, 1)
	expected[0], _ = testCatalogs[1].Metric("origin2", "source2", "metric3")
	assert.Equal(t, expected, testSearcher.Metrics("origin2", "source2", "metric3", 1))
}

func Test_Search_Metrics_Source(t *testing.T) {
	expected := make([]*Metric, 2)
	expected[0], _ = testCatalogs[0].Metric("origin1", "source1", "metric1")
	expected[1], _ = testCatalogs[0].Metric("origin1", "source1", "metric2")
	assert.Equal(t, expected, testSearcher.Metrics("", "source1", "", -1))
}

func Test_Search_Metrics_Source_Limit(t *testing.T) {
	expected := make([]*Metric, 1)
	expected[0], _ = testCatalogs[0].Metric("origin1", "source1", "metric1")
	assert.Equal(t, expected, testSearcher.Metrics("", "source1", "", 1))
}

func Test_Search_Metrics_SourceName(t *testing.T) {
	expected := make([]*Metric, 2)
	expected[0], _ = testCatalogs[1].Metric("origin2", "source2", "metric3")
	expected[1], _ = testCatalogs[0].Metric("origin2", "source2", "metric3")
	assert.Equal(t, expected, testSearcher.Metrics("", "source2", "metric3", -1))
}

func Test_Search_Metrics_SourceName_Limit(t *testing.T) {
	expected := make([]*Metric, 1)
	expected[0], _ = testCatalogs[1].Metric("origin2", "source2", "metric3")
	assert.Equal(t, expected, testSearcher.Metrics("", "source2", "metric3", 1))
}

func Test_Search_Metrics_Name(t *testing.T) {
	expected := make([]*Metric, 2)
	expected[0], _ = testCatalogs[1].Metric("origin2", "source2", "metric3")
	expected[1], _ = testCatalogs[0].Metric("origin2", "source2", "metric3")
	assert.Equal(t, expected, testSearcher.Metrics("", "", "metric3", -1))
}

func Test_Search_Metrics_Name_Limit(t *testing.T) {
	expected := make([]*Metric, 1)
	expected[0], _ = testCatalogs[1].Metric("origin2", "source2", "metric3")
	assert.Equal(t, expected, testSearcher.Metrics("", "", "metric3", 1))
}

func Test_Search_Unregister(t *testing.T) {
	expected := catalogList{}
	for _, c := range testCatalogs {
		testSearcher.Unregister(c)
	}
	assert.Equal(t, expected, testSearcher.catalogs)
}

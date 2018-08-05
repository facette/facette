package catalog

import (
	"sort"
	"testing"

	"facette.io/sliceutil"
	"github.com/stretchr/testify/assert"
)

var (
	testCatalogs catalogList
	testRecords  []*Record
)

func init() {
	testCatalogs = catalogList{
		NewCatalog("catalog1"),
		NewCatalog("catalog2"),
	}

	testRecords = []*Record{
		&Record{Origin: "origin1", Source: "source1", Metric: "metric1"},
		&Record{Origin: "origin1", Source: "source1", Metric: "metric2"},
		&Record{Origin: "origin2", Source: "source2", Metric: "metric3"},
	}

	for _, r := range testRecords {
		testCatalogs[0].Insert(r)
	}
	testCatalogs[1].Insert(testRecords[2])
}

func Test_Catalog_Name(t *testing.T) {
	assert.Equal(t, "catalog1", testCatalogs[0].Name())
}

func Test_Catalog_SetPriority(t *testing.T) {
	testCatalogs[0].SetPriority(10)
	assert.Equal(t, 10, testCatalogs[0].priority)
}

func Test_Catalog_Origin(t *testing.T) {
	origin, err := testCatalogs[0].Origin(testRecords[0].Origin)
	assert.Nil(t, err)
	assert.Equal(t, testCatalogs[0], origin.Catalog())

	expected := []string{}
	for _, r := range testRecords {
		if r.Origin == testRecords[0].Origin && !sliceutil.Has(expected, r.Source) {
			expected = append(expected, r.Source)
		}
	}
	sort.Strings(expected)

	actual := []string{}
	for _, s := range origin.Sources() {
		actual = append(actual, s.Name)
	}

	assert.Equal(t, expected, actual)
}

func Test_Catalog_Origin_Unknown(t *testing.T) {
	origin, err := testCatalogs[0].Origin("unknown")
	assert.Equal(t, ErrUnknownOrigin, err)
	assert.Nil(t, origin)
}

func Test_Catalog_Origin_Sources(t *testing.T) {
	origin, _ := testCatalogs[0].Origin(testRecords[0].Origin)
	actual := []string{}
	for _, o := range origin.Sources() {
		actual = append(actual, o.Name)
	}

	assert.Equal(t, []string{"source1"}, actual)
}

func Test_Catalog_Origins(t *testing.T) {
	expected := []string{}
	for _, r := range testRecords {
		if !sliceutil.Has(expected, r.Origin) {
			expected = append(expected, r.Origin)
		}
	}
	sort.Strings(expected)

	actual := []string{}
	for _, o := range testCatalogs[0].Origins() {
		actual = append(actual, o.Name)
	}

	assert.Equal(t, expected, actual)
}

func Test_Catalog_Source(t *testing.T) {
	origin, _ := testCatalogs[0].Origin(testRecords[1].Origin)

	source, err := testCatalogs[0].Source(testRecords[1].Origin, testRecords[1].Source)
	assert.Nil(t, err)
	assert.Equal(t, origin, source.Origin())

	expected := []string{}
	for _, r := range testRecords {
		if r.Origin == testRecords[1].Origin && r.Source == testRecords[1].Source &&
			!sliceutil.Has(expected, r.Metric) {
			expected = append(expected, r.Metric)
		}
	}
	sort.Strings(expected)

	actual := []string{}
	for _, m := range source.Metrics() {
		actual = append(actual, m.Name)
	}
	sort.Strings(actual)

	assert.Equal(t, expected, actual)
}

func Test_Catalog_Source_Unknown(t *testing.T) {
	source, err := testCatalogs[0].Source("unknown", testRecords[1].Source)
	assert.Equal(t, ErrUnknownOrigin, err)
	assert.Nil(t, source)

	source, err = testCatalogs[0].Source(testRecords[1].Origin, "unknown")
	assert.Equal(t, ErrUnknownSource, err)
	assert.Nil(t, source)
}

func Test_Catalog_Source_Metrics(t *testing.T) {
	source, _ := testCatalogs[0].Source(testRecords[0].Origin, testRecords[0].Source)
	actual := []string{}
	for _, o := range source.Metrics() {
		actual = append(actual, o.Name)
	}

	assert.Equal(t, []string{"metric1", "metric2"}, actual)
}

func Test_Catalog_Metric(t *testing.T) {
	source, _ := testCatalogs[0].Source(testRecords[2].Origin, testRecords[2].Source)

	metric, err := testCatalogs[0].Metric(testRecords[2].Origin, testRecords[2].Source, testRecords[2].Metric)
	assert.Nil(t, err)
	assert.Equal(t, source, metric.Source())
}

func Test_Catalog_Metric_Unknown(t *testing.T) {
	metric, err := testCatalogs[0].Metric("unknown", testRecords[2].Source, testRecords[2].Metric)
	assert.Equal(t, ErrUnknownOrigin, err)
	assert.Nil(t, metric)

	metric, err = testCatalogs[0].Metric(testRecords[2].Origin, "unknown", testRecords[2].Metric)
	assert.Equal(t, ErrUnknownSource, err)
	assert.Nil(t, metric)

	metric, err = testCatalogs[0].Metric(testRecords[2].Origin, testRecords[2].Source, "unknown")
	assert.Equal(t, ErrUnknownMetric, err)
	assert.Nil(t, metric)
}

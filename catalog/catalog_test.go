package catalog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testCatalogs []*Catalog
	testRecords  []*Record
)

func init() {
	testCatalogs = []*Catalog{
		New("catalog1", nil),
		New("catalog2", nil),
	}

	testRecords = []*Record{
		&Record{Origin: "origin1", Source: "source1", Metric: "metric1"},
		&Record{Origin: "origin1", Source: "source2", Metric: "metric2"},
		&Record{Origin: "origin2", Source: "source2", Metric: "metric3"},
	}

	for _, r := range testRecords {
		testCatalogs[0].Insert(r)
	}

	testCatalogs[1].Priority = 100
	testCatalogs[1].Insert(testRecords[2])
}

func Test_Catalog_Origin(t *testing.T) {
	actual, err := testCatalogs[0].Origin("origin1")
	assert.Nil(t, err)
	assert.Equal(t, "origin1", actual.Name)

	actual, err = testCatalogs[0].Origin("origin2")
	assert.Nil(t, err)
	assert.Equal(t, "origin2", actual.Name)

	actual, err = testCatalogs[1].Origin("origin2")
	assert.Nil(t, err)
	assert.Equal(t, "origin2", actual.Name)
}

func Test_Catalog_Origin_Unknown(t *testing.T) {
	actual, err := testCatalogs[1].Origin("unknown")
	assert.Equal(t, ErrUnknownOrigin, err)
	assert.Nil(t, actual)
}

func Test_Catalog_Source(t *testing.T) {
	actual, err := testCatalogs[0].Source("origin1", "source1")
	assert.Nil(t, err)
	assert.Equal(t, "source1", actual.Name)

	actual, err = testCatalogs[0].Source("origin2", "source2")
	assert.Nil(t, err)
	assert.Equal(t, "source2", actual.Name)

	actual, err = testCatalogs[1].Source("origin2", "source2")
	assert.Nil(t, err)
	assert.Equal(t, "source2", actual.Name)
}

func Test_Catalog_Source_Unknown(t *testing.T) {
	actual, err := testCatalogs[0].Source("unknown", "source1")
	assert.Equal(t, ErrUnknownOrigin, err)
	assert.Nil(t, actual)

	actual, err = testCatalogs[0].Source("origin1", "unknown")
	assert.Equal(t, ErrUnknownSource, err)
	assert.Nil(t, actual)
}

func Test_Catalog_Metric(t *testing.T) {
	actual, err := testCatalogs[0].Metric("origin1", "source1", "metric1")
	assert.Nil(t, err)
	assert.Equal(t, "metric1", actual.Name)

	actual, err = testCatalogs[0].Metric("origin1", "source2", "metric2")
	assert.Nil(t, err)
	assert.Equal(t, "metric2", actual.Name)

	actual, err = testCatalogs[0].Metric("origin2", "source2", "metric3")
	assert.Nil(t, err)
	assert.Equal(t, "metric3", actual.Name)

	actual, err = testCatalogs[1].Metric("origin2", "source2", "metric3")
	assert.Nil(t, err)
	assert.Equal(t, "metric3", actual.Name)
}

func Test_Catalog_Metric_Unknown(t *testing.T) {
	actual, err := testCatalogs[0].Metric("unknown", "source1", "metric1")
	assert.Equal(t, ErrUnknownOrigin, err)
	assert.Nil(t, actual)

	actual, err = testCatalogs[0].Metric("origin1", "unknown", "metric1")
	assert.Equal(t, ErrUnknownSource, err)
	assert.Nil(t, actual)

	actual, err = testCatalogs[0].Metric("origin1", "source1", "unknown")
	assert.Equal(t, ErrUnknownMetric, err)
	assert.Nil(t, actual)
}

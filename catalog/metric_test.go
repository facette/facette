package catalog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Metric_Catalog(t *testing.T) {
	metric, err := testCatalogs[0].Metric("origin1", "source1", "metric1")
	assert.Nil(t, err)
	assert.Equal(t, testCatalogs[0], metric.Catalog())
}

func Test_Metric_Origin(t *testing.T) {
	metric, err := testCatalogs[0].Metric("origin1", "source1", "metric1")
	assert.Nil(t, err)
	assert.Equal(t, testCatalogs[0].Origins["origin1"], metric.Origin())
}

func Test_Metric_Source(t *testing.T) {
	metric, err := testCatalogs[0].Metric("origin1", "source1", "metric1")
	assert.Nil(t, err)
	assert.Equal(t, testCatalogs[0].Origins["origin1"].Sources["source1"], metric.Source())
}

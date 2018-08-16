package catalog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Source_Catalog(t *testing.T) {
	source, err := testCatalogs[0].Source("origin1", "source1")
	assert.Nil(t, err)
	assert.Equal(t, testCatalogs[0], source.Catalog())
}

func Test_Source_Origin(t *testing.T) {
	source, err := testCatalogs[0].Source("origin1", "source1")
	assert.Nil(t, err)
	assert.Equal(t, testCatalogs[0].Origins["origin1"], source.Origin())
}

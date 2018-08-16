package catalog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Origin_Catalog(t *testing.T) {
	origin, err := testCatalogs[0].Origin("origin1")
	assert.Nil(t, err)
	assert.Equal(t, testCatalogs[0], origin.Catalog())
}

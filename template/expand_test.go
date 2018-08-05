package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Expand(t *testing.T) {
	actual, err := Expand("This {{ .a }} a {{ .b }}!", map[string]interface{}{"a": "is", "b": "sample text"})
	assert.Nil(t, err)
	assert.Equal(t, "This is a sample text!", actual)
}

func Test_Expand_Empty(t *testing.T) {
	actual, err := Expand("This {{ .a }} a {{ .b }}!", map[string]interface{}{"a": "is"})
	assert.Nil(t, err)
	assert.Equal(t, "This is a !", actual)
}

func Test_Expand_SyntaxFail(t *testing.T) {
	actual, err := Expand("This {{ .a } a {{ .b }}!", map[string]interface{}{"a": "is", "b": "sample text"})
	assert.Equal(t, ErrInvalidTemplate, err)
	assert.Equal(t, "", actual)
}

func Test_Expand_IdentFail(t *testing.T) {
	actual, err := Expand("This {{ .a }} a {{ b }}!", map[string]interface{}{"a": "is", "b": "sample text"})
	assert.Equal(t, ErrInvalidTemplate, err)
	assert.Equal(t, "", actual)
}

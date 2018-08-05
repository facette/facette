package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Parse(t *testing.T) {
	actual, err := Parse("This {{ .a }} a {{ .b }}!")
	assert.Nil(t, err)
	assert.Equal(t, []string{"a", "b"}, actual)
}

func Test_Parse_SyntaxFail(t *testing.T) {
	actual, err := Parse("This {{ .a } a {{ .b }}!")
	assert.Equal(t, ErrInvalidTemplate, err)
	assert.Nil(t, actual)
}

func Test_Parse_IdentFail(t *testing.T) {
	actual, err := Parse("This {{ .a }} a {{ b }}!")
	assert.Equal(t, ErrInvalidTemplate, err)
	assert.Nil(t, actual)
}

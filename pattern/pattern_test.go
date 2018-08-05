package pattern

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Match_Glob(t *testing.T) {
	ok, err := Match("glob:*b*", "abc")
	assert.Nil(t, err)
	assert.True(t, ok)
}

func Test_Match_Glob_Not(t *testing.T) {
	ok, err := Match("glob:*b*", "def")
	assert.Nil(t, err)
	assert.False(t, ok)
}

func Test_Match_Regexp(t *testing.T) {
	ok, err := Match("regexp:^[a-z]{3}$", "abc")
	assert.Nil(t, err)
	assert.True(t, ok)
}

func Test_Match_Regexp_Invalid(t *testing.T) {
	ok, err := Match("regexp:^[a-z", "abc")
	assert.NotNil(t, err)
	assert.False(t, ok)
}

func Test_Match_Regexp_Not(t *testing.T) {
	ok, err := Match("regexp:^[a-z]{3}$", "abcd")
	assert.Nil(t, err)
	assert.False(t, ok)
}

func Test_Match_Simple(t *testing.T) {
	ok, err := Match("a", "a")
	assert.Nil(t, err)
	assert.True(t, ok)
}

func Test_Match_Simple_Not(t *testing.T) {
	ok, err := Match("a", "b")
	assert.Nil(t, err)
	assert.False(t, ok)
}

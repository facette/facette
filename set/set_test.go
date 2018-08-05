package set

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Set(t *testing.T) {
	s := New("a", "b", "c")
	assert.Equal(t, 3, s.Len())
	assert.True(t, s.Has("a", "b", "c"))
	assert.False(t, s.Has(42))

	s.Add(42)
	assert.Equal(t, 4, s.Len())
	assert.True(t, s.Has("a", "b", "c", 42))

	s.Remove("a", 42)
	assert.Equal(t, 2, s.Len())
	assert.True(t, s.Has("b", "c"))
	assert.False(t, s.Has("a", 42))
}

func Test_StringSlice(t *testing.T) {
	s := New("a", "b", "c")
	actual := StringSlice(s)
	sort.Strings(actual)
	assert.Equal(t, []string{"a", "b", "c"}, actual)

	s.Add(42)
	actual = StringSlice(s)
	sort.Strings(actual)
	assert.Equal(t, []string{"42", "a", "b", "c"}, actual)

	s.Remove("a", 42)
	actual = StringSlice(s)
	sort.Strings(actual)
	assert.Equal(t, []string{"b", "c"}, actual)
}

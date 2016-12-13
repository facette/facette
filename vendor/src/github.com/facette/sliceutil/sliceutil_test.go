package sliceutil

import (
	"testing"
	"time"
)

func Test_IndexOf(t *testing.T) {
	now := time.Now()

	tests := []struct {
		slice  interface{}
		search interface{}
		index  int
	}{
		{[]int{1, 2, 3}, 3, 2},
		{[]int{1, 2, 3}, 4, -1},
		{[]int{1, 2, 3}, "a", -1},
		{[]string{"a", "b", "c"}, "a", 0},
		{[]string{"a", "b", "c"}, "d", -1},
		{[]string{"a", "b", "c"}, true, -1},
		{[]interface{}{"a", 2, false, now}, false, 2},
		{[]interface{}{"a", 2, false, now}, now, 3},
		{[]interface{}{"a", 2, false, now}, true, -1},
		{[]interface{}{"a", 2, false, now}, "", -1},
	}

	for _, entry := range tests {
		if result := IndexOf(entry.slice, entry.search); result != entry.index {
			t.Logf("\nExpected %d\nbut got  %d", entry.index, result)
			t.Fail()
		}
	}
}

func Test_Has(t *testing.T) {
	now := time.Now()

	tests := []struct {
		slice  interface{}
		search interface{}
		found  bool
	}{
		{[]int{1, 2, 3}, 3, true},
		{[]int{1, 2, 3}, 4, false},
		{[]int{1, 2, 3}, "a", false},
		{[]string{"a", "b", "c"}, "a", true},
		{[]string{"a", "b", "c"}, "d", false},
		{[]string{"a", "b", "c"}, true, false},
		{[]interface{}{"a", 2, false, now}, false, true},
		{[]interface{}{"a", 2, false, now}, now, true},
		{[]interface{}{"a", 2, false, now}, true, false},
		{[]interface{}{"a", 2, false, now}, "", false},
	}

	for _, entry := range tests {
		if result := Has(entry.slice, entry.search); result != entry.found {
			t.Logf("\nExpected %d\nbut got  %d", entry.found, result)
			t.Fail()
		}
	}
}

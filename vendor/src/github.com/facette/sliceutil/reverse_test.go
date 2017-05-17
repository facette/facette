// +build go1.8

package sliceutil

import (
	"reflect"
	"testing"
	"time"
)

func Test_Reverse(t *testing.T) {
	now := time.Now()

	tests := []struct {
		slice    interface{}
		expected interface{}
	}{
		{[]int{1, 2, 3}, []int{3, 2, 1}},
		{[]int{1, 2, 3, 4}, []int{4, 3, 2, 1}},
		{[]string{"a", "b", "c"}, []string{"c", "b", "a"}},
		{[]string{"a", "b", "c", "d"}, []string{"d", "c", "b", "a"}},
		{[]interface{}{"a", 1, false}, []interface{}{false, 1, "a"}},
		{[]interface{}{"a", 1, false, now}, []interface{}{now, false, 1, "a"}},
	}

	for _, entry := range tests {
		result := entry.slice
		Reverse(result)

		if !reflect.DeepEqual(result, entry.expected) {
			t.Logf("\nExpected %#v\nbut got  %#v", entry.expected, result)
			t.Fail()
		}
	}
}

// +build go1.8

package sliceutil

import "reflect"

// Reverse reverses the items order in a slice.
func Reverse(s interface{}) {
	// Check for slice type
	rv := reflect.ValueOf(s)
	if rv.Kind() != reflect.Slice {
		return
	}

	// Interate on slice elements
	swap := reflect.Swapper(s)
	for i, n := 0, rv.Len(); i < n/2; i++ {
		swap(i, n-(i+1))
	}
}

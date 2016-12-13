package sliceutil

import "reflect"

// IndexOf returns the index of a given value in a slice, or -1 if not found
func IndexOf(s, v interface{}) int {
	// Check for slice type
	rv := reflect.ValueOf(s)
	if rv.Kind() != reflect.Slice {
		return -1
	}

	// Interate on slice elements
	for i := 0; i < rv.Len(); i++ {
		if reflect.DeepEqual(rv.Index(i).Interface(), v) {
			return i
		}
	}

	return -1
}

// Has checks whether a given value is present in a slice or not.
func Has(s, v interface{}) bool {
	return IndexOf(s, v) != -1
}

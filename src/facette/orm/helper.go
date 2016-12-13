package orm

import (
	"reflect"
	"regexp"
	"strings"
)

var nameRegexp = regexp.MustCompile("([A-Z]?[a-z0-9]+|^[A-Z]+$|^[A-Z]|[A-Z]+)")

// FormatName formats table and columns names to snake case.
func FormatName(name string) string {
	return strings.ToLower(strings.Join(nameRegexp.FindAllString(name, -1), "_"))
}

// IsDefault returns whether or not a value equals its default value.
func IsDefault(rv reflect.Value) bool {
	return reflect.DeepEqual(rv.Interface(), reflect.Zero(rv.Type()).Interface())
}

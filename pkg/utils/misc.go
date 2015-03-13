package utils

import (
	"bytes"
	"encoding/gob"
	"path"
	"regexp"
	"strings"
)

// Clone performs a deep copy of an interface.
func Clone(src, dst interface{}) {
	buffer := new(bytes.Buffer)

	encoder := gob.NewEncoder(buffer)
	encoder.Encode(src)

	decoder := gob.NewDecoder(buffer)
	decoder.Decode(dst)
}

// FilterMatch checks a glob or a regexp pattern over a given value.
func FilterMatch(pattern, value string) bool {
	if strings.HasPrefix(pattern, "glob:") {
		// Remove slashes from pattern and value as `path.Match' does not handle them
		pattern = strings.ToLower(strings.Replace(pattern, "/", "\x1e", -1))
		value = strings.ToLower(strings.Replace(value, "/", "\x1e", -1))

		ok, _ := path.Match(pattern[5:], value)
		return ok
	} else if strings.HasPrefix(pattern, "regexp:") {
		re := regexp.MustCompile(pattern[7:])
		return re.MatchString(value)
	}

	return pattern == value
}

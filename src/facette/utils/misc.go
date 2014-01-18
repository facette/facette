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
	var (
		buffer  *bytes.Buffer
		decoder *gob.Decoder
		encoder *gob.Encoder
	)

	buffer = new(bytes.Buffer)

	encoder = gob.NewEncoder(buffer)
	decoder = gob.NewDecoder(buffer)

	encoder.Encode(src)
	decoder.Decode(dst)
}

// FilterMatch checks a glob or regexp pattern over a given value.
func FilterMatch(pattern, value string) bool {
	var (
		re *regexp.Regexp
	)

	if strings.HasPrefix(pattern, "glob:") {
		ok, _ := path.Match(pattern[5:], value)
		return ok
	} else if strings.HasPrefix(pattern, "regexp:") {
		re = regexp.MustCompile(pattern[7:])
		return re.MatchString(value)
	}

	return pattern == value
}

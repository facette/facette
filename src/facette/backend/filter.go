package backend

import (
	"path"
	"regexp"
	"strings"
)

// FilterMatch checks if a glob or a regexp pattern matches a given value.
func FilterMatch(pattern, value string) bool {
	if strings.HasPrefix(pattern, FilterGlobPrefix) {
		// Remove slashes from pattern and value as 'path.Match' does not handle them
		pattern = strings.ToLower(strings.Replace(pattern, "/", "\x1e", -1))
		value = strings.ToLower(strings.Replace(value, "/", "\x1e", -1))

		ok, _ := path.Match(strings.TrimPrefix(pattern, FilterGlobPrefix), value)
		return ok
	} else if strings.HasPrefix(pattern, FilterRegexpPrefix) {
		return regexp.MustCompile(strings.TrimPrefix(pattern, FilterRegexpPrefix)).MatchString(value)
	}

	return pattern == value
}

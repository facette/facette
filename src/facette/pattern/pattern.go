// package pattern provides pattern matching functions.
package pattern

import (
	"path"
	"regexp"
	"strings"
)

const (
	// GlobPrefix represents the glob pattern prefix.
	GlobPrefix = "glob:"

	// RegexpPrefix represents the regexp pattern prefix.
	RegexpPrefix = "regexp:"
)

// Match returns true if the value matches the pattern, or an error if pattern compilation fails. If the pattern is prefixed with "glob:" the value will be evaluated using shell-style globbing, if it prefixed with "regexp:" it will be evaluated using regular expression matching.
func Match(pattern, value string) (bool, error) {
	if strings.HasPrefix(pattern, GlobPrefix) {
		// Remove slashes from pattern and value as 'path.Match' does not handle them
		pattern = strings.ToLower(strings.Replace(pattern, "/", "\x1e", -1))
		value = strings.ToLower(strings.Replace(value, "/", "\x1e", -1))

		ok, _ := path.Match(strings.TrimPrefix(pattern, GlobPrefix), value)
		return ok, nil
	} else if strings.HasPrefix(pattern, RegexpPrefix) {
		re, err := regexp.Compile(strings.TrimPrefix(pattern, RegexpPrefix))
		if err != nil {
			return false, err
		}

		return re.MatchString(value), nil
	}

	return pattern == value, nil
}

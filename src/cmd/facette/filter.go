package main

import (
	"path"
	"regexp"
	"strings"

	"github.com/facette/sqlstorage"
)

const (
	filterGlobPrefix   = "glob:"
	filterRegexpPrefix = "regexp:"
)

func filterApplyModifier(pattern string) interface{} {
	if strings.HasPrefix(pattern, filterGlobPrefix) {
		return sqlstorage.GlobModifier(strings.TrimPrefix(pattern, filterGlobPrefix))
	} else if strings.HasPrefix(pattern, filterRegexpPrefix) {
		return sqlstorage.RegexpModifier(strings.TrimPrefix(pattern, filterRegexpPrefix))
	}

	return pattern
}

func filterMatch(pattern, value string) bool {
	if strings.HasPrefix(pattern, filterGlobPrefix) {
		// Remove slashes from pattern and value as 'path.Match' does not handle them
		pattern = strings.ToLower(strings.Replace(pattern, "/", "\x1e", -1))
		value = strings.ToLower(strings.Replace(value, "/", "\x1e", -1))

		ok, _ := path.Match(strings.TrimPrefix(pattern, filterGlobPrefix), value)
		return ok
	} else if strings.HasPrefix(pattern, filterRegexpPrefix) {
		return regexp.MustCompile(strings.TrimPrefix(pattern, filterRegexpPrefix)).MatchString(value)
	}

	return pattern == value
}

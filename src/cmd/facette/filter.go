package main

import (
	"path"
	"regexp"
	"strings"

	"facette/backend"

	"github.com/facette/sqlstorage"
)

func filterApplyModifier(pattern string) interface{} {
	if strings.HasPrefix(pattern, backend.GlobPrefix) {
		return sqlstorage.GlobModifier(strings.TrimPrefix(pattern, backend.GlobPrefix))
	} else if strings.HasPrefix(pattern, backend.RegexpPrefix) {
		return sqlstorage.RegexpModifier(strings.TrimPrefix(pattern, backend.RegexpPrefix))
	}

	return pattern
}

func filterMatch(pattern, value string) (bool, error) {
	if strings.HasPrefix(pattern, backend.GlobPrefix) {
		// Remove slashes from pattern and value as 'path.Match' does not handle them
		pattern = strings.ToLower(strings.Replace(pattern, "/", "\x1e", -1))
		value = strings.ToLower(strings.Replace(value, "/", "\x1e", -1))

		ok, _ := path.Match(strings.TrimPrefix(pattern, backend.GlobPrefix), value)
		return ok, nil
	} else if strings.HasPrefix(pattern, backend.RegexpPrefix) {
		if re, err := regexp.Compile(strings.TrimPrefix(pattern, backend.RegexpPrefix)); err != nil {
			return false, err
		}

		return re.MatchString(value), nil
	}

	return pattern == value, nil
}

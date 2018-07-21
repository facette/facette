package main

import (
	"strings"

	"facette.io/facette/pattern"
	"facette.io/sqlstorage"
)

func filterApplyModifier(input string) interface{} {
	if strings.HasPrefix(input, pattern.GlobPrefix) {
		return sqlstorage.GlobModifier(strings.TrimPrefix(input, pattern.GlobPrefix))
	} else if strings.HasPrefix(input, pattern.RegexpPrefix) {
		return sqlstorage.RegexpModifier(strings.TrimPrefix(input, pattern.RegexpPrefix))
	}

	return input
}

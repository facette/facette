package server

import (
	"html/template"
	"reflect"
	"strings"
)

func templateEqual(x, y interface{}) bool {
	return reflect.DeepEqual(x, y)
}

func templateNotEqual(x, y interface{}) bool {
	return !templateEqual(x, y)
}

func templateDumpMap(x map[string]string) string {
	var (
		chunks []string
	)

	for key, value := range x {
		chunks = append(chunks, key+": "+value)
	}

	return strings.Join(chunks, "; ")
}

func templateSubstr(x string, start, end int) string {
	return x[start:end]
}

func templateHighlight(x, y string) template.HTML {
	return template.HTML(strings.Replace(x, y, "<strong>"+y+"</strong>", -1))
}

package server

import (
	"html/template"
	"reflect"
	"strconv"
	"strings"
)

func (server *Server) templateAsset(x string) string {
	return x + "?" + strconv.FormatInt(server.startTime.Unix(), 10)
}

func templateEqual(x, y interface{}) bool {
	return reflect.DeepEqual(x, y)
}

func templateNotEqual(x, y interface{}) bool {
	return !templateEqual(x, y)
}

func templateDumpMap(x map[string]string) string {
	chunks := make([]string, 0)

	for key, value := range x {
		if value == "" {
			continue
		}

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

package server

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"path"
	"reflect"
	"strconv"
	"strings"

	"github.com/facette/facette/pkg/utils"
)

func (server *Server) execTemplate(writer http.ResponseWriter, status int, data interface{}, files ...string) error {
	var err error

	tmpl := template.New(path.Base(files[0])).Funcs(template.FuncMap{
		"asset":  server.templateAsset,
		"dump":   templateDumpMap,
		"eq":     templateEqual,
		"hl":     templateHighlight,
		"ne":     templateNotEqual,
		"substr": templateSubstr,
	})

	// Execute template
	tmpl, err = tmpl.ParseFiles(files...)
	if err != nil {
		return err
	}

	tmplData := bytes.NewBuffer(nil)

	err = tmpl.Execute(tmplData, data)
	if err != nil {
		return err
	}

	writer.WriteHeader(status)

	if utils.HTTPGetContentType(writer) == "text/xml" {
		writer.Write([]byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n"))
	}

	writer.Write(tmplData.Bytes())

	return err
}

func (server *Server) templateAsset(x string) string {
	return x + "?" + strconv.FormatInt(server.startTime.Unix(), 10)
}

func templateEqual(x, y interface{}) bool {
	return reflect.DeepEqual(x, y)
}

func templateNotEqual(x, y interface{}) bool {
	return !templateEqual(x, y)
}

func templateDumpMap(x map[string]interface{}) string {
	chunks := make([]string, 0)

	for key, value := range x {
		if value == "" {
			continue
		}

		switch value.(type) {
		case []interface{}:
			valueString := make([]string, len(value.([]interface{})))

			for index, entry := range value.([]interface{}) {
				valueString[index] = fmt.Sprintf("%v", entry)
			}

			chunks = append(chunks, fmt.Sprintf("%s: %v", key, strings.Join(valueString, ", ")))

		default:
			chunks = append(chunks, fmt.Sprintf("%s: %v", key, value))
		}
	}

	return strings.Join(chunks, "; ")
}

func templateSubstr(x string, start, end int) string {
	return x[start:end]
}

func templateHighlight(x, y string) template.HTML {
	return template.HTML(strings.Replace(x, y, "<strong>"+y+"</strong>", -1))
}

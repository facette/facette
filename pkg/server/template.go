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
	txttemplate "text/template"

	"github.com/facette/facette/pkg/utils"
)

func (server *Server) execTemplate(writer http.ResponseWriter, status int, data interface{}, files ...string) error {
	var err error

	tmpl := template.New(path.Base(files[0])).Funcs(template.FuncMap{
		"asset":  server.templateAsset,
		"dump":   templateDumpMap,
		"eq":     templateEqual,
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

func expandStringTemplate(t string, attr map[string]interface{}) (string, error) {
	var (
		buf *bytes.Buffer
		tpl *txttemplate.Template
		err error
	)

	buf = bytes.NewBuffer(nil)

	if tpl, err = txttemplate.New("template").Parse(t); err != nil {
		return t, fmt.Errorf("error while parsing template: %s", err)
	}

	if err = tpl.Execute(buf, attr); err != nil {
		return t, fmt.Errorf("error while executing template: %s", err)
	}

	return buf.String(), nil
}

func templateEqual(x, y interface{}) bool {
	return reflect.DeepEqual(x, y)
}

func templateNotEqual(x, y interface{}) bool {
	return !templateEqual(x, y)
}

func templateDumpMap(x map[string]interface{}) string {
	var chunks []string

	for key, value := range x {
		if value == "" || value == nil {
			continue
		}

		switch value.(type) {
		case []interface{}:
			valueLength := len(value.([]interface{}))
			if valueLength == 0 {
				continue
			}

			valueString := make([]string, 0)

			for _, entry := range value.([]interface{}) {
				if entry == "" || entry == nil {
					continue
				}

				valueString = append(valueString, fmt.Sprintf("%v", entry))
			}

			if len(valueString) == 0 {
				continue
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

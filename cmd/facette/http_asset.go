package main

import (
	"bytes"
	"html/template"
	"net/http"
)

const (
	httpDefaultPath = "html/index.html"
)

func (w *httpWorker) httpServeDefault(rw http.ResponseWriter, text string) {
	tmpl, err := template.New("index").Parse(text)
	if err != nil {
		w.service.log.Error("failed to parse template: %s", err)
		http.Error(rw, "", http.StatusInternalServerError)
		return
	}

	buf := bytes.NewBuffer(nil)

	data := struct{ RootPath string }{w.service.config.RootPath + "/"}

	err = tmpl.Execute(buf, data)
	if err != nil {
		w.service.log.Error("failed to execute template: %s", err)
		http.Error(rw, "", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "text/html")
	rw.WriteHeader(http.StatusOK)
	rw.Write(buf.Bytes())
}

package web

import (
	"bytes"
	"html/template"
	"net/http"
)

const (
	httpDefaultPath = "html/index.html"
)

func (h *Handler) serveDefault(rw http.ResponseWriter, text string) {
	tmpl, err := template.New("index").Parse(text)
	if err != nil {
		h.logger.Error("failed to parse template: %s", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	buf := bytes.NewBuffer(nil)

	data := struct{ BasePath string }{h.config.HTTP.BasePath + "/"}

	err = tmpl.Execute(buf, data)
	if err != nil {
		h.logger.Error("failed to execute template: %s", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "text/html")
	rw.WriteHeader(http.StatusOK)
	rw.Write(buf.Bytes())
}

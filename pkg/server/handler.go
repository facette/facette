package server

import (
	"bytes"
	"encoding/json"
	"html/template"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/facette/facette/thirdparty/github.com/fatih/set"
)

func (server *Server) serveError(writer http.ResponseWriter, status int) {
	var data struct {
		URLPrefix string
		Status    int
	}

	// Set template data
	data.URLPrefix = server.Config.URLPrefix
	data.Status = status

	// Execute template
	tmplData := bytes.NewBuffer(nil)

	tmpl, err := template.New("layout.html").Funcs(template.FuncMap{
		"asset": server.templateAsset,
		"eq":    templateEqual,
		"ne":    templateNotEqual,
	}).ParseFiles(
		path.Join(server.Config.BaseDir, "html", "layout.html"),
		path.Join(server.Config.BaseDir, "html", "error.html"),
	)
	if err == nil {
		err = tmpl.Execute(tmplData, data)
	}

	if err != nil {
		log.Println("ERROR: " + err.Error())
		server.serveResponse(writer, nil, status)
	}

	// Handle HTTP response with status code
	writer.WriteHeader(status)
	writer.Write(tmplData.Bytes())
}

func (server *Server) serveReload(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "GET" && request.Method != "HEAD" {
		server.serveResponse(writer, serverResponse{mesgMethodNotAllowed}, http.StatusMethodNotAllowed)
		return
	}

	server.Reload()

	server.serveResponse(writer, nil, http.StatusOK)
}

func (server *Server) serveResponse(writer http.ResponseWriter, data interface{}, status int) {
	var err error

	output := make([]byte, 0)

	if data != nil {
		output, err = json.Marshal(data)
		if err != nil {
			server.serveResponse(writer, nil, http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	}

	writer.WriteHeader(status)

	if len(output) > 0 {
		writer.Write(output)
		writer.Write([]byte("\n"))
	}
}

func (server *Server) serveStatic(writer http.ResponseWriter, request *http.Request) {
	mimeType := mime.TypeByExtension(filepath.Ext(request.URL.Path))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	writer.Header().Set("Content-Type", mimeType)

	// Handle static files
	http.ServeFile(writer, request, path.Join(server.Config.BaseDir, request.URL.Path))
}

func (server *Server) serveStats(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "GET" && request.Method != "HEAD" {
		server.serveResponse(writer, serverResponse{mesgMethodNotAllowed}, http.StatusMethodNotAllowed)
		return
	}

	server.serveResponse(writer, server.getStats(writer, request), http.StatusOK)
}

func (server *Server) serveWait(writer http.ResponseWriter, request *http.Request) {
	var data struct {
		URLPrefix string
	}

	// Set template data
	data.URLPrefix = server.Config.URLPrefix

	// Execute template
	tmpl, err := template.New("layout.html").Funcs(template.FuncMap{
		"asset": server.templateAsset,
	}).ParseFiles(
		path.Join(server.Config.BaseDir, "html", "layout.html"),
		path.Join(server.Config.BaseDir, "html", "wait.html"),
	)
	if err == nil {
		err = tmpl.Execute(writer, data)
	}

	if err != nil {
		if os.IsNotExist(err) {
			server.serveError(writer, http.StatusNotFound)
		} else {
			log.Println("ERROR: " + err.Error())
			server.serveError(writer, http.StatusInternalServerError)
		}
	}
}

func (server *Server) getStats(writer http.ResponseWriter, request *http.Request) *statsResponse {
	sourceSet := set.New(set.ThreadSafe)
	metricSet := set.New(set.ThreadSafe)

	for _, origin := range server.Catalog.Origins {
		for key, source := range origin.Sources {
			sourceSet.Add(key)

			for key := range source.Metrics {
				metricSet.Add(key)
			}
		}
	}

	return &statsResponse{
		Origins:        len(server.Catalog.Origins),
		Sources:        sourceSet.Size(),
		Metrics:        metricSet.Size(),
		CatalogUpdated: server.Catalog.Updated.Format(time.RFC3339),

		Graphs:      len(server.Library.Graphs),
		Collections: len(server.Library.Collections),
		Groups:      len(server.Library.Groups),
	}
}

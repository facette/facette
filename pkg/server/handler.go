package server

import (
	"bytes"
	"encoding/json"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/facette/facette/pkg/logger"
	"github.com/facette/facette/thirdparty/github.com/fatih/set"
)

func (server *Server) serveError(writer http.ResponseWriter, status int) {
	tmplData := bytes.NewBuffer(nil)

	err := server.execTemplate(
		writer,
		struct {
			URLPrefix string
			Status    int
		}{
			URLPrefix: server.Config.URLPrefix,
			Status:    status,
		},
		path.Join(server.Config.BaseDir, "template", "layout.html"),
		path.Join(server.Config.BaseDir, "template", "error.html"),
	)

	if err != nil {
		logger.Log(logger.LevelError, "server", "%s", err)
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

	// Reload resources without reloading configuration
	server.Reload(false)

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
	err := server.execTemplate(
		writer,
		struct {
			URLPrefix string
		}{
			URLPrefix: server.Config.URLPrefix,
		},
		path.Join(server.Config.BaseDir, "template", "layout.html"),
		path.Join(server.Config.BaseDir, "template", "wait.html"),
	)

	if err != nil {
		if os.IsNotExist(err) {
			server.serveError(writer, http.StatusNotFound)
		} else {
			logger.Log(logger.LevelError, "server", "%s", err)
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
		Origins:     len(server.Catalog.Origins),
		Sources:     sourceSet.Size(),
		Metrics:     metricSet.Size(),
		Graphs:      len(server.Library.Graphs),
		Collections: len(server.Library.Collections),
		Groups:      len(server.Library.Groups),
	}
}

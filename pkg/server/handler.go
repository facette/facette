package server

import (
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/facette/facette/pkg/logger"
)

func (server *Server) serveError(writer http.ResponseWriter, status int) {
	err := server.execTemplate(
		writer,
		status,
		struct {
			URLPrefix string
			ReadOnly  bool
			Status    int
		}{
			URLPrefix: server.Config.URLPrefix,
			ReadOnly:  server.Config.ReadOnly,
			Status:    status,
		},
		path.Join(server.Config.BaseDir, "template", "layout.html"),
		path.Join(server.Config.BaseDir, "template", "error.html"),
	)

	if err != nil {
		logger.Log(logger.LevelError, "server", "%s", err)
		server.serveResponse(writer, nil, status)
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

func (server *Server) serveWait(writer http.ResponseWriter, request *http.Request) {
	err := server.execTemplate(
		writer,
		http.StatusServiceUnavailable,
		struct {
			URLPrefix string
			ReadOnly  bool
		}{
			URLPrefix: server.Config.URLPrefix,
			ReadOnly:  server.Config.ReadOnly,
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

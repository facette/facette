package server

import (
	"mime"
	"net/http"
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

func init() {
	// Register default MIME types
	mime.AddExtensionType(".json", "application/json")
}

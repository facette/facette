package server

import (
	"net/http"
	"strings"

	"github.com/facette/facette/pkg/logger"
)

// ResponseWriter represents the structure of an HTTP response writer handling logged output.
type ResponseWriter struct {
	http.ResponseWriter
	request *http.Request
}

// WriteHeader sends an HTTP response header with along with its status code.
func (writer ResponseWriter) WriteHeader(status int) {
	writer.ResponseWriter.WriteHeader(status)

	logger.Log(logger.LevelDebug, "serveWorker", "\"%s %s %s\" %d", writer.request.Method, writer.request.URL,
		writer.request.Proto, status)
}

// Router represents the structure of an HTTP requests router.
type Router struct {
	*http.ServeMux
	server *Server
}

// NewRouter creates a new instance of router.
func NewRouter(server *Server) *Router {
	return &Router{
		ServeMux: http.NewServeMux(),
		server:   server,
	}
}

// ServerHTTP dispatches the requests to the router handlers.
func (router *Router) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if router.server.Config.URLPrefix != "" {
		request.URL.Path = strings.TrimPrefix(request.URL.Path, router.server.Config.URLPrefix)
	}

	router.ServeMux.ServeHTTP(ResponseWriter{writer, request}, request)
}

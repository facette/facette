package server

import (
	"net/http"
	"strings"
	"time"

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

// ServerHTTP dispatches the requests to the router handlers.
func (router *Router) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if router.server.Config.URLPrefix != "" {
		request.URL.Path = strings.TrimPrefix(request.URL.Path, router.server.Config.URLPrefix)
	}

	if router.server.loading {
		if strings.HasPrefix(request.URL.Path, urlAdminPath) || strings.HasPrefix(request.URL.Path, urlBrowsePath) {
			router.server.serveWait(writer, request)
			return
		} else if request.URL.Path == urlReloadPath {
			for {
				if !router.server.loading {
					break
				}

				time.Sleep(time.Second)
			}

			router.server.serveResponse(writer, nil, http.StatusOK)
			return
		} else if !strings.HasPrefix(request.URL.Path, urlStaticPath) {
			router.server.serveResponse(writer, serverResponse{mesgServiceLoading}, http.StatusServiceUnavailable)
			return
		}
	}

	router.ServeMux.ServeHTTP(ResponseWriter{writer, request}, request)
}

// NewRouter creates a new instance of router.
func NewRouter(server *Server) *Router {
	return &Router{
		ServeMux: http.NewServeMux(),
		server:   server,
	}
}

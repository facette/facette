package server

import (
	"log"
	"net/http"
	"strings"
	"time"
)

// ResponseWriter represents the structure of an HTTP response writer handling logged output.
type ResponseWriter struct {
	http.ResponseWriter
	request    *http.Request
	debugLevel int
}

// WriteHeader sends an HTTP response header with along with its status code.
func (writer ResponseWriter) WriteHeader(status int) {
	writer.ResponseWriter.WriteHeader(status)

	if writer.debugLevel > 2 {
		log.Printf("DEBUG: \"%s %s %s\" %d", writer.request.Method, writer.request.URL, writer.request.Proto, status)
	}
}

// Router represents the structure of an HTTP requests router.
type Router struct {
	*http.ServeMux
	server *Server
}

// ServerHTTP dispatches the requests to the router handlers.
func (router *Router) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if router.server.Loading {
		if strings.HasPrefix(request.URL.Path, urlAdminPath) || strings.HasPrefix(request.URL.Path, urlBrowsePath) {
			router.server.handleWait(writer, request)
			return
		} else if request.URL.Path == urlReloadPath {
			for {
				if !router.server.Loading {
					break
				}

				time.Sleep(time.Second)
			}

			router.server.handleResponse(writer, nil, http.StatusOK)
			return
		} else if !strings.HasPrefix(request.URL.Path, urlStaticPath) {
			router.server.handleResponse(writer, serverResponse{mesgServiceLoading}, http.StatusServiceUnavailable)
			return
		}
	}

	router.ServeMux.ServeHTTP(ResponseWriter{writer, request, router.server.debugLevel}, request)
}

// NewRouter creates a new instance of router.
func NewRouter(server *Server) *Router {
	return &Router{
		ServeMux: http.NewServeMux(),
		server:   server,
	}
}

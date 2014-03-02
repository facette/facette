package server

import (
	"log"
	"net/http"
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
	debugLevel int
}

// ServerHTTP dispatches the requests to the router handlers.
func (router *Router) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	router.ServeMux.ServeHTTP(ResponseWriter{writer, request, router.debugLevel}, request)
}

// NewRouter creates a new instance of router.
func NewRouter(debugLevel int) *Router {
	return &Router{
		ServeMux:   http.NewServeMux(),
		debugLevel: debugLevel,
	}
}

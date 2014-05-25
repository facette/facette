package server

import (
	"net/http"
	"strings"
)

func (server *Server) serveLibrary(writer http.ResponseWriter, request *http.Request) {
	setHTTPCacheHeaders(writer)

	if strings.HasPrefix(request.URL.Path, urlLibraryPath+"sourcegroups/") {
		server.serveGroup(writer, request)
	} else if strings.HasPrefix(request.URL.Path, urlLibraryPath+"metricgroups/") {
		server.serveGroup(writer, request)
	} else if strings.HasPrefix(request.URL.Path, urlLibraryPath+"scales/") {
		server.serveScale(writer, request)
	} else if request.URL.Path == urlLibraryPath+"expand" {
		server.serveGroupExpand(writer, request)
	} else if request.URL.Path == urlLibraryPath+"graphs/plots" {
		server.serveGraphPlots(writer, request)
	} else if strings.HasPrefix(request.URL.Path, urlLibraryPath+"graphs/") {
		server.serveGraph(writer, request)
	} else if strings.HasPrefix(request.URL.Path, urlLibraryPath+"collections/") {
		server.serveCollection(writer, request)
	} else {
		server.serveResponse(writer, nil, http.StatusNotFound)
	}
}

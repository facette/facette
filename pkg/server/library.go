package server

import (
	"net/http"
	"strings"
)

func (server *Server) handleLibrary(writer http.ResponseWriter, request *http.Request) {
	setHTTPCacheHeaders(writer)

	if strings.HasPrefix(request.URL.Path, urlLibraryPath+"sourcegroups/") {
		server.handleGroup(writer, request)
	} else if strings.HasPrefix(request.URL.Path, urlLibraryPath+"metricgroups/") {
		server.handleGroup(writer, request)
	} else if request.URL.Path == urlLibraryPath+"expand" {
		server.handleGroupExpand(writer, request)
	} else if request.URL.Path == urlLibraryPath+"graphs/plots" {
		server.handleGraphPlots(writer, request)
	} else if strings.HasPrefix(request.URL.Path, urlLibraryPath+"graphs/") {
		server.handleGraph(writer, request)
	} else if strings.HasPrefix(request.URL.Path, urlLibraryPath+"collections/") {
		server.handleCollection(writer, request)
	} else {
		server.handleResponse(writer, nil, http.StatusNotFound)
	}
}

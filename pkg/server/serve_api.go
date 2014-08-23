package server

import (
	"encoding/json"
	"net/http"

	"github.com/facette/facette/pkg/library"
	"github.com/facette/facette/thirdparty/github.com/fatih/set"
)

func (server *Server) serveResponse(writer http.ResponseWriter, data interface{}, status int) {
	var (
		err    error
		output []byte
	)

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

func (server *Server) serveStats(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "GET" && request.Method != "HEAD" {
		server.serveResponse(writer, serverResponse{mesgMethodNotAllowed}, http.StatusMethodNotAllowed)
		return
	}

	server.serveResponse(writer, server.getStats(writer, request), http.StatusOK)
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

	sourceGroupsCount := 0
	metricGroupsCount := 0

	for _, group := range server.Library.Groups {
		if group.Type == library.LibraryItemSourceGroup {
			sourceGroupsCount++
		} else {
			metricGroupsCount++
		}
	}

	return &statsResponse{
		Origins:      len(server.Catalog.Origins),
		Sources:      sourceSet.Size(),
		Metrics:      metricSet.Size(),
		Graphs:       len(server.Library.Graphs),
		Collections:  len(server.Library.Collections),
		SourceGroups: sourceGroupsCount,
		MetricGroups: metricGroupsCount,
	}
}

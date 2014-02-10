package server

import (
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/facette/facette/pkg/library"
	"github.com/facette/facette/pkg/types"
	"github.com/facette/facette/pkg/utils"
	"github.com/facette/facette/thirdparty/github.com/fatih/set"
	"github.com/facette/facette/thirdparty/github.com/gorilla/mux"
)

func (server *Server) metricList(writer http.ResponseWriter, request *http.Request) {
	var (
		err    error
		limit  int
		offset int
	)

	if request.Method != "GET" && request.Method != "HEAD" {
		server.handleResponse(writer, http.StatusMethodNotAllowed)
		return
	}

	if request.FormValue("offset") != "" {
		if offset, err = strconv.Atoi(request.FormValue("offset")); err != nil {
			server.handleResponse(writer, http.StatusBadRequest)
			return
		}
	}

	if request.FormValue("limit") != "" {
		if limit, err = strconv.Atoi(request.FormValue("limit")); err != nil {
			server.handleResponse(writer, http.StatusBadRequest)
			return
		}
	}

	// Parse catalog for metrics list
	originName := request.FormValue("origin")

	sourceSet := set.New()
	responseSet := set.New()

	if strings.HasPrefix(request.FormValue("source"), "group:") {
		for _, sourceName := range server.Library.ExpandGroup(request.FormValue("source")[6:],
			library.LibraryItemSourceGroup) {
			sourceSet.Add(sourceName)
		}
	} else if request.FormValue("source") != "" {
		sourceSet.Add(request.FormValue("source"))
	}

	for _, origin := range server.Catalog.Origins {
		if originName != "" && origin.Name != originName {
			continue
		}

		for _, source := range origin.Sources {
			if request.FormValue("source") != "" && sourceSet.IsEmpty() ||
				!sourceSet.IsEmpty() && !sourceSet.Has(source.Name) {
				continue
			}

			for key := range source.Metrics {
				if request.FormValue("filter") != "" {
					if !utils.FilterMatch(strings.ToLower(request.FormValue("filter")), strings.ToLower(key)) {
						continue
					}
				}

				responseSet.Add(key)
			}
		}
	}

	if offset != 0 && offset >= responseSet.Size() {
		server.handleResponse(writer, http.StatusBadRequest)
		return
	}

	writer.Header().Add("X-Total-Records", strconv.Itoa(responseSet.Size()))

	response := responseSet.StringSlice()
	sort.Strings(response)

	// Shrink responses if limit is set
	if limit != 0 && len(response) > offset+limit {
		response = response[offset : offset+limit]
	} else if offset != 0 {
		response = response[offset:]
	}

	server.handleJSON(writer, response)
}

func (server *Server) metricShow(writer http.ResponseWriter, request *http.Request) {
	found := false

	if request.Method != "GET" && request.Method != "HEAD" {
		server.handleResponse(writer, http.StatusMethodNotAllowed)
		return
	}

	// Parse catalog for metric information
	metricName := mux.Vars(request)["path"]

	originSet := set.New()
	sourceSet := set.New()

	for _, origin := range server.Catalog.Origins {
		for _, source := range origin.Sources {
			if _, ok := source.Metrics[metricName]; ok {
				originSet.Add(origin.Name)
				sourceSet.Add(source.Name)
				found = true
			}
		}
	}

	if !found {
		server.handleResponse(writer, http.StatusNotFound)
		return
	}

	response := types.MetricResponse{
		Name:    metricName,
		Origins: originSet.StringSlice(),
		Sources: sourceSet.StringSlice(),
		Updated: server.Catalog.Updated.Format(time.RFC3339),
	}

	server.handleJSON(writer, response)
}

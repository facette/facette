package server

import (
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/facette/facette/pkg/types"
	"github.com/facette/facette/pkg/utils"
	"github.com/facette/facette/thirdparty/github.com/fatih/set"
	"github.com/facette/facette/thirdparty/github.com/gorilla/mux"
)

func (server *Server) sourceList(writer http.ResponseWriter, request *http.Request) {
	var (
		err         error
		limit       int
		offset      int
		originName  string
		response    []string
		responseSet *set.Set
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

	// Parse catalog for sources list
	originName = request.FormValue("origin")

	responseSet = set.New()

	for _, origin := range server.Catalog.Origins {
		if originName != "" && origin.Name != originName {
			continue
		}

		for key := range origin.Sources {
			if request.FormValue("filter") != "" {
				if !utils.FilterMatch(strings.ToLower(request.FormValue("filter")), strings.ToLower(key)) {
					continue
				}
			}

			responseSet.Add(key)
		}
	}

	if offset != 0 && offset >= responseSet.Size() {
		server.handleResponse(writer, http.StatusBadRequest)
		return
	}

	writer.Header().Add("X-Total-Records", strconv.Itoa(responseSet.Size()))

	response = responseSet.StringSlice()
	sort.Strings(response)

	// Shrink responses if limit is set
	if limit != 0 && len(response) > offset+limit {
		response = response[offset : offset+limit]
	} else if offset != 0 {
		response = response[offset:]
	}

	server.handleJSON(writer, response)
}

func (server *Server) sourceShow(writer http.ResponseWriter, request *http.Request) {
	var (
		found      bool
		sourceName string
		response   types.SourceResponse
	)

	if request.Method != "GET" && request.Method != "HEAD" {
		server.handleResponse(writer, http.StatusMethodNotAllowed)
		return
	}

	// Parse catalog for source information
	sourceName = mux.Vars(request)["name"]

	for _, origin := range server.Catalog.Origins {
		if _, ok := origin.Sources[sourceName]; ok {
			response.Origins = append(response.Origins, origin.Name)
			found = true
		}
	}

	if !found {
		server.handleResponse(writer, http.StatusNotFound)
		return
	}

	response.Name = sourceName
	response.Updated = server.Catalog.Updated.Format(time.RFC3339)

	server.handleJSON(writer, response)
}

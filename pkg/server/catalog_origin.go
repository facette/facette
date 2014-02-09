package server

import (
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/facette/facette/pkg/utils"
	"github.com/facette/facette/thirdparty/github.com/fatih/set"
	"github.com/facette/facette/thirdparty/github.com/gorilla/mux"
)

type originShowResponse struct {
	Name    string `json:"name"`
	Backend string `json:"backend"`
	Updated string `json:"updated"`
}

func (server *Server) originList(writer http.ResponseWriter, request *http.Request) {
	var (
		err      error
		limit    int
		offset   int
		response []string
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
	for _, origin := range server.Catalog.Origins {
		if request.FormValue("filter") != "" {
			if !utils.FilterMatch(strings.ToLower(request.FormValue("filter")), strings.ToLower(origin.Name)) {
				continue
			}
		}

		response = append(response, origin.Name)
	}

	if offset != 0 && offset >= len(response) {
		server.handleResponse(writer, http.StatusBadRequest)
		return
	}

	writer.Header().Add("X-Total-Records", strconv.Itoa(len(response)))

	sort.Strings(response)

	// Shrink responses if limit is set
	if limit != 0 && len(response) > offset+limit {
		response = response[offset : offset+limit]
	} else if offset != 0 {
		response = response[offset:]
	}

	server.handleJSON(writer, response)
}

func (server *Server) originShow(writer http.ResponseWriter, request *http.Request) {
	var (
		metrics    *set.Set
		originName string
		response   originShowResponse
	)

	if request.Method != "GET" && request.Method != "HEAD" {
		server.handleResponse(writer, http.StatusMethodNotAllowed)
		return
	}

	// Parse catalog for source information
	originName = mux.Vars(request)["name"]

	if _, ok := server.Catalog.Origins[originName]; !ok {
		server.handleResponse(writer, http.StatusNotFound)
		return
	}

	metrics = set.New()

	for _, source := range server.Catalog.Origins[originName].Sources {
		for _, metric := range source.Metrics {
			metrics.Add(metric.Name)
		}
	}

	response.Name = originName
	response.Backend = server.Config.Origins[originName].Backend["type"]
	response.Updated = server.Catalog.Updated.Format(time.RFC3339)

	server.handleJSON(writer, response)
}

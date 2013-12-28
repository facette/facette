package server

import (
	"encoding/json"
	"facette/library"
	"github.com/fatih/set"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"
)

type expandRequest [][3]string

func (tuple expandRequest) Len() int {
	return len(tuple)
}

func (tuple expandRequest) Less(i, j int) bool {
	return tuple[i][0]+tuple[i][1]+tuple[i][2] < tuple[j][0]+tuple[j][1]+tuple[j][2]
}

func (tuple expandRequest) Swap(i, j int) {
	tuple[i], tuple[j] = tuple[j], tuple[i]
}

type originShowResponse struct {
	Name    string `json:"name"`
	Backend string `json:"backend"`
	Updated string `json:"updated"`
}

type sourceShowResponse struct {
	Name    string   `json:"name"`
	Origins []string `json:"origins"`
	Updated string   `json:"updated"`
}

type metricShowResponse struct {
	Name    string   `json:"name"`
	Origins []string `json:"origins"`
	Sources []string `json:"sources"`
	Updated string   `json:"updated"`
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
			if ok, _ := path.Match(strings.ToLower(request.FormValue("filter")), strings.ToLower(origin.Name)); !ok {
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
				if ok, _ := path.Match(strings.ToLower(request.FormValue("filter")), strings.ToLower(key)); !ok {
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
		response   sourceShowResponse
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

func (server *Server) metricList(writer http.ResponseWriter, request *http.Request) {
	var (
		err         error
		limit       int
		offset      int
		originName  string
		response    []string
		responseSet *set.Set
		sourceSet   *set.Set
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
	originName = request.FormValue("origin")

	sourceSet = set.New()
	responseSet = set.New()

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
					if ok, _ := path.Match(strings.ToLower(request.FormValue("filter")), strings.ToLower(key)); !ok {
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

func (server *Server) metricShow(writer http.ResponseWriter, request *http.Request) {
	var (
		found      bool
		metricName string
		originSet  *set.Set
		response   metricShowResponse
		sourceSet  *set.Set
	)

	if request.Method != "GET" && request.Method != "HEAD" {
		server.handleResponse(writer, http.StatusMethodNotAllowed)
		return
	}

	// Parse catalog for metric information
	metricName = mux.Vars(request)["path"]

	originSet = set.New()
	sourceSet = set.New()

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

	response = metricShowResponse{
		Name:    metricName,
		Origins: originSet.StringSlice(),
		Sources: sourceSet.StringSlice(),
		Updated: server.Catalog.Updated.Format(time.RFC3339),
	}

	server.handleJSON(writer, response)
}

func (server *Server) expandList(writer http.ResponseWriter, request *http.Request) {
	var (
		body     []byte
		err      error
		item     expandRequest
		query    expandRequest
		response []expandRequest
	)

	if request.Method != "POST" {
		server.handleResponse(writer, http.StatusMethodNotAllowed)
		return
	}

	body, _ = ioutil.ReadAll(request.Body)

	if err = json.Unmarshal(body, &query); err != nil {
		log.Println("ERROR: " + err.Error())
		server.handleResponse(writer, http.StatusBadRequest)
		return
	}

	for _, entry := range query {
		item = expandRequest{}

		if strings.HasPrefix(entry[1], "group:") {
			for _, sourceName := range server.Library.ExpandGroup(entry[1][6:], library.LibraryItemSourceGroup) {
				if strings.HasPrefix(entry[2], "group:") {
					for _, metricName := range server.Library.ExpandGroup(entry[2][6:],
						library.LibraryItemMetricGroup) {
						item = append(item, [3]string{entry[0], sourceName, metricName})
					}
				} else {
					item = append(item, [3]string{entry[0], sourceName, entry[2]})
				}
			}
		} else if strings.HasPrefix(entry[2], "group:") {
			for _, metricName := range server.Library.ExpandGroup(entry[2][6:], library.LibraryItemMetricGroup) {
				item = append(item, [3]string{entry[0], entry[1], metricName})
			}
		} else {
			item = append(item, entry)
		}

		sort.Sort(item)
		response = append(response, item)
	}

	server.handleJSON(writer, response)
}

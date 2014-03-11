package server

import (
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/facette/facette/pkg/library"
	"github.com/facette/facette/pkg/utils"
	"github.com/facette/facette/thirdparty/github.com/fatih/set"
)

func (server *Server) handleCatalog(writer http.ResponseWriter, request *http.Request) {
	if strings.HasPrefix(request.URL.Path, urlCatalogPath+"origins/") {
		server.handleOrigin(writer, request)
	} else if strings.HasPrefix(request.URL.Path, urlCatalogPath+"sources/") {
		server.handleSource(writer, request)
	} else if strings.HasPrefix(request.URL.Path, urlCatalogPath+"metrics/") {
		server.handleMetric(writer, request)
	} else {
		server.handleResponse(writer, nil, http.StatusNotFound)
	}
}

func (server *Server) handleOrigin(writer http.ResponseWriter, request *http.Request) {
	originName := strings.TrimPrefix(request.URL.Path, urlCatalogPath+"origins/")

	if originName == "" {
		server.handleOriginList(writer, request)
		return
	}

	if response, status := server.parseShowRequest(writer, request); status != http.StatusOK {
		server.handleResponse(writer, response, status)
		return
	} else if _, ok := server.Catalog.Origins[originName]; !ok {
		server.handleResponse(writer, serverResponse{mesgResourceNotFound}, http.StatusNotFound)
		return
	}

	response := OriginResponse{
		Name:      originName,
		Connector: server.Config.Origins[originName].Connector["type"],
		Updated:   server.Catalog.Updated.Format(time.RFC3339),
	}

	server.handleResponse(writer, response, http.StatusOK)
}

func (server *Server) handleOriginList(writer http.ResponseWriter, request *http.Request) {
	var offset, limit int

	if response, status := server.parseListRequest(writer, request, &offset, &limit); status != http.StatusOK {
		server.handleResponse(writer, response, status)
		return
	}

	originSet := set.New()

	for _, origin := range server.Catalog.Origins {
		if request.FormValue("filter") != "" && !utils.FilterMatch(request.FormValue("filter"), origin.Name) {
			continue
		}

		originSet.Add(origin.Name)
	}

	response := &listResponse{
		list:   StringListResponse(originSet.StringSlice()),
		offset: offset,
		limit:  limit,
	}

	server.applyResponseLimit(writer, request, response)

	server.handleResponse(writer, response.list, http.StatusOK)
}

func (server *Server) handleSource(writer http.ResponseWriter, request *http.Request) {
	sourceName := strings.TrimPrefix(request.URL.Path, urlCatalogPath+"sources/")

	if sourceName == "" {
		server.handleSourceList(writer, request)
		return
	} else if response, status := server.parseShowRequest(writer, request); status != http.StatusOK {
		server.handleResponse(writer, response, status)
		return
	}

	originSet := set.New()

	for _, origin := range server.Catalog.Origins {
		if _, ok := origin.Sources[sourceName]; ok {
			originSet.Add(origin.Name)
		}
	}

	if originSet.Size() == 0 {
		server.handleResponse(writer, serverResponse{mesgResourceNotFound}, http.StatusNotFound)
		return
	}

	origins := originSet.StringSlice()
	sort.Strings(origins)

	response := SourceResponse{
		Name:    sourceName,
		Origins: origins,
		Updated: server.Catalog.Updated.Format(time.RFC3339),
	}

	server.handleResponse(writer, response, http.StatusOK)
}

func (server *Server) handleSourceList(writer http.ResponseWriter, request *http.Request) {
	var offset, limit int

	if response, status := server.parseListRequest(writer, request, &offset, &limit); status != http.StatusOK {
		server.handleResponse(writer, response, status)
		return
	}

	originName := request.FormValue("origin")

	sourceSet := set.New()

	for _, origin := range server.Catalog.Origins {
		if originName != "" && origin.Name != originName {
			continue
		}

		for key := range origin.Sources {
			if request.FormValue("filter") != "" && !utils.FilterMatch(request.FormValue("filter"), key) {
				continue
			}

			sourceSet.Add(key)
		}
	}

	response := &listResponse{
		list:   StringListResponse(sourceSet.StringSlice()),
		offset: offset,
		limit:  limit,
	}

	server.applyResponseLimit(writer, request, response)

	server.handleResponse(writer, response.list, http.StatusOK)
}

func (server *Server) handleMetric(writer http.ResponseWriter, request *http.Request) {
	metricName := strings.TrimPrefix(request.URL.Path, urlCatalogPath+"metrics/")

	if metricName == "" {
		server.handleMetricList(writer, request)
		return
	} else if response, status := server.parseShowRequest(writer, request); status != http.StatusOK {
		server.handleResponse(writer, response, status)
		return
	}

	originSet := set.New()
	sourceSet := set.New()

	for _, origin := range server.Catalog.Origins {
		for _, source := range origin.Sources {
			if _, ok := source.Metrics[metricName]; ok {
				originSet.Add(origin.Name)
				sourceSet.Add(source.Name)
			}
		}
	}

	if originSet.Size() == 0 {
		server.handleResponse(writer, serverResponse{mesgResourceNotFound}, http.StatusNotFound)
		return
	}

	origins := originSet.StringSlice()
	sort.Strings(origins)

	sources := sourceSet.StringSlice()
	sort.Strings(sources)

	response := MetricResponse{
		Name:    metricName,
		Origins: origins,
		Sources: sources,
		Updated: server.Catalog.Updated.Format(time.RFC3339),
	}

	server.handleResponse(writer, response, http.StatusOK)
}

func (server *Server) handleMetricList(writer http.ResponseWriter, request *http.Request) {
	var offset, limit int

	if response, status := server.parseListRequest(writer, request, &offset, &limit); status != http.StatusOK {
		server.handleResponse(writer, response, status)
		return
	}

	originName := request.FormValue("origin")
	sourceName := request.FormValue("source")

	sourceSet := set.New()

	if strings.HasPrefix(sourceName, library.LibraryGroupPrefix) {
		for _, entryName := range server.Library.ExpandGroup(
			strings.TrimPrefix(sourceName, library.LibraryGroupPrefix),
			library.LibraryItemSourceGroup,
		) {
			sourceSet.Add(entryName)
		}
	} else if sourceName != "" {
		sourceSet.Add(sourceName)
	}

	metricSet := set.New()

	for _, origin := range server.Catalog.Origins {
		if originName != "" && origin.Name != originName {
			continue
		}

		for _, source := range origin.Sources {
			if sourceName != "" && sourceSet.IsEmpty() || !sourceSet.IsEmpty() && !sourceSet.Has(source.Name) {
				continue
			}

			for key := range source.Metrics {
				if request.FormValue("filter") != "" && !utils.FilterMatch(request.FormValue("filter"), key) {
					continue
				}

				metricSet.Add(key)
			}
		}
	}

	response := &listResponse{
		list:   StringListResponse(metricSet.StringSlice()),
		offset: offset,
		limit:  limit,
	}

	server.applyResponseLimit(writer, request, response)

	server.handleResponse(writer, response.list, http.StatusOK)
}

package server

import (
	"net/http"
	"sort"
	"strings"

	"github.com/facette/facette/pkg/library"
	"github.com/facette/facette/pkg/utils"
	"github.com/fatih/set"
)

func (server *Server) serveCatalog(writer http.ResponseWriter, request *http.Request) {
	setHTTPCacheHeaders(writer)

	// Dispatch library API routes
	if routeMatch(request.URL.Path, urlCatalogPath) {
		server.serveFullCatalog(writer, request)
	} else if routeMatch(request.URL.Path, urlCatalogPath+"origins") {
		server.serveOrigin(writer, request)
	} else if routeMatch(request.URL.Path, urlCatalogPath+"sources") {
		server.serveSource(writer, request)
	} else if routeMatch(request.URL.Path, urlCatalogPath+"metrics") {
		server.serveMetric(writer, request)
	} else {
		server.serveResponse(writer, nil, http.StatusNotFound)
	}
}

func (server *Server) serveFullCatalog(writer http.ResponseWriter, request *http.Request) {
	catalog := make(map[string]map[string][]string)

	for _, origin := range server.Catalog.GetOrigins() {
		catalog[origin.Name] = make(map[string][]string)

		for _, source := range origin.GetSources() {
			catalog[origin.Name][source.Name] = make([]string, 0)

			for _, metric := range source.GetMetrics() {
				catalog[origin.Name][source.Name] = append(catalog[origin.Name][source.Name], metric.Name)
			}

			sort.Strings(catalog[origin.Name][source.Name])
		}
	}

	server.serveResponse(writer, catalog, http.StatusOK)
}

func (server *Server) serveOrigin(writer http.ResponseWriter, request *http.Request) {
	name := routeTrimPrefix(request.URL.Path, urlCatalogPath+"origins")

	if name == "" {
		server.serveOriginList(writer, request)
		return
	}

	if response, status := server.parseShowRequest(writer, request); status != http.StatusOK {
		server.serveResponse(writer, response, status)
		return
	} else if !server.Catalog.OriginExists(name) {
		server.serveResponse(writer, serverResponse{mesgResourceNotFound}, http.StatusNotFound)
		return
	}

	connectorType, ok := server.Config.Providers[name].Connector["type"].(string)
	if !ok {
		server.serveResponse(writer, serverResponse{mesgUnhandledError}, http.StatusInternalServerError)
		return
	}

	response := OriginResponse{
		Name:      name,
		Connector: connectorType,
	}

	server.serveResponse(writer, response, http.StatusOK)
}

func (server *Server) serveOriginList(writer http.ResponseWriter, request *http.Request) {
	var offset, limit int

	if response, status := server.parseListRequest(writer, request, &offset, &limit); status != http.StatusOK {
		server.serveResponse(writer, response, status)
		return
	}

	originSet := set.New(set.ThreadSafe)
	for _, origin := range server.Catalog.GetOrigins() {
		if request.FormValue("filter") == "" || utils.FilterMatch(request.FormValue("filter"), origin.Name) {
			originSet.Add(origin.Name)
		}
	}

	response := &listResponse{
		list:   StringListResponse(set.StringSlice(originSet)),
		offset: offset,
		limit:  limit,
	}

	server.applyResponseLimit(writer, request, response)

	server.serveResponse(writer, response.list, http.StatusOK)
}

func (server *Server) serveSource(writer http.ResponseWriter, request *http.Request) {
	name := routeTrimPrefix(request.URL.Path, urlCatalogPath+"sources")

	if name == "" {
		server.serveSourceList(writer, request)
		return
	} else if response, status := server.parseShowRequest(writer, request); status != http.StatusOK {
		server.serveResponse(writer, response, status)
		return
	}

	originSet := set.New(set.ThreadSafe)
	for _, origin := range server.Catalog.GetOrigins() {
		if origin.SourceExists(name) {
			originSet.Add(origin.Name)
		}
	}

	if originSet.Size() == 0 {
		server.serveResponse(writer, serverResponse{mesgResourceNotFound}, http.StatusNotFound)
		return
	}

	origins := set.StringSlice(originSet)
	sort.Strings(origins)

	response := SourceResponse{
		Name:    name,
		Origins: origins,
	}

	server.serveResponse(writer, response, http.StatusOK)
}

func (server *Server) serveSourceList(writer http.ResponseWriter, request *http.Request) {
	var offset, limit int

	if response, status := server.parseListRequest(writer, request, &offset, &limit); status != http.StatusOK {
		server.serveResponse(writer, response, status)
		return
	}

	originName := request.FormValue("origin")

	sourceSet := set.New(set.ThreadSafe)

	for _, origin := range server.Catalog.GetOrigins() {
		if originName != "" && origin.Name != originName {
			continue
		}

		for _, source := range origin.GetSources() {
			if request.FormValue("filter") != "" && !utils.FilterMatch(request.FormValue("filter"), source.Name) {
				continue
			}

			sourceSet.Add(source.Name)
		}
	}

	response := &listResponse{
		list:   StringListResponse(set.StringSlice(sourceSet)),
		offset: offset,
		limit:  limit,
	}

	server.applyResponseLimit(writer, request, response)

	server.serveResponse(writer, response.list, http.StatusOK)
}

func (server *Server) serveMetric(writer http.ResponseWriter, request *http.Request) {
	metricName := routeTrimPrefix(request.URL.Path, urlCatalogPath+"metrics")

	if metricName == "" {
		server.serveMetricList(writer, request)
		return
	} else if response, status := server.parseShowRequest(writer, request); status != http.StatusOK {
		server.serveResponse(writer, response, status)
		return
	}

	originSet := set.New(set.ThreadSafe)
	sourceSet := set.New(set.ThreadSafe)

	for _, origin := range server.Catalog.GetOrigins() {
		for _, source := range origin.GetSources() {
			if source.MetricExists(metricName) {
				originSet.Add(origin.Name)
				sourceSet.Add(source.Name)
			}
		}
	}

	if originSet.Size() == 0 {
		server.serveResponse(writer, serverResponse{mesgResourceNotFound}, http.StatusNotFound)
		return
	}

	origins := set.StringSlice(originSet)
	sort.Strings(origins)

	sources := set.StringSlice(sourceSet)
	sort.Strings(sources)

	response := MetricResponse{
		Name:    metricName,
		Origins: origins,
		Sources: sources,
	}

	server.serveResponse(writer, response, http.StatusOK)
}

func (server *Server) serveMetricList(writer http.ResponseWriter, request *http.Request) {
	var offset, limit int

	if response, status := server.parseListRequest(writer, request, &offset, &limit); status != http.StatusOK {
		server.serveResponse(writer, response, status)
		return
	}

	originName := request.FormValue("origin")
	sourceName := request.FormValue("source")

	sourceSet := set.New(set.ThreadSafe)

	if strings.HasPrefix(sourceName, library.LibraryGroupPrefix) {
		for _, entryName := range server.Library.ExpandSourceGroup(
			strings.TrimPrefix(sourceName, library.LibraryGroupPrefix),
		) {
			sourceSet.Add(entryName)
		}
	} else if sourceName != "" {
		sourceSet.Add(sourceName)
	}

	metricSet := set.New(set.ThreadSafe)

	for _, origin := range server.Catalog.GetOrigins() {
		if originName != "" && origin.Name != originName {
			continue
		}

		for _, source := range origin.GetSources() {
			if sourceName != "" && sourceSet.IsEmpty() || !sourceSet.IsEmpty() && !sourceSet.Has(source.Name) {
				continue
			}

			for _, metric := range source.GetMetrics() {
				if request.FormValue("filter") != "" && !utils.FilterMatch(request.FormValue("filter"), metric.Name) {
					continue
				}

				metricSet.Add(metric.Name)
			}
		}
	}

	response := &listResponse{
		list:   StringListResponse(set.StringSlice(metricSet)),
		offset: offset,
		limit:  limit,
	}

	server.applyResponseLimit(writer, request, response)

	server.serveResponse(writer, response.list, http.StatusOK)
}

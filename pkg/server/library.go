package server

import (
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/facette/facette/pkg/library"
	"github.com/facette/facette/pkg/types"
	"github.com/facette/facette/pkg/utils"
	"github.com/facette/facette/thirdparty/github.com/fatih/set"
)

func (server *Server) libraryList(writer http.ResponseWriter, request *http.Request) {
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

	response := types.ItemListResponse{}

	if request.URL.Path == URLLibraryPath+"/sourcegroups" || request.URL.Path == URLLibraryPath+"/metricgroups" {
		isSource := request.URL.Path == URLLibraryPath+"/sourcegroups"

		// Get and filter source groups list
		for _, group := range server.Library.Groups {
			if isSource && group.Type != library.LibraryItemSourceGroup ||
				!isSource && group.Type != library.LibraryItemMetricGroup {
				continue
			}

			if request.FormValue("filter") != "" {
				if !utils.FilterMatch(strings.ToLower(request.FormValue("filter")), strings.ToLower(group.Name)) {
					continue
				}
			}

			response.Items = append(response.Items, &types.ItemResponse{
				ID:          group.ID,
				Name:        group.Name,
				Description: group.Description,
				Modified:    group.Modified.Format(time.RFC3339),
			})
		}
	} else if request.URL.Path == URLLibraryPath+"/graphs" {
		skip := false

		graphSet := set.New()

		// Filter by collection
		if request.FormValue("collection") != "" {
			item, err := server.Library.GetItem(request.FormValue("collection"), library.LibraryItemCollection)
			if os.IsNotExist(err) {
				skip = true
			} else if err != nil {
				server.handleResponse(writer, http.StatusInternalServerError)
				return
			}

			collection := item.(*library.Collection)

			for _, graph := range collection.Entries {
				graphSet.Add(graph.ID)
			}
		}

		// Get and filter graphs list
		if !skip {
			for _, graph := range server.Library.Graphs {
				if graph.Volatile || !graphSet.IsEmpty() && !graphSet.Has(graph.ID) {
					continue
				}

				if request.FormValue("filter") != "" {
					if !utils.FilterMatch(strings.ToLower(request.FormValue("filter")), strings.ToLower(graph.Name)) {
						continue
					}
				}

				response.Items = append(response.Items, &types.ItemResponse{
					ID:          graph.ID,
					Name:        graph.Name,
					Description: graph.Description,
					Modified:    graph.Modified.Format(time.RFC3339),
				})
			}
		}
	}

	if offset != 0 && offset >= len(response.Items) {
		server.handleResponse(writer, http.StatusBadRequest)
		return
	}

	writer.Header().Add("X-Total-Records", strconv.Itoa(len(response.Items)))

	sort.Sort(response)

	// Shrink responses if limit is set
	if limit != 0 && len(response.Items) > offset+limit {
		response.Items = response.Items[offset : offset+limit]
	} else if offset != 0 {
		response.Items = response.Items[offset:]
	}

	server.handleJSON(writer, response.Items)
}

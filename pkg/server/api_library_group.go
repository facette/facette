package server

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/facette/facette/pkg/library"
	"github.com/facette/facette/pkg/logger"
	"github.com/facette/facette/pkg/utils"
)

func (server *Server) serveGroup(writer http.ResponseWriter, request *http.Request) {
	var (
		groupID   string
		groupType int
	)

	if request.Method != "GET" && request.Method != "HEAD" && server.Config.ReadOnly {
		server.serveResponse(writer, serverResponse{mesgReadOnlyMode}, http.StatusForbidden)
		return
	}

	if routeMatch(request.URL.Path, urlLibraryPath+"sourcegroups") {
		groupID = routeTrimPrefix(request.URL.Path, urlLibraryPath+"sourcegroups")
		groupType = library.LibraryItemSourceGroup
	} else if routeMatch(request.URL.Path, urlLibraryPath+"metricgroups") {
		groupID = routeTrimPrefix(request.URL.Path, urlLibraryPath+"metricgroups")
		groupType = library.LibraryItemMetricGroup
	}

	switch request.Method {
	case "DELETE":
		if groupID == "" {
			server.serveResponse(writer, serverResponse{mesgMethodNotAllowed}, http.StatusMethodNotAllowed)
			return
		}

		err := server.Library.DeleteItem(groupID, groupType)
		if os.IsNotExist(err) {
			server.serveResponse(writer, serverResponse{mesgResourceNotFound}, http.StatusNotFound)
			return
		} else if err != nil {
			logger.Log(logger.LevelError, "server", "%s", err)
			server.serveResponse(writer, serverResponse{mesgUnhandledError}, http.StatusInternalServerError)
			return
		}

		server.serveResponse(writer, nil, http.StatusOK)

	case "GET", "HEAD":
		if groupID == "" {
			server.serveGroupList(writer, request)
			return
		}

		item, err := server.Library.GetItem(groupID, groupType)
		if os.IsNotExist(err) {
			server.serveResponse(writer, serverResponse{mesgResourceNotFound}, http.StatusNotFound)
			return
		} else if err != nil {
			logger.Log(logger.LevelError, "server", "%s", err)
			server.serveResponse(writer, serverResponse{mesgUnhandledError}, http.StatusInternalServerError)
			return
		}

		server.serveResponse(writer, item, http.StatusOK)

	case "POST", "PUT":
		var group *library.Group

		if response, status := server.parseStoreRequest(writer, request, groupID); status != http.StatusOK {
			server.serveResponse(writer, response, status)
			return
		}

		if request.Method == "POST" && request.FormValue("inherit") != "" {
			// Get group from library
			item, err := server.Library.GetItem(request.FormValue("inherit"), groupType)
			if os.IsNotExist(err) {
				server.serveResponse(writer, serverResponse{mesgResourceNotFound}, http.StatusNotFound)
				return
			} else if err != nil {
				logger.Log(logger.LevelError, "server", "%s", err)
				server.serveResponse(writer, serverResponse{mesgUnhandledError}, http.StatusInternalServerError)
				return
			}

			// Clone item
			group = &library.Group{}
			utils.Clone(item.(*library.Group), group)

			// Reset item identifier
			group.ID = ""
		} else {
			// Create a new group instance
			group = &library.Group{Item: library.Item{ID: groupID}, Type: groupType}
		}

		group.Modified = time.Now()

		// Parse input JSON for group data
		body, _ := ioutil.ReadAll(request.Body)

		if err := json.Unmarshal(body, group); err != nil {
			logger.Log(logger.LevelError, "server", "%s", err)
			server.serveResponse(writer, serverResponse{mesgResourceInvalid}, http.StatusBadRequest)
			return
		}

		// Store group data
		err := server.Library.StoreItem(group, groupType)
		if response, status := server.parseError(writer, request, err); status != http.StatusOK {
			logger.Log(logger.LevelError, "server", "%s", err)
			server.serveResponse(writer, response, status)
			return
		}

		if request.Method == "POST" {
			writer.Header().Add("Location", strings.TrimRight(request.URL.Path, "/")+"/"+group.ID)
			server.serveResponse(writer, nil, http.StatusCreated)
		} else {
			server.serveResponse(writer, nil, http.StatusOK)
		}

	default:
		server.serveResponse(writer, serverResponse{mesgMethodNotAllowed}, http.StatusMethodNotAllowed)
	}
}

func (server *Server) serveGroupList(writer http.ResponseWriter, request *http.Request) {
	var (
		items         ItemListResponse
		offset, limit int
	)

	if response, status := server.parseListRequest(writer, request, &offset, &limit); status != http.StatusOK {
		server.serveResponse(writer, response, status)
		return
	}

	// Fill groups list
	items = make(ItemListResponse, 0)

	isSource := strings.HasPrefix(request.URL.Path, urlLibraryPath+"sourcegroups/")

	for _, group := range server.Library.Groups {
		if isSource && group.Type != library.LibraryItemSourceGroup ||
			!isSource && group.Type != library.LibraryItemMetricGroup {
			continue
		}

		if request.FormValue("filter") != "" && !utils.FilterMatch(request.FormValue("filter"), group.Name) {
			continue
		}

		items = append(items, &ItemResponse{
			ID:          group.ID,
			Name:        group.Name,
			Description: group.Description,
			Modified:    group.Modified.Format(time.RFC3339),
		})
	}

	response := &listResponse{
		list:   items,
		offset: offset,
		limit:  limit,
	}

	server.applyResponseLimit(writer, request, response)

	server.serveResponse(writer, response.list, http.StatusOK)
}

func (server *Server) serveGroupExpand(writer http.ResponseWriter, request *http.Request) {
	var response []ExpandRequest

	if request.Method != "POST" {
		server.serveResponse(writer, serverResponse{mesgMethodNotAllowed}, http.StatusMethodNotAllowed)
		return
	}

	body, _ := ioutil.ReadAll(request.Body)

	query := ExpandRequest{}
	if err := json.Unmarshal(body, &query); err != nil {
		logger.Log(logger.LevelError, "server", "%s", err)
		server.serveResponse(writer, serverResponse{mesgResourceInvalid}, http.StatusBadRequest)
		return
	}

	for _, entry := range query {
		item := ExpandRequest{}

		origin, err := server.Catalog.GetOrigin(entry[0])
		if err != nil {
			continue
		}

		if strings.HasPrefix(entry[1], library.LibraryGroupPrefix) {
			for _, sourceName := range server.Library.ExpandSourceGroup(
				strings.TrimPrefix(entry[1], library.LibraryGroupPrefix),
			) {
				source, err := origin.GetSource(sourceName)
				if err != nil {
					continue
				}

				if strings.HasPrefix(entry[2], library.LibraryGroupPrefix) {
					for _, metricName := range server.Library.ExpandMetricGroup(
						sourceName,
						strings.TrimPrefix(entry[2], library.LibraryGroupPrefix),
					) {
						if !source.MetricExists(metricName) {
							continue
						}

						item = append(item, [3]string{entry[0], sourceName, metricName})
					}
				} else {
					if !source.MetricExists(entry[2]) {
						continue
					}

					item = append(item, [3]string{entry[0], sourceName, entry[2]})
				}
			}
		} else if strings.HasPrefix(entry[2], library.LibraryGroupPrefix) {
			source, err := origin.GetSource(entry[1])
			if err != nil {
				continue
			}

			for _, metricName := range server.Library.ExpandMetricGroup(
				entry[1],
				strings.TrimPrefix(entry[2], library.LibraryGroupPrefix),
			) {
				if !source.MetricExists(metricName) {
					continue
				}

				item = append(item, [3]string{entry[0], entry[1], metricName})
			}
		} else {
			item = append(item, entry)
		}

		sort.Sort(item)
		response = append(response, item)
	}

	server.serveResponse(writer, response, http.StatusOK)
}

package server

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/facette/facette/pkg/library"
	"github.com/facette/facette/pkg/logger"
	"github.com/facette/facette/pkg/utils"
	"github.com/fatih/set"
)

func (server *Server) serveGraph(writer http.ResponseWriter, request *http.Request) {
	var graph *library.Graph

	if request.Method != "GET" && request.Method != "HEAD" && server.Config.ReadOnly {
		server.serveResponse(writer, serverResponse{mesgReadOnlyMode}, http.StatusForbidden)
		return
	}

	graphID := routeTrimPrefix(request.URL.Path, urlLibraryPath+"graphs")

	switch request.Method {
	case "DELETE":
		if graphID == "" {
			server.serveResponse(writer, serverResponse{mesgMethodNotAllowed}, http.StatusMethodNotAllowed)
			return
		}

		err := server.Library.DeleteItem(graphID, library.LibraryItemGraph)
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
		if graphID == "" {
			server.serveGraphList(writer, request)
			return
		}

		item, err := server.Library.GetItem(graphID, library.LibraryItemGraph)
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
		if response, status := server.parseStoreRequest(writer, request, graphID); status != http.StatusOK {
			server.serveResponse(writer, response, status)
			return
		}

		// Inheritance requested: clone an existing graph
		if request.Method == "POST" && request.FormValue("inherit") != "" {
			item, err := server.Library.GetItem(request.FormValue("inherit"), library.LibraryItemGraph)
			if os.IsNotExist(err) {
				server.serveResponse(writer, serverResponse{mesgResourceNotFound}, http.StatusNotFound)
				return
			} else if err != nil {
				logger.Log(logger.LevelError, "server", "%s", err)
				server.serveResponse(writer, serverResponse{mesgUnhandledError}, http.StatusInternalServerError)
				return
			}

			// Clone item
			graph = &library.Graph{}
			utils.Clone(item.(*library.Graph), graph)

			// Reset item identifier
			graph.ID = ""
		} else {
			// Create a new graph instance
			graph = &library.Graph{Item: library.Item{ID: graphID}}
		}

		// Parse input JSON for graph data
		body, _ := ioutil.ReadAll(request.Body)

		if err := json.Unmarshal(body, graph); err != nil {
			logger.Log(logger.LevelError, "server", "%s", err)
			server.serveResponse(writer, serverResponse{mesgResourceInvalid}, http.StatusBadRequest)
			return
		}

		// Check for graph type consistency (either a graph or an instance but not both)
		if !graph.Template && (graph.Link != "" || len(graph.Attributes) > 0) &&
			(graph.Description != "" || graph.Title != "" || graph.Type != 0 || graph.StackMode != 0 ||
				graph.UnitType != 0 || graph.UnitLegend != "" || graph.Groups != nil) ||
			graph.Template && (graph.Link != "" || len(graph.Attributes) > 0) {
			server.serveResponse(writer, serverResponse{mesgResourceInvalid}, http.StatusBadRequest)
			return
		}

		// Store graph item
		err := server.Library.StoreItem(graph, library.LibraryItemGraph)
		if response, status := server.parseError(writer, request, err); status != http.StatusOK {
			logger.Log(logger.LevelError, "server", "%s", err)
			server.serveResponse(writer, response, status)
			return
		}

		if request.Method == "POST" {
			writer.Header().Add("Location", strings.TrimRight(request.URL.Path, "/")+"/"+graph.ID)
			server.serveResponse(writer, nil, http.StatusCreated)
		} else {
			server.serveResponse(writer, nil, http.StatusOK)
		}

	default:
		server.serveResponse(writer, serverResponse{mesgMethodNotAllowed}, http.StatusMethodNotAllowed)
	}
}

func (server *Server) serveGraphList(writer http.ResponseWriter, request *http.Request) {
	var (
		items         GraphListResponse
		offset, limit int
	)

	if response, status := server.parseListRequest(writer, request, &offset, &limit); status != http.StatusOK {
		server.serveResponse(writer, response, status)
		return
	}

	graphSet := set.New(set.ThreadSafe)

	// Filter on collection if any
	if request.FormValue("collection") != "" {
		item, err := server.Library.GetItem(request.FormValue("collection"), library.LibraryItemCollection)
		if os.IsNotExist(err) {
			server.serveResponse(writer, serverResponse{mesgResourceNotFound}, http.StatusNotFound)
			return
		} else if err != nil {
			logger.Log(logger.LevelError, "server", "%s", err)
			server.serveResponse(writer, serverResponse{mesgUnhandledError}, http.StatusInternalServerError)
			return
		}

		collection := item.(*library.Collection)

		for _, graph := range collection.Entries {
			graphSet.Add(graph.ID)
		}
	}

	// Fill graphs list
	items = make(GraphListResponse, 0)

	// Flag for listing only graph templates
	listType := request.FormValue("type")
	if listType == "" {
		listType = "all"
	} else if listType != "raw" && listType != "template" && listType != "all" {
		logger.Log(logger.LevelWarning, "server", "unknown list type: %s", listType)
		server.serveResponse(writer, serverResponse{mesgRequestInvalid}, http.StatusBadRequest)
		return
	}

	for _, graph := range server.Library.Graphs {
		// Depending on the template flag, filter out either graphs or graph templates
		if request.FormValue("type") != "all" && (graph.Template && listType == "raw" ||
			!graph.Template && listType == "template") {
			continue
		}

		// Filter out graphs that don't belong in the targeted collection
		if !graphSet.IsEmpty() && !graphSet.Has(graph.ID) {
			continue
		}

		if request.FormValue("filter") != "" && !utils.FilterMatch(request.FormValue("filter"), graph.Name) {
			continue
		}

		// If linked graph, expand the templated description field
		description := graph.Description

		if graph.Link != "" {
			item, err := server.Library.GetItem(graph.Link, library.LibraryItemGraph)

			if err != nil {
				logger.Log(logger.LevelError, "server", "graph template not found")
			} else {
				graphTemplate := item.(*library.Graph)

				if description, err = expandStringTemplate(
					graphTemplate.Description,
					graph.Attributes,
				); err != nil {
					logger.Log(logger.LevelError, "server", "failed to expand graph description: %s", err)
				}
			}
		}

		items = append(items, &GraphResponse{
			ItemResponse: ItemResponse{
				ID:          graph.ID,
				Name:        graph.Name,
				Description: description,
				Modified:    graph.Modified.Format(time.RFC3339),
			},
			Link:     graph.Link,
			Template: graph.Template,
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

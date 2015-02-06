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
)

func (server *Server) serveScale(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "GET" && request.Method != "HEAD" && server.Config.ReadOnly {
		server.serveResponse(writer, serverResponse{mesgReadOnlyMode}, http.StatusForbidden)
		return
	}

	scaleID := routeTrimPrefix(request.URL.Path, urlLibraryPath+"scales")

	switch request.Method {
	case "DELETE":
		if scaleID == "" {
			server.serveResponse(writer, serverResponse{mesgMethodNotAllowed}, http.StatusMethodNotAllowed)
			return
		}

		err := server.Library.DeleteItem(scaleID, library.LibraryItemScale)
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
		if scaleID == "" {
			server.serveScaleList(writer, request)
			return
		}

		item, err := server.Library.GetItem(scaleID, library.LibraryItemScale)
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
		var scale *library.Scale

		if response, status := server.parseStoreRequest(writer, request, scaleID); status != http.StatusOK {
			server.serveResponse(writer, response, status)
			return
		}

		if request.Method == "POST" && request.FormValue("inherit") != "" {
			// Get scale from library
			item, err := server.Library.GetItem(request.FormValue("inherit"), library.LibraryItemScale)
			if os.IsNotExist(err) {
				server.serveResponse(writer, serverResponse{mesgResourceNotFound}, http.StatusNotFound)
				return
			} else if err != nil {
				logger.Log(logger.LevelError, "server", "%s", err)
				server.serveResponse(writer, serverResponse{mesgUnhandledError}, http.StatusInternalServerError)
				return
			}

			// Clone item
			scale = &library.Scale{}
			utils.Clone(item.(*library.Scale), scale)

			// Reset item identifier
			scale.ID = ""
		} else {
			// Create a new scale instance
			scale = &library.Scale{Item: library.Item{ID: scaleID}}
		}

		scale.Modified = time.Now()

		// Parse input JSON for scale data
		body, _ := ioutil.ReadAll(request.Body)

		if err := json.Unmarshal(body, scale); err != nil {
			logger.Log(logger.LevelError, "server", "%s", err)
			server.serveResponse(writer, serverResponse{mesgResourceInvalid}, http.StatusBadRequest)
			return
		}

		// Store scale data
		err := server.Library.StoreItem(scale, library.LibraryItemScale)
		if response, status := server.parseError(writer, request, err); status != http.StatusOK {
			logger.Log(logger.LevelError, "server", "%s", err)
			server.serveResponse(writer, response, status)
			return
		}

		if request.Method == "POST" {
			writer.Header().Add("Location", strings.TrimRight(request.URL.Path, "/")+"/"+scale.ID)
			server.serveResponse(writer, nil, http.StatusCreated)
		} else {
			server.serveResponse(writer, nil, http.StatusOK)
		}

	default:
		server.serveResponse(writer, serverResponse{mesgMethodNotAllowed}, http.StatusMethodNotAllowed)
	}
}

func (server *Server) serveScaleList(writer http.ResponseWriter, request *http.Request) {
	var (
		items         ItemListResponse
		offset, limit int
	)

	if response, status := server.parseListRequest(writer, request, &offset, &limit); status != http.StatusOK {
		server.serveResponse(writer, response, status)
		return
	}

	// Fill scales list
	items = make(ItemListResponse, 0)

	for _, scale := range server.Library.Scales {
		if request.FormValue("filter") != "" && !utils.FilterMatch(request.FormValue("filter"), scale.Name) {
			continue
		}

		items = append(items, &ItemResponse{
			ID:          scale.ID,
			Name:        scale.Name,
			Description: scale.Description,
			Modified:    scale.Modified.Format(time.RFC3339),
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

func (server *Server) serveScaleValues(writer http.ResponseWriter, request *http.Request) {
	var (
		items         ScaleValueListResponse
		offset, limit int
	)

	if request.Method != "GET" && request.Method != "HEAD" {
		server.serveResponse(writer, nil, http.StatusMethodNotAllowed)
		return
	}

	if response, status := server.parseListRequest(writer, request, &offset, &limit); status != http.StatusOK {
		server.serveResponse(writer, response, status)
		return
	}

	// Fill scales values list
	items = make(ScaleValueListResponse, 0)

	for _, scale := range server.Library.Scales {
		items = append(items, &ScaleValueResponse{
			Name:  scale.Name,
			Value: scale.Value,
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

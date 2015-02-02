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

func (server *Server) serveUnit(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "GET" && request.Method != "HEAD" && server.Config.ReadOnly {
		server.serveResponse(writer, serverResponse{mesgReadOnlyMode}, http.StatusForbidden)
		return
	}

	unitID := routeTrimPrefix(request.URL.Path, urlLibraryPath+"units")

	switch request.Method {
	case "DELETE":
		if unitID == "" {
			server.serveResponse(writer, serverResponse{mesgMethodNotAllowed}, http.StatusMethodNotAllowed)
			return
		}

		err := server.Library.DeleteItem(unitID, library.LibraryItemUnit)
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
		if unitID == "" {
			server.serveUnitList(writer, request)
			return
		}

		item, err := server.Library.GetItem(unitID, library.LibraryItemUnit)
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
		var unit *library.Unit

		if response, status := server.parseStoreRequest(writer, request, unitID); status != http.StatusOK {
			server.serveResponse(writer, response, status)
			return
		}

		if request.Method == "POST" && request.FormValue("inherit") != "" {
			// Get unit from library
			item, err := server.Library.GetItem(request.FormValue("inherit"), library.LibraryItemUnit)
			if os.IsNotExist(err) {
				server.serveResponse(writer, serverResponse{mesgResourceNotFound}, http.StatusNotFound)
				return
			} else if err != nil {
				logger.Log(logger.LevelError, "server", "%s", err)
				server.serveResponse(writer, serverResponse{mesgUnhandledError}, http.StatusInternalServerError)
				return
			}

			// Clone item
			unit = &library.Unit{}
			utils.Clone(item.(*library.Unit), unit)

			// Reset item identifier
			unit.ID = ""
		} else {
			// Create a new unit instance
			unit = &library.Unit{Item: library.Item{ID: unitID}}
		}

		unit.Modified = time.Now()

		// Parse input JSON for unit data
		body, _ := ioutil.ReadAll(request.Body)

		if err := json.Unmarshal(body, unit); err != nil {
			logger.Log(logger.LevelError, "server", "%s", err)
			server.serveResponse(writer, serverResponse{mesgResourceInvalid}, http.StatusBadRequest)
			return
		}

		// Store unit data
		err := server.Library.StoreItem(unit, library.LibraryItemUnit)
		if response, status := server.parseError(writer, request, err); status != http.StatusOK {
			logger.Log(logger.LevelError, "server", "%s", err)
			server.serveResponse(writer, response, status)
			return
		}

		if request.Method == "POST" {
			writer.Header().Add("Location", strings.TrimRight(request.URL.Path, "/")+"/"+unit.ID)
			server.serveResponse(writer, nil, http.StatusCreated)
		} else {
			server.serveResponse(writer, nil, http.StatusOK)
		}

	default:
		server.serveResponse(writer, serverResponse{mesgMethodNotAllowed}, http.StatusMethodNotAllowed)
	}
}

func (server *Server) serveUnitList(writer http.ResponseWriter, request *http.Request) {
	var (
		items         ItemListResponse
		offset, limit int
	)

	if response, status := server.parseListRequest(writer, request, &offset, &limit); status != http.StatusOK {
		server.serveResponse(writer, response, status)
		return
	}

	// Fill units list
	items = make(ItemListResponse, 0)

	for _, unit := range server.Library.Units {
		if request.FormValue("filter") != "" && !utils.FilterMatch(request.FormValue("filter"), unit.Name) {
			continue
		}

		items = append(items, &ItemResponse{
			ID:          unit.ID,
			Name:        unit.Name,
			Description: unit.Description,
			Modified:    unit.Modified.Format(time.RFC3339),
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

func (server *Server) serveUnitLabels(writer http.ResponseWriter, request *http.Request) {
	var (
		items         UnitValueListResponse
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

	// Fill units values list
	items = make(UnitValueListResponse, 0)

	for _, unit := range server.Library.Units {
		items = append(items, &UnitValueResponse{
			Name:  unit.Name,
			Label: unit.Label,
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

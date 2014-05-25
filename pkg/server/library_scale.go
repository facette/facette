package server

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/facette/facette/pkg/library"
	"github.com/facette/facette/pkg/utils"
)

func (server *Server) handleScale(writer http.ResponseWriter, request *http.Request) {
	scaleID := strings.TrimPrefix(request.URL.Path, urlLibraryPath+"scales/")

	switch request.Method {
	case "DELETE":
		if scaleID == "" {
			server.handleResponse(writer, serverResponse{mesgMethodNotAllowed}, http.StatusMethodNotAllowed)
			return
		}

		err := server.Library.DeleteItem(scaleID, library.LibraryItemScale)
		if os.IsNotExist(err) {
			server.handleResponse(writer, serverResponse{mesgResourceNotFound}, http.StatusNotFound)
			return
		} else if err != nil {
			log.Println("ERROR: " + err.Error())
			server.handleResponse(writer, serverResponse{mesgUnhandledError}, http.StatusInternalServerError)
			return
		}

		server.handleResponse(writer, nil, http.StatusOK)

	case "GET", "HEAD":
		if scaleID == "" {
			server.handleScaleList(writer, request)
			return
		}

		item, err := server.Library.GetItem(scaleID, library.LibraryItemScale)
		if os.IsNotExist(err) {
			server.handleResponse(writer, serverResponse{mesgResourceNotFound}, http.StatusNotFound)
			return
		} else if err != nil {
			log.Println("ERROR: " + err.Error())
			server.handleResponse(writer, serverResponse{mesgUnhandledError}, http.StatusInternalServerError)
			return
		}

		server.handleResponse(writer, item, http.StatusOK)

	case "POST", "PUT":
		var scale *library.Scale

		if response, status := server.parseStoreRequest(writer, request, scaleID); status != http.StatusOK {
			server.handleResponse(writer, response, status)
			return
		}

		if request.Method == "POST" && request.FormValue("inherit") != "" {
			// Get scale from library
			item, err := server.Library.GetItem(request.FormValue("inherit"), library.LibraryItemScale)
			if os.IsNotExist(err) {
				server.handleResponse(writer, serverResponse{mesgResourceNotFound}, http.StatusNotFound)
				return
			} else if err != nil {
				log.Println("ERROR: " + err.Error())
				server.handleResponse(writer, serverResponse{mesgUnhandledError}, http.StatusInternalServerError)
				return
			}

			scale = &library.Scale{}
			*scale = *item.(*library.Scale)

			scale.ID = ""
		} else {
			// Create a new scale instance
			scale = &library.Scale{Item: library.Item{ID: scaleID}}
		}

		scale.Modified = time.Now()

		// Parse input JSON for scale data
		body, _ := ioutil.ReadAll(request.Body)

		if err := json.Unmarshal(body, scale); err != nil {
			log.Println("ERROR: " + err.Error())
			server.handleResponse(writer, serverResponse{mesgResourceInvalid}, http.StatusBadRequest)
			return
		}

		// Store scale data
		err := server.Library.StoreItem(scale, library.LibraryItemScale)
		if response, status := server.parseError(writer, request, err); status != http.StatusOK {
			log.Println("ERROR: " + err.Error())
			server.handleResponse(writer, response, status)
			return
		}

		if request.Method == "POST" {
			writer.Header().Add("Location", strings.TrimRight(request.URL.Path, "/")+"/"+scale.ID)
			server.handleResponse(writer, nil, http.StatusCreated)
		} else {
			server.handleResponse(writer, nil, http.StatusOK)
		}

	default:
		server.handleResponse(writer, serverResponse{mesgMethodNotAllowed}, http.StatusMethodNotAllowed)
	}
}

func (server *Server) handleScaleList(writer http.ResponseWriter, request *http.Request) {
	var offset, limit int

	if response, status := server.parseListRequest(writer, request, &offset, &limit); status != http.StatusOK {
		server.handleResponse(writer, response, status)
		return
	}

	// Fill scales list
	items := make(ItemListResponse, 0)

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

	server.handleResponse(writer, response.list, http.StatusOK)
}

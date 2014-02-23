package server

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/facette/facette/pkg/library"
	"github.com/facette/facette/pkg/utils"
	"github.com/facette/facette/thirdparty/github.com/fatih/set"
	"github.com/facette/facette/thirdparty/github.com/gorilla/mux"
)

// CollectionResponse represents a collection response struct in the server library.
type CollectionResponse struct {
	ItemResponse
	Parent      *string `json:"parent"`
	HasChildren bool    `json:"has_children"`
}

// CollectionListResponse represents a collections list response struct in the server library.
type CollectionListResponse struct {
	Items []*CollectionResponse `json:"items"`
}

func (response CollectionListResponse) Len() int {
	return len(response.Items)
}

func (response CollectionListResponse) Less(i, j int) bool {
	return response.Items[i].Name < response.Items[j].Name
}

func (response CollectionListResponse) Swap(i, j int) {
	response.Items[i], response.Items[j] = response.Items[j], response.Items[i]
}

func (server *Server) collectionHandle(writer http.ResponseWriter, request *http.Request) {
	type tmpCollection struct {
		*library.Collection
		Parent string `json:"parent"`
	}

	collectionID := mux.Vars(request)["id"]

	switch request.Method {
	case "DELETE":
		if collectionID == "" {
			server.handleResponse(writer, http.StatusMethodNotAllowed)
			return
		} else if !server.handleAuth(writer, request) {
			server.handleResponse(writer, http.StatusUnauthorized)
			return
		}

		// Remove collection from library
		err := server.Library.DeleteItem(collectionID, library.LibraryItemCollection)
		if os.IsNotExist(err) {
			server.handleResponse(writer, http.StatusNotFound)
			return
		} else if err != nil {
			log.Println("ERROR: " + err.Error())
			server.handleResponse(writer, http.StatusInternalServerError)
			return
		}

		break

	case "GET", "HEAD":
		if collectionID == "" {
			server.collectionList(writer, request)
			return
		}

		// Get collection from library
		item, err := server.Library.GetItem(collectionID, library.LibraryItemCollection)
		if os.IsNotExist(err) {
			server.handleResponse(writer, http.StatusNotFound)
			return
		} else if err != nil {
			log.Println("ERROR: " + err.Error())
			server.handleResponse(writer, http.StatusInternalServerError)
			return
		}

		// Dump JSON response
		server.handleJSON(writer, item)

		break

	case "POST", "PUT":
		if request.Method == "POST" && collectionID != "" || request.Method == "PUT" && collectionID == "" {
			server.handleResponse(writer, http.StatusMethodNotAllowed)
			return
		} else if utils.RequestGetContentType(request) != "application/json" {
			server.handleResponse(writer, http.StatusUnsupportedMediaType)
			return
		} else if !server.handleAuth(writer, request) {
			server.handleResponse(writer, http.StatusUnauthorized)
			return
		}

		collectionTemp := &tmpCollection{
			Collection: &library.Collection{
				Item: library.Item{ID: collectionID},
			},
		}

		if request.Method == "POST" && request.FormValue("inherit") != "" {
			// Get collection from library
			item, err := server.Library.GetItem(request.FormValue("inherit"), library.LibraryItemCollection)
			if os.IsNotExist(err) {
				server.handleResponse(writer, http.StatusNotFound)
				return
			} else if err != nil {
				log.Println("ERROR: " + err.Error())
				server.handleResponse(writer, http.StatusInternalServerError)
				return
			}

			*collectionTemp.Collection = *item.(*library.Collection)
			collectionTemp.Collection.ID = ""
			collectionTemp.Collection.Children = nil
		}

		collectionTemp.Collection.Modified = time.Now()

		// Parse input JSON for collection data
		body, _ := ioutil.ReadAll(request.Body)

		if err := json.Unmarshal(body, &collectionTemp); err != nil {
			log.Println("ERROR: " + err.Error())
			server.handleResponse(writer, http.StatusBadRequest)
			return
		}

		// Update parent relation
		if item, _ := server.Library.GetItem(collectionTemp.Parent, library.LibraryItemCollection); item != nil {
			collection := item.(*library.Collection)

			// Register parent relation
			collectionTemp.Collection.Parent = collection
			collectionTemp.Collection.ParentID = collectionTemp.Collection.Parent.ID
			collection.Children = append(collection.Children, collectionTemp.Collection)
		} else {
			// Remove existing parent relation
			if item, _ := server.Library.GetItem(collectionTemp.Collection.ID,
				library.LibraryItemCollection); item != nil {
				collection := item.(*library.Collection)

				if collection.Parent != nil {
					for index, child := range collection.Parent.Children {
						if reflect.DeepEqual(child, collection) {
							collection.Parent.Children = append(collection.Parent.Children[:index],
								collection.Parent.Children[index+1:]...)
							break
						}
					}
				}
			}
		}

		// Keep current children list
		if item, _ := server.Library.GetItem(collectionTemp.Collection.ID, library.LibraryItemCollection); item != nil {
			collectionTemp.Collection.Children = item.(*library.Collection).Children
		}

		// Store collection data
		err := server.Library.StoreItem(collectionTemp.Collection, library.LibraryItemCollection)
		if err == os.ErrInvalid {
			server.handleResponse(writer, http.StatusBadRequest)
			return
		} else if os.IsExist(err) {
			server.handleResponse(writer, http.StatusConflict)
			return
		} else if os.IsNotExist(err) {
			server.handleResponse(writer, http.StatusNotFound)
			return
		} else if err != nil {
			log.Println("ERROR: " + err.Error())
			server.handleResponse(writer, http.StatusInternalServerError)
			return
		}

		if request.Method == "POST" {
			writer.Header().Add("Location", strings.TrimRight(request.URL.Path, "/")+"/"+collectionTemp.Collection.ID)
			server.handleResponse(writer, http.StatusCreated)
		}

		break

	default:
		server.handleResponse(writer, http.StatusMethodNotAllowed)
	}
}

func (server *Server) collectionList(writer http.ResponseWriter, request *http.Request) {
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

	// Check for item exclusion
	excludeSet := set.New()

	collectionStack := []*library.Collection{}

	if request.FormValue("exclude") != "" {
		if item, err := server.Library.GetItem(request.FormValue("exclude"), library.LibraryItemCollection); err == nil {
			collectionStack = append(collectionStack, item.(*library.Collection))
		}

		for len(collectionStack) > 0 {
			collection, collectionStack := collectionStack[0], collectionStack[1:]
			excludeSet.Add(collection.ID)
			collectionStack = append(collectionStack, collection.Children...)
		}
	}

	// Get and filter collections list
	response := CollectionListResponse{}

	for _, collection := range server.Library.Collections {
		if request.FormValue("parent") != "" && (request.FormValue("parent") == "" &&
			collection.Parent != nil || request.FormValue("parent") != "" &&
			(collection.Parent == nil || collection.Parent.ID != request.FormValue("parent"))) {
			continue
		}

		if request.FormValue("filter") != "" && !utils.FilterMatch(strings.ToLower(request.FormValue("filter")),
			strings.ToLower(collection.Name)) {
			continue
		}

		// Skip excluded items
		if excludeSet.Has(collection.ID) {
			continue
		}

		collectionItem := &CollectionResponse{ItemResponse: ItemResponse{
			ID:          collection.ID,
			Name:        collection.Name,
			Description: collection.Description,
			Modified:    collection.Modified.Format(time.RFC3339),
		}, HasChildren: len(collection.Children) > 0}

		if collection.Parent != nil {
			collectionItem.Parent = &collection.Parent.ID
		}

		response.Items = append(response.Items, collectionItem)
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

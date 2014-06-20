package server

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/facette/facette/pkg/library"
	"github.com/facette/facette/pkg/utils"
	"github.com/facette/facette/thirdparty/github.com/fatih/set"
)

func (server *Server) serveCollection(writer http.ResponseWriter, request *http.Request) {
	type tmpCollection struct {
		*library.Collection
		Parent string `json:"parent"`
	}

	collectionID := strings.TrimPrefix(request.URL.Path, urlLibraryPath+"collections/")

	switch request.Method {
	case "DELETE":
		if collectionID == "" {
			server.serveResponse(writer, serverResponse{mesgMethodNotAllowed}, http.StatusMethodNotAllowed)
			return
		}

		err := server.Library.DeleteItem(collectionID, library.LibraryItemCollection)
		if os.IsNotExist(err) {
			server.serveResponse(writer, serverResponse{mesgResourceNotFound}, http.StatusNotFound)
			return
		} else if err != nil {
			log.Println("ERROR: " + err.Error())
			server.serveResponse(writer, serverResponse{mesgUnhandledError}, http.StatusInternalServerError)
			return
		}

		server.serveResponse(writer, nil, http.StatusOK)

	case "GET", "HEAD":
		if collectionID == "" {
			server.serveCollectionList(writer, request)
			return
		}

		item, err := server.Library.GetItem(collectionID, library.LibraryItemCollection)
		if os.IsNotExist(err) {
			server.serveResponse(writer, serverResponse{mesgResourceNotFound}, http.StatusNotFound)
			return
		} else if err != nil {
			log.Println("ERROR: " + err.Error())
			server.serveResponse(writer, serverResponse{mesgUnhandledError}, http.StatusInternalServerError)
			return
		}

		server.serveResponse(writer, item, http.StatusOK)

	case "POST", "PUT":
		if response, status := server.parseStoreRequest(writer, request, collectionID); status != http.StatusOK {
			server.serveResponse(writer, response, status)
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
				server.serveResponse(writer, serverResponse{mesgResourceNotFound}, http.StatusNotFound)
				return
			} else if err != nil {
				log.Println("ERROR: " + err.Error())
				server.serveResponse(writer, serverResponse{mesgUnhandledError}, http.StatusInternalServerError)
				return
			}

			utils.Clone(item.(*library.Collection), collectionTemp.Collection)

			collectionTemp.Collection.ID = ""
			collectionTemp.Collection.Children = nil
		}

		collectionTemp.Collection.Modified = time.Now()

		// Parse input JSON for collection data
		body, _ := ioutil.ReadAll(request.Body)

		if err := json.Unmarshal(body, &collectionTemp); err != nil {
			log.Println("ERROR: " + err.Error())
			server.serveResponse(writer, serverResponse{mesgResourceInvalid}, http.StatusBadRequest)
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
		if response, status := server.parseError(writer, request, err); status != http.StatusOK {
			log.Println("ERROR: " + err.Error())
			server.serveResponse(writer, response, status)
			return
		}

		if request.Method == "POST" {
			writer.Header().Add("Location", strings.TrimRight(request.URL.Path, "/")+"/"+collectionTemp.Collection.ID)
			server.serveResponse(writer, nil, http.StatusCreated)
		} else {
			server.serveResponse(writer, nil, http.StatusOK)
		}

	default:
		server.serveResponse(writer, serverResponse{mesgMethodNotAllowed}, http.StatusMethodNotAllowed)
	}
}

func (server *Server) serveCollectionList(writer http.ResponseWriter, request *http.Request) {
	var (
		collection    *library.Collection
		offset, limit int
	)

	if response, status := server.parseListRequest(writer, request, &offset, &limit); status != http.StatusOK {
		server.serveResponse(writer, response, status)
		return
	}

	// Check for item exclusion
	excludeSet := set.New(set.ThreadSafe)

	collectionStack := make([]*library.Collection, 0)

	if request.FormValue("exclude") != "" {
		if item, err := server.Library.GetItem(request.FormValue("exclude"), library.LibraryItemCollection); err == nil {
			collectionStack = append(collectionStack, item.(*library.Collection))
		}

		for len(collectionStack) > 0 {
			collection, collectionStack = collectionStack[0], collectionStack[1:]
			excludeSet.Add(collection.ID)
			collectionStack = append(collectionStack, collection.Children...)
		}
	}

	// Fill collections list
	items := make(CollectionListResponse, 0)

	for _, collection := range server.Library.Collections {
		if request.FormValue("parent") != "" && (request.FormValue("parent") == "" &&
			collection.Parent != nil || request.FormValue("parent") != "" &&
			(collection.Parent == nil || collection.Parent.ID != request.FormValue("parent"))) {
			continue
		}

		if request.FormValue("filter") != "" && !utils.FilterMatch(request.FormValue("filter"), collection.Name) {
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

		items = append(items, collectionItem)
	}

	response := &listResponse{
		list:   items,
		offset: offset,
		limit:  limit,
	}

	server.applyResponseLimit(writer, request, response)

	server.serveResponse(writer, response.list, http.StatusOK)
}

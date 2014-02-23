package server

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/facette/facette/pkg/library"
	"github.com/facette/facette/pkg/utils"
	"github.com/facette/facette/thirdparty/github.com/gorilla/mux"
)

// ExpandRequest represents an expand request struct in the server library.
type ExpandRequest [][3]string

func (tuple ExpandRequest) Len() int {
	return len(tuple)
}

func (tuple ExpandRequest) Less(i, j int) bool {
	return tuple[i][0]+tuple[i][1]+tuple[i][2] < tuple[j][0]+tuple[j][1]+tuple[j][2]
}

func (tuple ExpandRequest) Swap(i, j int) {
	tuple[i], tuple[j] = tuple[j], tuple[i]
}

func (server *Server) groupExpand(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "POST" {
		server.handleResponse(writer, http.StatusMethodNotAllowed)
		return
	}

	body, _ := ioutil.ReadAll(request.Body)
	query := ExpandRequest{}

	if err := json.Unmarshal(body, &query); err != nil {
		log.Println("ERROR: " + err.Error())
		server.handleResponse(writer, http.StatusBadRequest)
		return
	}

	response := []ExpandRequest{}

	for _, entry := range query {
		item := ExpandRequest{}

		if strings.HasPrefix(entry[1], library.LibraryGroupPrefix) {
			for _, sourceName := range server.Library.ExpandGroup(strings.TrimPrefix(entry[1],
				library.LibraryGroupPrefix), library.LibraryItemSourceGroup) {
				if strings.HasPrefix(entry[2], library.LibraryGroupPrefix) {
					for _, metricName := range server.Library.ExpandGroup(strings.TrimPrefix(entry[2],
						library.LibraryGroupPrefix), library.LibraryItemMetricGroup) {
						item = append(item, [3]string{entry[0], sourceName, metricName})
					}
				} else {
					item = append(item, [3]string{entry[0], sourceName, entry[2]})
				}
			}
		} else if strings.HasPrefix(entry[2], library.LibraryGroupPrefix) {
			for _, metricName := range server.Library.ExpandGroup(strings.TrimPrefix(entry[2],
				library.LibraryGroupPrefix), library.LibraryItemMetricGroup) {
				item = append(item, [3]string{entry[0], entry[1], metricName})
			}
		} else {
			item = append(item, entry)
		}

		sort.Sort(item)
		response = append(response, item)
	}

	server.handleJSON(writer, response)
}

func (server *Server) groupHandle(writer http.ResponseWriter, request *http.Request) {
	var groupType int

	groupID := mux.Vars(request)["id"]

	if strings.HasPrefix(request.URL.Path, URLLibraryPath+"/sourcegroups") {
		groupType = library.LibraryItemSourceGroup
	} else if strings.HasPrefix(request.URL.Path, URLLibraryPath+"/metricgroups") {
		groupType = library.LibraryItemMetricGroup
	}

	switch request.Method {
	case "DELETE":
		if groupID == "" {
			server.handleResponse(writer, http.StatusMethodNotAllowed)
			return
		} else if !server.handleAuth(writer, request) {
			server.handleResponse(writer, http.StatusUnauthorized)
			return
		}

		// Remove group from library
		err := server.Library.DeleteItem(groupID, groupType)
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
		if groupID == "" {
			server.libraryList(writer, request)
			return
		}

		// Get group from library
		item, err := server.Library.GetItem(groupID, groupType)
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
		var group *library.Group

		if request.Method == "POST" && groupID != "" || request.Method == "PUT" && groupID == "" {
			server.handleResponse(writer, http.StatusMethodNotAllowed)
			return
		} else if utils.RequestGetContentType(request) != "application/json" {
			server.handleResponse(writer, http.StatusUnsupportedMediaType)
			return
		} else if !server.handleAuth(writer, request) {
			server.handleResponse(writer, http.StatusUnauthorized)
			return
		}

		if request.Method == "POST" && request.FormValue("inherit") != "" {
			// Get group from library
			item, err := server.Library.GetItem(request.FormValue("inherit"), groupType)
			if os.IsNotExist(err) {
				server.handleResponse(writer, http.StatusNotFound)
				return
			} else if err != nil {
				log.Println("ERROR: " + err.Error())
				server.handleResponse(writer, http.StatusInternalServerError)
				return
			}

			group = &library.Group{}
			*group = *item.(*library.Group)

			group.ID = ""
		} else {
			group = &library.Group{Item: library.Item{ID: groupID}, Type: groupType}
		}

		group.Modified = time.Now()

		// Parse input JSON for group data
		body, _ := ioutil.ReadAll(request.Body)

		if err := json.Unmarshal(body, group); err != nil {
			log.Println("ERROR: " + err.Error())
			server.handleResponse(writer, http.StatusBadRequest)
			return
		}

		// Store group data
		err := server.Library.StoreItem(group, groupType)
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
			writer.Header().Add("Location", strings.TrimRight(request.URL.Path, "/")+"/"+group.ID)
			server.handleResponse(writer, http.StatusCreated)
		}

		break

	default:
		server.handleResponse(writer, http.StatusMethodNotAllowed)
		break
	}
}

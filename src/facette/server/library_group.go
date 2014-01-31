package server

import (
	"encoding/json"
	"facette/library"
	"facette/utils"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

type expandRequest [][3]string

func (tuple expandRequest) Len() int {
	return len(tuple)
}

func (tuple expandRequest) Less(i, j int) bool {
	return tuple[i][0]+tuple[i][1]+tuple[i][2] < tuple[j][0]+tuple[j][1]+tuple[j][2]
}

func (tuple expandRequest) Swap(i, j int) {
	tuple[i], tuple[j] = tuple[j], tuple[i]
}

func (server *Server) groupExpand(writer http.ResponseWriter, request *http.Request) {
	var (
		body     []byte
		err      error
		item     expandRequest
		query    expandRequest
		response []expandRequest
	)

	if request.Method != "POST" {
		server.handleResponse(writer, http.StatusMethodNotAllowed)
		return
	}

	body, _ = ioutil.ReadAll(request.Body)

	if err = json.Unmarshal(body, &query); err != nil {
		log.Println("ERROR: " + err.Error())
		server.handleResponse(writer, http.StatusBadRequest)
		return
	}

	for _, entry := range query {
		item = expandRequest{}

		if strings.HasPrefix(entry[1], "group:") {
			for _, sourceName := range server.Library.ExpandGroup(entry[1][6:], library.LibraryItemSourceGroup) {
				if strings.HasPrefix(entry[2], "group:") {
					for _, metricName := range server.Library.ExpandGroup(entry[2][6:],
						library.LibraryItemMetricGroup) {
						item = append(item, [3]string{entry[0], sourceName, metricName})
					}
				} else {
					item = append(item, [3]string{entry[0], sourceName, entry[2]})
				}
			}
		} else if strings.HasPrefix(entry[2], "group:") {
			for _, metricName := range server.Library.ExpandGroup(entry[2][6:], library.LibraryItemMetricGroup) {
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
	var (
		body      []byte
		err       error
		group     *library.Group
		groupID   string
		groupType int
		item      interface{}
	)

	groupID = mux.Vars(request)["id"]

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
		err = server.Library.DeleteItem(groupID, groupType)
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
		item, err = server.Library.GetItem(groupID, groupType)

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
			item, err = server.Library.GetItem(request.FormValue("inherit"), groupType)
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
		body, _ = ioutil.ReadAll(request.Body)

		if err = json.Unmarshal(body, &group); err != nil {
			log.Println("ERROR: " + err.Error())
			server.handleResponse(writer, http.StatusBadRequest)
			return
		}

		// Store group data
		err = server.Library.StoreItem(group, groupType)
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

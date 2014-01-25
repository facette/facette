package server

import (
	"encoding/json"
	"facette/backend"
	"facette/common"
	"facette/library"
	"facette/utils"
	"fmt"
	"github.com/fatih/set"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	defaultPlotSample = 400
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

type plotRequest struct {
	Time        string    `json:"time"`
	Range       string    `json:"range"`
	Sample      int       `json:"sample"`
	Constants   []float64 `json:"constants"`
	Percentiles []float64 `json:"percentiles"`
	Graph       string    `json:"graph"`
	Origin      string    `json:"origin"`
	Source      string    `json:"source"`
	Metric      string    `json:"metric"`
	Template    string    `json:"template"`
	Filter      string    `json:"filter"`
}

type statusResponse struct {
	Message string `json:"message"`
}

type libraryItemResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Modified    string `json:"modified"`
}

type libraryListResponse struct {
	Items []*libraryItemResponse `json:"items"`
}

func (response libraryListResponse) Len() int {
	return len(response.Items)
}

func (response libraryListResponse) Less(i, j int) bool {
	return response.Items[i].Name < response.Items[j].Name
}

func (response libraryListResponse) Swap(i, j int) {
	response.Items[i], response.Items[j] = response.Items[j], response.Items[i]
}

type collectionItemResponse struct {
	libraryItemResponse
	Parent      *string `json:"parent"`
	HasChildren bool    `json:"has_children"`
}

type collectionListResponse struct {
	Items []*collectionItemResponse `json:"items"`
}

func (response collectionListResponse) Len() int {
	return len(response.Items)
}

func (response collectionListResponse) Less(i, j int) bool {
	return response.Items[i].Name < response.Items[j].Name
}

func (response collectionListResponse) Swap(i, j int) {
	response.Items[i], response.Items[j] = response.Items[j], response.Items[i]
}

type serieResponse struct {
	Name    string                      `json:"name"`
	Plots   []common.PlotValue          `json:"plots"`
	Info    map[string]common.PlotValue `json:"info"`
	Options map[string]interface{}      `json:"options"`
}

type stackResponse struct {
	Name   string           `json:"name"`
	Series []*serieResponse `json:"series"`
}

type plotResponse struct {
	ID          string           `json:"id"`
	Start       string           `json:"start"`
	End         string           `json:"end"`
	Step        float64          `json:"step"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Type        int              `json:"type"`
	StackMode   int              `json:"stack_mode"`
	Stacks      []*stackResponse `json:"stacks"`
	Modified    time.Time        `json:"modified"`
}

func (server *Server) libraryList(writer http.ResponseWriter, request *http.Request) {
	var (
		collection *library.Collection
		err        error
		graphSet   *set.Set
		isSource   bool
		item       interface{}
		limit      int
		offset     int
		response   libraryListResponse
		skip       bool
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

	if request.URL.Path == URLLibraryPath+"/sourcegroups" || request.URL.Path == URLLibraryPath+"/metricgroups" {
		isSource = request.URL.Path == URLLibraryPath+"/sourcegroups"

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

			response.Items = append(response.Items, &libraryItemResponse{
				ID:          group.ID,
				Name:        group.Name,
				Description: group.Description,
				Modified:    group.Modified.Format(time.RFC3339),
			})
		}
	} else if request.URL.Path == URLLibraryPath+"/graphs" {
		graphSet = set.New()

		// Filter by collection
		if request.FormValue("collection") != "" {
			item, err = server.Library.GetItem(request.FormValue("collection"), library.LibraryItemCollection)
			if os.IsNotExist(err) {
				skip = true
			} else if err != nil {
				server.handleResponse(writer, http.StatusInternalServerError)
				return
			}

			collection = item.(*library.Collection)

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

				response.Items = append(response.Items, &libraryItemResponse{
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

func (server *Server) graphHandle(writer http.ResponseWriter, request *http.Request) {
	var (
		body    []byte
		graph   *library.Graph
		graphID string
		err     error
		item    interface{}
	)

	graphID = mux.Vars(request)["id"]

	switch request.Method {
	case "DELETE":
		if graphID == "" {
			server.handleResponse(writer, http.StatusMethodNotAllowed)
			return
		} else if !server.handleAuth(writer, request) {
			return
		}

		// Remove graph from library
		err = server.Library.DeleteItem(graphID, library.LibraryItemGraph)
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
		if graphID == "" {
			server.libraryList(writer, request)
			return
		}

		// Get graph from library
		item, err = server.Library.GetItem(graphID, library.LibraryItemGraph)

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
		if request.Method == "POST" && graphID != "" || request.Method == "PUT" && graphID == "" {
			server.handleResponse(writer, http.StatusMethodNotAllowed)
			return
		} else if utils.RequestGetContentType(request) != "application/json" {
			server.handleResponse(writer, http.StatusUnsupportedMediaType)
			return
		} else if !server.handleAuth(writer, request) {
			return
		}

		if request.Method == "POST" && request.FormValue("inherit") != "" {
			// Get graph from library
			item, err = server.Library.GetItem(request.FormValue("inherit"), library.LibraryItemGraph)

			if os.IsNotExist(err) {
				server.handleResponse(writer, http.StatusNotFound)
				return
			} else if err != nil {
				log.Println("ERROR: " + err.Error())
				server.handleResponse(writer, http.StatusInternalServerError)
				return
			}

			graph = &library.Graph{}
			*graph = *item.(*library.Graph)

			graph.ID = ""
		} else {
			graph = &library.Graph{Item: library.Item{ID: graphID}}
		}

		graph.Modified = time.Now()

		// Parse input JSON for graph data
		body, _ = ioutil.ReadAll(request.Body)

		if err = json.Unmarshal(body, graph); err != nil {
			log.Println("ERROR: " + err.Error())
			server.handleResponse(writer, http.StatusBadRequest)
			return
		}

		// Store graph data
		if request.FormValue("volatile") != "" {
			graph.Volatile = true
		} else {
			graph.Volatile = false
		}

		err = server.Library.StoreItem(graph, library.LibraryItemGraph)
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
			writer.Header().Add("Location", strings.TrimRight(request.URL.Path, "/")+"/"+graph.ID)
			server.handleResponse(writer, http.StatusCreated)
		}

		break

	default:
		server.handleResponse(writer, http.StatusMethodNotAllowed)
	}
}

func (server *Server) collectionHandle(writer http.ResponseWriter, request *http.Request) {
	type tmpCollection struct {
		*library.Collection
		Parent string `json:"parent"`
	}

	var (
		body           []byte
		collection     *library.Collection
		collectionID   string
		collectionTemp *tmpCollection
		err            error
		item           interface{}
	)

	collectionID = mux.Vars(request)["id"]

	switch request.Method {
	case "DELETE":
		if collectionID == "" {
			server.handleResponse(writer, http.StatusMethodNotAllowed)
			return
		} else if !server.handleAuth(writer, request) {
			return
		}

		// Remove collection from library
		err = server.Library.DeleteItem(collectionID, library.LibraryItemCollection)
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
		item, err = server.Library.GetItem(collectionID, library.LibraryItemCollection)

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
			return
		}

		collectionTemp = &tmpCollection{
			Collection: &library.Collection{
				Item: library.Item{ID: collectionID},
			},
		}

		if request.Method == "POST" && request.FormValue("inherit") != "" {
			// Get collection from library
			item, err = server.Library.GetItem(request.FormValue("inherit"), library.LibraryItemCollection)

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
		body, _ = ioutil.ReadAll(request.Body)

		if err = json.Unmarshal(body, &collectionTemp); err != nil {
			log.Println("ERROR: " + err.Error())
			server.handleResponse(writer, http.StatusBadRequest)
			return
		}

		// Update parent relation
		if item, _ = server.Library.GetItem(collectionTemp.Parent, library.LibraryItemCollection); item != nil {
			collection = item.(*library.Collection)

			// Register parent relation
			collectionTemp.Collection.Parent = collection
			collectionTemp.Collection.ParentID = collectionTemp.Collection.Parent.ID
			collection.Children = append(collection.Children, collectionTemp.Collection)
		} else {
			// Remove existing parent relation
			if item, _ = server.Library.GetItem(collectionTemp.Collection.ID,
				library.LibraryItemCollection); item != nil {
				collection = item.(*library.Collection)

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

		// Store collection data
		err = server.Library.StoreItem(collectionTemp.Collection, library.LibraryItemCollection)
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
		collection      *library.Collection
		collectionItem  *collectionItemResponse
		collectionStack []*library.Collection
		err             error
		excludeSet      *set.Set
		item            interface{}
		limit           int
		offset          int
		response        collectionListResponse
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
	excludeSet = set.New()

	if request.FormValue("exclude") != "" {
		if item, err = server.Library.GetItem(request.FormValue("exclude"),
			library.LibraryItemCollection); err == nil {
			collectionStack = append(collectionStack, item.(*library.Collection))
		}

		for len(collectionStack) > 0 {
			collection, collectionStack = collectionStack[0], collectionStack[1:]
			excludeSet.Add(collection.ID)
			collectionStack = append(collectionStack, collection.Children...)
		}
	}

	// Get and filter collections list
	for _, collection := range server.Library.Collections {
		if request.FormValue("parent") != "" && (request.FormValue("parent") == "" &&
			collection.Parent != nil || request.FormValue("parent") != "" &&
			(collection.Parent == nil || collection.Parent.ID != request.FormValue("parent"))) {
			continue
		}

		if request.FormValue("filter") != "" {
			if !utils.FilterMatch(strings.ToLower(request.FormValue("filter")), strings.ToLower(collection.Name)) {
				continue
			}
		}

		// Skip excluded items
		if excludeSet.Has(collection.ID) {
			continue
		}

		collectionItem = &collectionItemResponse{libraryItemResponse: libraryItemResponse{
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

func (server *Server) plotHandle(writer http.ResponseWriter, request *http.Request) {
	var (
		body          []byte
		graph         *library.Graph
		data          []map[string]*backend.PlotResult
		endTime       time.Time
		err           error
		groupOptions  map[string]map[string]interface{}
		item          interface{}
		originBackend backend.BackendHandler
		plotMax       int
		plotReq       *plotRequest
		plotResult    map[string]*backend.PlotResult
		query         *backend.GroupQuery
		response      *plotResponse
		stack         *stackResponse
		startTime     time.Time
		step          time.Duration
	)

	if request.Method != "POST" && request.Method != "HEAD" {
		server.handleResponse(writer, http.StatusMethodNotAllowed)
		return
	} else if utils.RequestGetContentType(request) != "application/json" {
		server.handleResponse(writer, http.StatusUnsupportedMediaType)
		return
	}

	// Parse input JSON for graph data
	body, _ = ioutil.ReadAll(request.Body)

	if err = json.Unmarshal(body, &plotReq); err != nil {
		log.Println("ERROR: " + err.Error())
		server.handleResponse(writer, http.StatusBadRequest)
		return
	}

	if plotReq.Origin != "" && plotReq.Template != "" {
		plotReq.Graph = plotReq.Origin + "\x30" + plotReq.Template
	} else if plotReq.Origin != "" && plotReq.Metric != "" {
		plotReq.Graph = plotReq.Origin + "\x30" + plotReq.Metric
	}

	if plotReq.Time == "" {
		endTime = time.Now()
	} else if strings.HasPrefix(strings.Trim(plotReq.Range, " "), "-") {
		if endTime, err = time.Parse(time.RFC3339, plotReq.Time); err != nil {
			log.Println("ERROR: " + err.Error())
			server.handleResponse(writer, http.StatusBadRequest)
			return
		}
	} else {
		if startTime, err = time.Parse(time.RFC3339, plotReq.Time); err != nil {
			log.Println("ERROR: " + err.Error())
			server.handleResponse(writer, http.StatusBadRequest)
			return
		}
	}

	if startTime.IsZero() {
		if startTime, err = utils.TimeApplyRange(endTime, plotReq.Range); err != nil {
			log.Println("ERROR: " + err.Error())
			server.handleResponse(writer, http.StatusBadRequest)
			return
		}
	} else if endTime, err = utils.TimeApplyRange(startTime, plotReq.Range); err != nil {
		log.Println("ERROR: " + err.Error())
		server.handleResponse(writer, http.StatusBadRequest)
		return
	}

	if plotReq.Sample == 0 {
		plotReq.Sample = defaultPlotSample
	}

	// Get graph from library
	if plotReq.Template != "" {
		graph, err = server.Library.GetGraphTemplate(plotReq.Origin, plotReq.Source, plotReq.Template, plotReq.Filter)
	} else if plotReq.Metric != "" {
		graph, err = server.Library.GetGraphMetric(plotReq.Origin, plotReq.Source, plotReq.Metric)
	} else if item, err = server.Library.GetItem(plotReq.Graph, library.LibraryItemGraph); err == nil {
		graph = item.(*library.Graph)
	}

	if err != nil {
		log.Println("ERROR: " + err.Error())

		if os.IsNotExist(err) {
			server.handleResponse(writer, http.StatusNotFound)
		} else {
			server.handleResponse(writer, http.StatusBadRequest)
		}

		return
	}

	step = endTime.Sub(startTime) / time.Duration(plotReq.Sample)

	// Get plots data
	groupOptions = make(map[string]map[string]interface{})

	for _, stackItem := range graph.Stacks {
		for _, groupItem := range stackItem.Groups {
			if query, originBackend, err = server.plotPrepareQuery(plotReq, groupItem); err != nil {
				log.Println("ERROR: " + err.Error())
				server.handleResponse(writer, http.StatusBadRequest)
				return
			}

			groupOptions[groupItem.Name] = groupItem.Options

			if plotResult, err = originBackend.GetPlots(query, startTime, endTime, step,
				plotReq.Percentiles); err != nil {
				log.Println("ERROR: " + err.Error())
				server.handleResponse(writer, http.StatusInternalServerError)
				return
			}

			data = append(data, plotResult)
		}
	}

	response = &plotResponse{
		ID:          graph.ID,
		Start:       startTime.Format(time.RFC3339),
		End:         endTime.Format(time.RFC3339),
		Step:        step.Seconds(),
		Name:        graph.Name,
		Description: graph.Description,
		Type:        graph.Type,
		StackMode:   graph.StackMode,
		Modified:    graph.Modified,
	}

	if len(data) == 0 {
		server.handleJSON(writer, statusResponse{"No data"})
		return
	}

	for _, stackItem := range graph.Stacks {
		stack = &stackResponse{Name: stackItem.Name}

		for _, groupItem := range stackItem.Groups {
			plotResult, data = data[0], data[1:]

			for serieName, serieResult := range plotResult {
				if len(serieResult.Plots) > plotMax {
					plotMax = len(serieResult.Plots)
				}

				stack.Series = append(stack.Series, &serieResponse{
					Name:    serieName,
					Plots:   serieResult.Plots,
					Info:    serieResult.Info,
					Options: groupOptions[groupItem.Name],
				})
			}
		}

		response.Stacks = append(response.Stacks, stack)
	}

	if plotMax > 0 {
		response.Step = (endTime.Sub(startTime) / time.Duration(plotMax)).Seconds()
	}

	server.handleJSON(writer, response)
}

func (server *Server) plotPrepareQuery(plotReq *plotRequest,
	groupItem *library.OperGroup) (*backend.GroupQuery, backend.BackendHandler, error) {
	var (
		query         *backend.GroupQuery
		originBackend backend.BackendHandler
		serieSources  []string
	)

	query = &backend.GroupQuery{
		Name:  groupItem.Name,
		Type:  groupItem.Type,
		Scale: groupItem.Scale,
	}

	originBackend = nil

	for _, serieItem := range groupItem.Series {
		// Check for backend errors or conflicts
		if _, ok := server.Catalog.Origins[serieItem.Origin]; !ok {
			return nil, nil, fmt.Errorf("unknown `%s' serie origin", serieItem.Origin)
		} else if originBackend == nil {
			originBackend = server.Catalog.Origins[serieItem.Origin].Backend
		} else if originBackend != server.Catalog.Origins[serieItem.Origin].Backend {
			return nil, nil, fmt.Errorf("backends differ between series")
		}

		if plotReq.Template != "" {
			serieSources = []string{plotReq.Source}
		} else if strings.HasPrefix(serieItem.Source, "group:") {
			serieSources = server.Library.ExpandGroup(serieItem.Source[6:], library.LibraryItemSourceGroup)
		} else {
			serieSources = []string{serieItem.Source}
		}

		for _, serieSource := range serieSources {
			if strings.HasPrefix(serieItem.Metric, "group:") {
				for index, serieChunk := range server.Library.ExpandGroup(serieItem.Metric[6:],
					library.LibraryItemMetricGroup) {
					query.Series = append(query.Series, &backend.SerieQuery{
						Name: fmt.Sprintf("%s-%d", serieItem.Name, index),
						Metric: server.Catalog.GetMetric(
							serieItem.Origin,
							serieSource,
							serieChunk,
						),
						Scale: serieItem.Scale,
					})
				}
			} else {
				query.Series = append(query.Series, &backend.SerieQuery{
					Name: serieItem.Name,
					Metric: server.Catalog.GetMetric(
						serieItem.Origin,
						serieSource,
						serieItem.Metric,
					),
					Scale: serieItem.Scale,
				})
			}
		}
	}

	if len(query.Series) == 0 {
		return nil, nil, fmt.Errorf("no serie defined")
	}

	return query, originBackend, nil
}

func (server *Server) plotValues(writer http.ResponseWriter, request *http.Request) {
	var (
		body          []byte
		err           error
		graph         *library.Graph
		item          interface{}
		originBackend backend.BackendHandler
		plotReq       *plotRequest
		query         *backend.GroupQuery
		refTime       time.Time
		response      map[string]map[string]common.PlotValue
		values        map[string]map[string]common.PlotValue
	)

	if request.Method != "POST" && request.Method != "HEAD" {
		server.handleResponse(writer, http.StatusMethodNotAllowed)
		return
	} else if utils.RequestGetContentType(request) != "application/json" {
		server.handleResponse(writer, http.StatusUnsupportedMediaType)
		return
	}

	// Parse input JSON for graph data
	body, _ = ioutil.ReadAll(request.Body)

	if err = json.Unmarshal(body, &plotReq); err != nil {
		log.Println("ERROR: " + err.Error())
		server.handleResponse(writer, http.StatusBadRequest)
		return
	}

	if plotReq.Origin != "" && plotReq.Template != "" {
		plotReq.Graph = plotReq.Origin + "\x30" + plotReq.Template
	}

	if plotReq.Time == "" {
		refTime = time.Now()
	} else if refTime, err = time.Parse(time.RFC3339, plotReq.Time); err != nil {
		log.Println("ERROR: " + err.Error())
		server.handleResponse(writer, http.StatusBadRequest)
		return
	}

	if plotReq.Sample == 0 {
		plotReq.Sample = defaultPlotSample
	}

	// Get graph from library
	if plotReq.Template != "" {
		graph, err = server.Library.GetGraphTemplate(plotReq.Origin, plotReq.Source, plotReq.Template, plotReq.Filter)
	} else {
		item, err = server.Library.GetItem(plotReq.Graph, library.LibraryItemGraph)
		graph = item.(*library.Graph)
	}

	if err != nil {
		log.Println("ERROR: " + err.Error())

		if os.IsNotExist(err) {
			server.handleResponse(writer, http.StatusNotFound)
		} else {
			server.handleResponse(writer, http.StatusInternalServerError)
		}

		return
	}

	// Get plots data
	response = make(map[string]map[string]common.PlotValue)
	values = make(map[string]map[string]common.PlotValue)

	for _, stackItem := range graph.Stacks {
		for _, groupItem := range stackItem.Groups {
			if query, originBackend, err = server.plotPrepareQuery(plotReq, groupItem); err != nil {
				log.Println("ERROR: " + err.Error())
				server.handleResponse(writer, http.StatusBadRequest)
				return
			}

			if values, err = originBackend.GetValue(query, refTime, plotReq.Percentiles); err != nil {
				log.Println("ERROR: " + err.Error())
				server.handleResponse(writer, http.StatusInternalServerError)
				return
			}

			for key, value := range values {
				response[key] = value
			}
		}
	}

	server.handleJSON(writer, response)
}

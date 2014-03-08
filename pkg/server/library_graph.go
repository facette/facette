package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/facette/facette/pkg/config"
	"github.com/facette/facette/pkg/connector"
	"github.com/facette/facette/pkg/library"
	"github.com/facette/facette/pkg/utils"
	"github.com/facette/facette/thirdparty/github.com/fatih/set"
)

func (server *Server) handleGraph(writer http.ResponseWriter, request *http.Request) {
	graphID := strings.TrimPrefix(request.URL.Path, urlLibraryPath+"graphs/")

	switch request.Method {
	case "DELETE":
		if graphID == "" {
			server.handleResponse(writer, serverResponse{mesgMethodNotAllowed}, http.StatusMethodNotAllowed)
			return
		} else if !server.handleAuth(writer, request) {
			server.handleResponse(writer, serverResponse{mesgAuthenticationRequired}, http.StatusUnauthorized)
			return
		}

		err := server.Library.DeleteItem(graphID, library.LibraryItemGraph)
		if os.IsNotExist(err) {
			server.handleResponse(writer, serverResponse{mesgResourceNotFound}, http.StatusNotFound)
			return
		} else if err != nil {
			log.Println("ERROR: " + err.Error())
			server.handleResponse(writer, serverResponse{mesgUnhandledError}, http.StatusInternalServerError)
			return
		}

		server.handleResponse(writer, nil, http.StatusOK)

		break

	case "GET", "HEAD":
		if graphID == "" {
			server.handleGraphList(writer, request)
			return
		}

		item, err := server.Library.GetItem(graphID, library.LibraryItemGraph)
		if os.IsNotExist(err) {
			server.handleResponse(writer, serverResponse{mesgResourceNotFound}, http.StatusNotFound)
			return
		} else if err != nil {
			log.Println("ERROR: " + err.Error())
			server.handleResponse(writer, serverResponse{mesgUnhandledError}, http.StatusInternalServerError)
			return
		}

		server.handleResponse(writer, item, http.StatusOK)

		break

	case "POST", "PUT":
		var graph *library.Graph

		if response, status := server.parseStoreRequest(writer, request, graphID); status != http.StatusOK {
			server.handleResponse(writer, response, status)
			return
		}

		if request.Method == "POST" && request.FormValue("inherit") != "" {
			// Get graph from library
			item, err := server.Library.GetItem(request.FormValue("inherit"), library.LibraryItemGraph)
			if os.IsNotExist(err) {
				server.handleResponse(writer, serverResponse{mesgResourceNotFound}, http.StatusNotFound)
				return
			} else if err != nil {
				log.Println("ERROR: " + err.Error())
				server.handleResponse(writer, serverResponse{mesgUnhandledError}, http.StatusInternalServerError)
				return
			}

			graph = &library.Graph{}
			*graph = *item.(*library.Graph)

			graph.ID = ""
		} else {
			// Create a new graph instance
			graph = &library.Graph{Item: library.Item{ID: graphID}}
		}

		graph.Modified = time.Now()

		// Parse input JSON for graph data
		body, _ := ioutil.ReadAll(request.Body)

		if err := json.Unmarshal(body, graph); err != nil {
			log.Println("ERROR: " + err.Error())
			server.handleResponse(writer, serverResponse{mesgResourceInvalid}, http.StatusBadRequest)
			return
		}

		// Store graph data
		if request.FormValue("volatile") != "" {
			graph.Volatile = true
		} else {
			graph.Volatile = false
		}

		err := server.Library.StoreItem(graph, library.LibraryItemGraph)
		if response, status := server.parseError(writer, request, err); status != http.StatusOK {
			log.Println("ERROR: " + err.Error())
			server.handleResponse(writer, response, status)
			return
		}

		if request.Method == "POST" {
			writer.Header().Add("Location", strings.TrimRight(request.URL.Path, "/")+"/"+graph.ID)
			server.handleResponse(writer, nil, http.StatusCreated)
		} else {
			server.handleResponse(writer, nil, http.StatusOK)
		}

		break

	default:
		server.handleResponse(writer, serverResponse{mesgMethodNotAllowed}, http.StatusMethodNotAllowed)
	}
}

func (server *Server) handleGraphList(writer http.ResponseWriter, request *http.Request) {
	var offset, limit int

	if response, status := server.parseListRequest(writer, request, &offset, &limit); status != http.StatusOK {
		server.handleResponse(writer, response, status)
		return
	}

	graphSet := set.New()

	// Filter on collection if any
	if request.FormValue("collection") != "" {
		item, err := server.Library.GetItem(request.FormValue("collection"), library.LibraryItemCollection)
		if os.IsNotExist(err) {
			server.handleResponse(writer, serverResponse{mesgResourceNotFound}, http.StatusNotFound)
			return
		} else if err != nil {
			log.Println("ERROR: " + err.Error())
			server.handleResponse(writer, serverResponse{mesgUnhandledError}, http.StatusInternalServerError)
			return
		}

		collection := item.(*library.Collection)

		for _, graph := range collection.Entries {
			graphSet.Add(graph.ID)
		}
	}

	// Fill graphs list
	items := make(ItemListResponse, 0)

	for _, graph := range server.Library.Graphs {
		if graph.Volatile || !graphSet.IsEmpty() && !graphSet.Has(graph.ID) {
			continue
		}

		if request.FormValue("filter") != "" && !utils.FilterMatch(request.FormValue("filter"), graph.Name) {
			continue
		}

		items = append(items, &ItemResponse{
			ID:          graph.ID,
			Name:        graph.Name,
			Description: graph.Description,
			Modified:    graph.Modified.Format(time.RFC3339),
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

func (server *Server) handleGraphPlots(writer http.ResponseWriter, request *http.Request) {
	var (
		graph              *library.Graph
		err                error
		startTime, endTime time.Time
	)

	if request.Method != "POST" && request.Method != "HEAD" {
		server.handleResponse(writer, serverResponse{mesgMethodNotAllowed}, http.StatusMethodNotAllowed)
		return
	} else if utils.RequestGetContentType(request) != "application/json" {
		server.handleResponse(writer, serverResponse{mesgUnsupportedMediaType}, http.StatusUnsupportedMediaType)
		return
	}

	// Parse input JSON for graph data
	body, _ := ioutil.ReadAll(request.Body)

	plotReq := &PlotRequest{}

	if err := json.Unmarshal(body, plotReq); err != nil {
		log.Println("ERROR: " + err.Error())
		server.handleResponse(writer, serverResponse{mesgResourceInvalid}, http.StatusBadRequest)
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
			server.handleResponse(writer, serverResponse{mesgResourceInvalid}, http.StatusBadRequest)
			return
		}
	} else {
		if startTime, err = time.Parse(time.RFC3339, plotReq.Time); err != nil {
			log.Println("ERROR: " + err.Error())
			server.handleResponse(writer, serverResponse{mesgResourceInvalid}, http.StatusBadRequest)
			return
		}
	}

	if startTime.IsZero() {
		if startTime, err = utils.TimeApplyRange(endTime, plotReq.Range); err != nil {
			log.Println("ERROR: " + err.Error())
			server.handleResponse(writer, serverResponse{mesgResourceInvalid}, http.StatusBadRequest)
			return
		}
	} else if endTime, err = utils.TimeApplyRange(startTime, plotReq.Range); err != nil {
		log.Println("ERROR: " + err.Error())
		server.handleResponse(writer, serverResponse{mesgResourceInvalid}, http.StatusBadRequest)
		return
	}

	if plotReq.Sample == 0 {
		plotReq.Sample = config.DefaultPlotSample
	}

	// Get graph from library
	if plotReq.Template != "" {
		graph, err = server.Library.GetGraphTemplate(
			plotReq.Origin,
			plotReq.Source,
			plotReq.Template,
			plotReq.Filter,
		)
	} else if plotReq.Metric != "" {
		graph, err = server.Library.GetGraphMetric(
			plotReq.Origin,
			plotReq.Source,
			plotReq.Metric,
		)
	} else if item, err := server.Library.GetItem(plotReq.Graph, library.LibraryItemGraph); err == nil {
		graph = item.(*library.Graph)
	}

	if err != nil {
		log.Println("ERROR: " + err.Error())

		if os.IsNotExist(err) {
			server.handleResponse(writer, serverResponse{mesgResourceNotFound}, http.StatusNotFound)
		} else {
			server.handleResponse(writer, serverResponse{mesgUnhandledError}, http.StatusInternalServerError)
		}

		return
	}

	step := endTime.Sub(startTime) / time.Duration(plotReq.Sample)

	// Get plots data
	groupOptions := make(map[string]map[string]interface{})

	data := make([]map[string]*connector.PlotResult, 0)

	for _, stackItem := range graph.Stacks {
		for _, groupItem := range stackItem.Groups {
			query, originConnector, err := server.preparePlotQuery(plotReq, groupItem)
			if err != nil {
				log.Println("ERROR: " + err.Error())
				server.handleResponse(writer, serverResponse{mesgUnhandledError}, http.StatusInternalServerError)
				return
			}

			groupOptions[groupItem.Name] = groupItem.Options

			plotResult, err := originConnector.GetPlots(query, startTime, endTime, step, plotReq.Percentiles)
			if err != nil {
				log.Println("ERROR: " + err.Error())
				server.handleResponse(writer, serverResponse{mesgUnhandledError}, http.StatusInternalServerError)
				return
			}

			data = append(data, plotResult)
		}
	}

	response := &PlotResponse{
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
		server.handleResponse(writer, serverResponse{mesgEmptyData}, http.StatusOK)
		return
	}

	plotMax := 0

	for _, stackItem := range graph.Stacks {
		stack := &StackResponse{Name: stackItem.Name}

		for _, groupItem := range stackItem.Groups {
			var plotResult map[string]*connector.PlotResult

			plotResult, data = data[0], data[1:]

			for serieName, serieResult := range plotResult {
				if len(serieResult.Plots) > plotMax {
					plotMax = len(serieResult.Plots)
				}

				stack.Series = append(stack.Series, &SerieResponse{
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

	server.handleResponse(writer, response, http.StatusOK)
}

func (server *Server) preparePlotQuery(plotReq *PlotRequest, groupItem *library.OperGroup) (*connector.GroupQuery,
	connector.Connector, error) {

	var originConnector connector.Connector

	query := &connector.GroupQuery{
		Name:  groupItem.Name,
		Type:  groupItem.Type,
		Scale: groupItem.Scale,
	}

	originConnector = nil

	for _, serieItem := range groupItem.Series {
		// Check for connectors errors or conflicts
		if _, ok := server.Catalog.Origins[serieItem.Origin]; !ok {
			return nil, nil, fmt.Errorf("unknown `%s' serie origin", serieItem.Origin)
		} else if originConnector == nil {
			originConnector = server.Catalog.Origins[serieItem.Origin].Connector
		} else if originConnector != server.Catalog.Origins[serieItem.Origin].Connector {
			return nil, nil, fmt.Errorf("connectors differ between series")
		}

		serieSources := make([]string, 0)

		if plotReq.Template != "" {
			serieSources = []string{plotReq.Source}
		} else if strings.HasPrefix(serieItem.Source, library.LibraryGroupPrefix) {
			serieSources = server.Library.ExpandGroup(strings.TrimPrefix(serieItem.Source, library.LibraryGroupPrefix),
				library.LibraryItemSourceGroup)
		} else {
			serieSources = []string{serieItem.Source}
		}

		index := 0

		for _, serieSource := range serieSources {
			if strings.HasPrefix(serieItem.Metric, library.LibraryGroupPrefix) {
				for _, serieChunk := range server.Library.ExpandGroup(strings.TrimPrefix(serieItem.Metric,
					library.LibraryGroupPrefix), library.LibraryItemMetricGroup) {
					metric := server.Catalog.GetMetric(
						serieItem.Origin,
						serieSource,
						serieChunk,
					)

					if metric == nil {
						log.Printf("unknown `%s' metric for source `%s' (origin: %s)", serieChunk, serieSource,
							serieItem.Origin)
					}

					query.Series = append(query.Series, &connector.SerieQuery{
						Name: fmt.Sprintf("%s-%d", serieItem.Name, index),
						Metric: &connector.MetricQuery{
							Name:       metric.OriginalName,
							SourceName: metric.Source.OriginalName,
						},
						Scale: serieItem.Scale,
					})

					index += 1
				}
			} else {
				metric := server.Catalog.GetMetric(
					serieItem.Origin,
					serieSource,
					serieItem.Metric,
				)

				if metric == nil {
					log.Printf("unknown `%s' metric for source `%s' (origin: %s)", serieItem.Metric, serieSource,
						serieItem.Origin)
				}

				serie := &connector.SerieQuery{
					Metric: &connector.MetricQuery{
						Name:       metric.OriginalName,
						SourceName: metric.Source.OriginalName,
					},
					Scale: serieItem.Scale,
				}

				if len(serieSources) > 1 {
					serie.Name = fmt.Sprintf("%s-%d", serieItem.Name, index)
				} else {
					serie.Name = serieItem.Name
				}

				query.Series = append(query.Series, serie)

				index += 1
			}
		}
	}

	if len(query.Series) == 0 {
		return nil, nil, fmt.Errorf("no serie defined")
	}

	return query, originConnector, nil
}

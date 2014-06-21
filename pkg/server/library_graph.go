package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/facette/facette/pkg/config"
	"github.com/facette/facette/pkg/connector"
	"github.com/facette/facette/pkg/library"
	"github.com/facette/facette/pkg/logger"
	"github.com/facette/facette/pkg/types"
	"github.com/facette/facette/pkg/utils"
	"github.com/facette/facette/thirdparty/github.com/fatih/set"
)

func (server *Server) serveGraph(writer http.ResponseWriter, request *http.Request) {
	graphID := strings.TrimPrefix(request.URL.Path, urlLibraryPath+"graphs/")

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
		var graph *library.Graph

		if response, status := server.parseStoreRequest(writer, request, graphID); status != http.StatusOK {
			server.serveResponse(writer, response, status)
			return
		}

		if request.Method == "POST" && request.FormValue("inherit") != "" {
			// Get graph from library
			item, err := server.Library.GetItem(request.FormValue("inherit"), library.LibraryItemGraph)
			if os.IsNotExist(err) {
				server.serveResponse(writer, serverResponse{mesgResourceNotFound}, http.StatusNotFound)
				return
			} else if err != nil {
				logger.Log(logger.LevelError, "server", "%s", err)
				server.serveResponse(writer, serverResponse{mesgUnhandledError}, http.StatusInternalServerError)
				return
			}

			graph = &library.Graph{}
			utils.Clone(item.(*library.Graph), graph)

			graph.ID = ""
		} else {
			// Create a new graph instance
			graph = &library.Graph{Item: library.Item{ID: graphID}}
		}

		graph.Modified = time.Now()

		// Parse input JSON for graph data
		body, _ := ioutil.ReadAll(request.Body)

		if err := json.Unmarshal(body, graph); err != nil {
			logger.Log(logger.LevelError, "server", "%s", err)
			server.serveResponse(writer, serverResponse{mesgResourceInvalid}, http.StatusBadRequest)
			return
		}

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
	var offset, limit int

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
	items := make(ItemListResponse, 0)

	for _, graph := range server.Library.Graphs {
		if !graphSet.IsEmpty() && !graphSet.Has(graph.ID) {
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

	server.serveResponse(writer, response.list, http.StatusOK)
}

func (server *Server) serveGraphPlots(writer http.ResponseWriter, request *http.Request) {
	var (
		err                error
		graph              *library.Graph
		item               interface{}
		startTime, endTime time.Time
	)

	if request.Method != "POST" && request.Method != "HEAD" {
		server.serveResponse(writer, serverResponse{mesgMethodNotAllowed}, http.StatusMethodNotAllowed)
		return
	} else if utils.HTTPGetContentType(request) != "application/json" {
		server.serveResponse(writer, serverResponse{mesgUnsupportedMediaType}, http.StatusUnsupportedMediaType)
		return
	}

	// Parse input JSON for graph data
	body, _ := ioutil.ReadAll(request.Body)

	plotReq := PlotRequest{}

	if err := json.Unmarshal(body, &plotReq); err != nil {
		logger.Log(logger.LevelError, "server", "%s", err)
		server.serveResponse(writer, serverResponse{mesgResourceInvalid}, http.StatusBadRequest)
		return
	}

	if plotReq.Time == "" {
		endTime = time.Now()
	} else if strings.HasPrefix(strings.Trim(plotReq.Range, " "), "-") {
		if endTime, err = time.Parse(time.RFC3339, plotReq.Time); err != nil {
			logger.Log(logger.LevelError, "server", "%s", err)
			server.serveResponse(writer, serverResponse{mesgResourceInvalid}, http.StatusBadRequest)
			return
		}
	} else {
		if startTime, err = time.Parse(time.RFC3339, plotReq.Time); err != nil {
			logger.Log(logger.LevelError, "server", "%s", err)
			server.serveResponse(writer, serverResponse{mesgResourceInvalid}, http.StatusBadRequest)
			return
		}
	}

	if startTime.IsZero() {
		if startTime, err = utils.TimeApplyRange(endTime, plotReq.Range); err != nil {
			logger.Log(logger.LevelError, "server", "%s", err)
			server.serveResponse(writer, serverResponse{mesgResourceInvalid}, http.StatusBadRequest)
			return
		}
	} else if endTime, err = utils.TimeApplyRange(startTime, plotReq.Range); err != nil {
		logger.Log(logger.LevelError, "server", "%s", err)
		server.serveResponse(writer, serverResponse{mesgResourceInvalid}, http.StatusBadRequest)
		return
	}

	if plotReq.Sample == 0 {
		plotReq.Sample = config.DefaultPlotSample
	}

	// Get graph from library
	graph = plotReq.Graph

	if plotReq.ID != "" {
		if item, err = server.Library.GetItem(plotReq.ID, library.LibraryItemGraph); err == nil {
			graph = item.(*library.Graph)
		}
	}

	if graph == nil {
		err = os.ErrNotExist
	}

	if err != nil {
		if os.IsNotExist(err) {
			server.serveResponse(writer, serverResponse{mesgResourceNotFound}, http.StatusNotFound)
		} else {
			logger.Log(logger.LevelError, "server", "%s", err)
			server.serveResponse(writer, serverResponse{mesgUnhandledError}, http.StatusInternalServerError)
		}

		return
	}

	step := endTime.Sub(startTime) / time.Duration(plotReq.Sample)

	// Get plots data
	groupOptions := make(map[string]map[string]interface{})

	data := make([]map[string]*types.PlotResult, 0)

	for _, groupItem := range graph.Groups {
		groupOptions[groupItem.Name] = groupItem.Options

		query, providerConnector, err := server.preparePlotQuery(&plotReq, groupItem)
		if err != nil {
			if err != os.ErrInvalid {
				logger.Log(logger.LevelError, "server", "%s", err)
			}

			data = append(data, nil)
			continue
		}

		plotResult, err := providerConnector.GetPlots(&types.PlotQuery{query, startTime, endTime, step,
			plotReq.Percentiles})
		if err != nil {
			logger.Log(logger.LevelError, "server", "%s", err)
		}

		data = append(data, plotResult)
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
		server.serveResponse(writer, serverResponse{mesgEmptyData}, http.StatusOK)
		return
	}

	plotMax := 0

	for _, groupItem := range graph.Groups {
		var plotResult map[string]*types.PlotResult

		plotResult, data = data[0], data[1:]

		for serieName, serieResult := range plotResult {
			if len(serieResult.Plots) > plotMax {
				plotMax = len(serieResult.Plots)
			}

			response.Series = append(response.Series, &SerieResponse{
				Name:    serieName,
				Plots:   serieResult.Plots,
				Info:    serieResult.Info,
				Options: groupOptions[groupItem.Name],
			})
		}
	}

	if plotMax > 0 {
		response.Step = (endTime.Sub(startTime) / time.Duration(plotMax)).Seconds()
	}

	server.serveResponse(writer, response, http.StatusOK)
}

func (server *Server) preparePlotQuery(plotReq *PlotRequest, groupItem *library.OperGroup) (*types.GroupQuery,
	connector.Connector, error) {
	var providerConnector connector.Connector

	query := &types.GroupQuery{
		Name:  groupItem.Name,
		Type:  groupItem.Type,
		Scale: groupItem.Scale,
	}

	for _, serieItem := range groupItem.Series {
		// Check for connectors errors or conflicts
		if _, ok := server.Catalog.Origins[serieItem.Origin]; !ok {
			return nil, nil, fmt.Errorf("unknown serie origin `%s'", serieItem.Origin)
		}

		serieSources := make([]string, 0)

		if strings.HasPrefix(serieItem.Source, library.LibraryGroupPrefix) {
			serieSources = server.Library.ExpandGroup(
				strings.TrimPrefix(serieItem.Source, library.LibraryGroupPrefix),
				library.LibraryItemSourceGroup,
			)
		} else {
			serieSources = []string{serieItem.Source}
		}

		index := 0

		for _, serieSource := range serieSources {
			if strings.HasPrefix(serieItem.Metric, library.LibraryGroupPrefix) {
				for _, serieChunk := range server.Library.ExpandGroup(
					strings.TrimPrefix(serieItem.Metric, library.LibraryGroupPrefix),
					library.LibraryItemMetricGroup,
				) {
					metric := server.Catalog.GetMetric(serieItem.Origin, serieSource, serieChunk)

					if metric == nil {
						logger.Log(
							logger.LevelError,
							"server",
							"unknown metric `%s' for source `%s' (origin: %s)",
							serieChunk,
							serieSource,
							serieItem.Origin,
						)

						continue
					}

					if providerConnector == nil {
						providerConnector = metric.Connector.(connector.Connector)
					} else if providerConnector != metric.Connector.(connector.Connector) {
						return nil, nil, fmt.Errorf("connectors differ between series")
					}

					query.Series = append(query.Series, &types.SerieQuery{
						Name: fmt.Sprintf("%s-%d", serieItem.Name, index),
						Metric: &types.MetricQuery{
							Name:   metric.OriginalName,
							Origin: metric.Source.Origin.OriginalName,
							Source: metric.Source.OriginalName,
						},
						Scale: serieItem.Scale,
					})

					index += 1
				}
			} else {
				metric := server.Catalog.GetMetric(serieItem.Origin, serieSource, serieItem.Metric)

				if metric == nil {
					logger.Log(
						logger.LevelError,
						"server",
						"unknown metric `%s' for source `%s' (origin: %s)",
						serieItem.Metric,
						serieSource,
						serieItem.Origin,
					)

					continue
				}

				if providerConnector == nil {
					providerConnector = metric.Connector.(connector.Connector)
				} else if providerConnector != metric.Connector.(connector.Connector) {
					return nil, nil, fmt.Errorf("connectors differ between series")
				}

				serie := &types.SerieQuery{
					Metric: &types.MetricQuery{
						Name:   metric.OriginalName,
						Origin: metric.Source.Origin.OriginalName,
						Source: metric.Source.OriginalName,
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
		return nil, nil, os.ErrInvalid
	}

	return query, providerConnector, nil
}

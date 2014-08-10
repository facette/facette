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
	"github.com/facette/facette/pkg/plot"
	"github.com/facette/facette/pkg/utils"
	"github.com/facette/facette/thirdparty/github.com/fatih/set"
)

func (server *Server) serveGraph(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "GET" && request.Method != "HEAD" && server.Config.API.ReadOnly {
		server.serveResponse(writer, serverResponse{mesgReadOnlyMode}, http.StatusForbidden)
		return
	}

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
	var (
		items         ItemListResponse
		offset, limit int
	)

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
	items = make(ItemListResponse, 0)

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
		graphPlotSeries    [][]plot.Series
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

	// Parse input JSON for graph series
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

	// Get graph plots series
	groupOptions := make(map[string]map[string]interface{})

	for _, groupItem := range graph.Groups {
		groupOptions[groupItem.Name] = groupItem.Options

		query, providerConnector, err := server.prepareQuery(&plotReq, groupItem)
		if err != nil {
			if err != os.ErrInvalid {
				logger.Log(logger.LevelError, "server", "%s", err)
			}

			graphPlotSeries = append(graphPlotSeries, nil)
			continue
		}

		plotSeries, err := providerConnector.GetPlots(&plot.Query{query, startTime, endTime, plotReq.Sample})
		if err != nil {
			logger.Log(logger.LevelError, "server", "%s", err)
		}

		if len(plotSeries) > 1 {
			for index, entry := range plotSeries {
				entry.Name = fmt.Sprintf("%s (%s)", groupItem.Name, query.Series[index].Metric.Name)
				entry.Summarize(plotReq.Percentiles)
				entry.Downsample(plotReq.Sample, plot.ConsolidateAverage)
			}
		} else if len(plotSeries) == 1 {
			plotSeries[0].Name = groupItem.Name
			plotSeries[0].Summarize(plotReq.Percentiles)
			plotSeries[0].Downsample(plotReq.Sample, plot.ConsolidateAverage)
		}

		graphPlotSeries = append(graphPlotSeries, plotSeries)
	}

	response := &PlotResponse{
		ID:          graph.ID,
		Start:       startTime.Format(time.RFC3339),
		End:         endTime.Format(time.RFC3339),
		Step:        (endTime.Sub(startTime) / time.Duration(plotReq.Sample)).Seconds(),
		Name:        graph.Name,
		Description: graph.Description,
		Type:        graph.Type,
		StackMode:   graph.StackMode,
		UnitType:    graph.UnitType,
		UnitLegend:  graph.UnitLegend,
		Modified:    graph.Modified,
	}

	if len(graphPlotSeries) == 0 {
		server.serveResponse(writer, serverResponse{mesgEmptyData}, http.StatusOK)
		return
	}

	plotMax := 0

	for _, groupItem := range graph.Groups {
		var series []plot.Series

		series, graphPlotSeries = graphPlotSeries[0], graphPlotSeries[1:]

		for _, seriesResult := range series {
			if len(seriesResult.Plots) > plotMax {
				plotMax = len(seriesResult.Plots)
			}

			response.Series = append(response.Series, &SeriesResponse{
				Name:    seriesResult.Name,
				Plots:   seriesResult.Plots,
				Summary: seriesResult.Summary,
				Options: groupOptions[groupItem.Name],
			})
		}
	}

	if plotMax > 0 {
		response.Step = (endTime.Sub(startTime) / time.Duration(plotMax)).Seconds()
	}

	server.serveResponse(writer, response, http.StatusOK)
}

func (server *Server) prepareQuery(plotReq *PlotRequest, groupItem *library.OperGroup) (*plot.QueryGroup,
	connector.Connector, error) {

	var (
		providerConnector connector.Connector
		seriesSources     []string
	)

	query := &plot.QueryGroup{
		Type:    groupItem.Type,
		Options: groupItem.Options,
	}

	for _, seriesItem := range groupItem.Series {
		// Check for connectors errors or conflicts
		if _, ok := server.Catalog.Origins[seriesItem.Origin]; !ok {
			return nil, nil, fmt.Errorf("unknown series origin `%s'", seriesItem.Origin)
		}

		if strings.HasPrefix(seriesItem.Source, library.LibraryGroupPrefix) {
			seriesSources = server.Library.ExpandGroup(
				strings.TrimPrefix(seriesItem.Source, library.LibraryGroupPrefix),
				library.LibraryItemSourceGroup,
			)
		} else {
			seriesSources = []string{seriesItem.Source}
		}

		index := 0

		for _, seriesEntry := range seriesSources {
			if strings.HasPrefix(seriesItem.Metric, library.LibraryGroupPrefix) {
				for _, seriesChunk := range server.Library.ExpandGroup(
					strings.TrimPrefix(seriesItem.Metric, library.LibraryGroupPrefix),
					library.LibraryItemMetricGroup,
				) {
					metric := server.Catalog.GetMetric(seriesItem.Origin, seriesEntry, seriesChunk)

					if metric == nil {
						logger.Log(
							logger.LevelWarning,
							"server",
							"unknown metric `%s' for source `%s' (origin: %s)",
							seriesChunk,
							seriesEntry,
							seriesItem.Origin,
						)

						continue
					}

					if providerConnector == nil {
						providerConnector = metric.Connector.(connector.Connector)
					} else if providerConnector != metric.Connector.(connector.Connector) {
						return nil, nil, fmt.Errorf("connectors differ between series")
					}

					query.Series = append(query.Series, &plot.QuerySeries{
						Metric: &plot.QueryMetric{
							Name:   metric.OriginalName,
							Origin: metric.Source.Origin.OriginalName,
							Source: metric.Source.OriginalName,
						},
						Options: seriesItem.Options,
					})

					index++
				}
			} else {
				metric := server.Catalog.GetMetric(seriesItem.Origin, seriesEntry, seriesItem.Metric)

				if metric == nil {
					logger.Log(
						logger.LevelWarning,
						"server",
						"unknown metric `%s' for source `%s' (origin: %s)",
						seriesItem.Metric,
						seriesEntry,
						seriesItem.Origin,
					)

					continue
				}

				if providerConnector == nil {
					providerConnector = metric.Connector.(connector.Connector)
				} else if providerConnector != metric.Connector.(connector.Connector) {
					return nil, nil, fmt.Errorf("connectors differ between series")
				}

				query.Series = append(query.Series, &plot.QuerySeries{
					Metric: &plot.QueryMetric{
						Name:   metric.OriginalName,
						Origin: metric.Source.Origin.OriginalName,
						Source: metric.Source.OriginalName,
					},
					Options: seriesItem.Options,
				})

				index++
			}
		}
	}

	if len(query.Series) == 0 {
		return nil, nil, os.ErrInvalid
	}

	return query, providerConnector, nil
}

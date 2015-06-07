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
)

func (server *Server) servePlots(writer http.ResponseWriter, request *http.Request) {
	var (
		err   error
		graph *library.Graph
		item  interface{}
	)

	if request.Method != "POST" && request.Method != "HEAD" {
		server.serveResponse(writer, serverResponse{mesgMethodNotAllowed}, http.StatusMethodNotAllowed)
		return
	} else if utils.HTTPGetContentType(request) != "application/json" {
		server.serveResponse(writer, serverResponse{mesgUnsupportedMediaType}, http.StatusUnsupportedMediaType)
		return
	}

	// Check for requests loop
	if request.Header.Get("X-Facette-Requestor") == server.ID {
		logger.Log(logger.LevelWarning, "server", "request loop detected, cancelled")
		server.serveResponse(writer, serverResponse{mesgEmptyData}, http.StatusBadRequest)
		return
	}

	// Parse plots request
	plotReq, err := parsePlotRequest(request)
	if err != nil {
		logger.Log(logger.LevelError, "server", "%s", err)
		server.serveResponse(writer, serverResponse{mesgResourceInvalid}, http.StatusBadRequest)
		return
	}

	// Get graph definition from the plot request
	graph = plotReq.Graph

	// If a Graph ID has been provided in the plot request, fetch the graph definition from the library instead
	if plotReq.ID != "" {
		graph = &library.Graph{}
		if item, err = server.Library.GetItem(plotReq.ID, library.LibraryItemGraph); err == nil {
			utils.Clone(item.(*library.Graph), graph)
		}
	}

	if graph == nil {
		err = os.ErrNotExist
	}

	// Stop if an error was encountered
	if err != nil {
		if os.IsNotExist(err) {
			server.serveResponse(writer, serverResponse{mesgResourceNotFound}, http.StatusNotFound)
		} else {
			logger.Log(logger.LevelError, "server", "%s", err)
			server.serveResponse(writer, serverResponse{mesgUnhandledError}, http.StatusInternalServerError)
		}

		return
	}

	// If linked graph, expand its template
	if graph.Link != "" || graph.ID == "" && len(graph.Attributes) > 0 {
		if graph.Link != "" {
			// Get graph template from library
			item, err := server.Library.GetItem(graph.Link, library.LibraryItemGraph)
			if err != nil {
				logger.Log(logger.LevelError, "server", "graph template not found: %s", graph.Link)
				return
			}

			utils.Clone(item.(*library.Graph), graph)
		}

		if err = server.expandGraphTemplate(graph); err != nil {
			logger.Log(logger.LevelError, "server", "unable to apply graph template: %s", err)
			server.serveResponse(writer, serverResponse{mesgUnhandledError}, http.StatusInternalServerError)
		}
	}

	// Prepare queries to be executed by the providers
	providerQueries, err := server.prepareProviderQueries(plotReq, graph)
	if err != nil {
		if os.IsNotExist(err) {
			server.serveResponse(writer, serverResponse{mesgResourceNotFound}, http.StatusNotFound)
		} else {
			logger.Log(logger.LevelError, "server", "%s", err)
			server.serveResponse(writer, serverResponse{mesgUnhandledError}, http.StatusInternalServerError)
		}

		return
	}

	plotSeries, err := executeQueries(providerQueries)
	if err != nil {
		logger.Log(logger.LevelError, "server", "unable to execute provider queries: %s", err)
		server.serveResponse(writer, serverResponse{mesgProviderQueryError}, http.StatusInternalServerError)
		return
	}

	if len(plotSeries) == 0 {
		server.serveResponse(writer, serverResponse{mesgEmptyData}, http.StatusOK)
		return
	}

	response, err := makePlotsResponse(plotSeries, plotReq, graph)
	if err != nil {
		logger.Log(logger.LevelError, "server", "unable to make plots response: %s", err)
		server.serveResponse(writer, serverResponse{mesgPlotOperationError}, http.StatusInternalServerError)
		return
	}

	server.serveResponse(writer, response, http.StatusOK)
}

func (server *Server) expandGraphTemplate(graph *library.Graph) error {
	var err error

	if graph.Title, err = expandStringTemplate(graph.Title, graph.Attributes); err != nil {
		return fmt.Errorf("failed to expand graph title: %s", err)
	}

	for _, group := range graph.Groups {
		for _, series := range group.Series {
			if series.Name, err = expandStringTemplate(series.Name, graph.Attributes); err != nil {
				return fmt.Errorf("failed to expand graph series %q name: %s", series.Name, err)
			}

			if series.Source, err = expandStringTemplate(series.Source, graph.Attributes); err != nil {
				return fmt.Errorf("failed to expand graph series %q source: %s", series.Name, err)
			}

			if series.Metric, err = expandStringTemplate(series.Metric, graph.Attributes); err != nil {
				return fmt.Errorf("failed to expand graph series %q metric: %s", series.Name, err)
			}
		}
	}

	return nil
}

func (server *Server) prepareProviderQueries(plotReq *PlotRequest,
	graph *library.Graph) (map[string]*providerQuery, error) {

	providerQueries := make(map[string]*providerQuery)

	for _, groupItem := range graph.Groups {
		for _, seriesItem := range groupItem.Series {
			var seriesSources []string

			if seriesItem == nil {
				return nil, os.ErrNotExist
			}

			// Expand source groups
			if strings.HasPrefix(seriesItem.Source, library.LibraryGroupPrefix) {
				seriesSources = server.Library.ExpandSourceGroup(
					strings.TrimPrefix(seriesItem.Source, library.LibraryGroupPrefix),
				)
			} else {
				seriesSources = []string{seriesItem.Source}
			}

			// Process series metrics
			for _, sourceItem := range seriesSources {
				var seriesMetrics []string

				// Expand metric groups
				if strings.HasPrefix(seriesItem.Metric, library.LibraryGroupPrefix) {
					seriesMetrics = server.Library.ExpandMetricGroup(
						sourceItem,
						strings.TrimPrefix(seriesItem.Metric, library.LibraryGroupPrefix),
					)
				} else {
					seriesMetrics = []string{seriesItem.Metric}
				}

				for _, metricItem := range seriesMetrics {
					// Get series metric
					metric, err := server.Catalog.GetMetric(seriesItem.Origin, sourceItem, metricItem)
					if err != nil {
						logger.Log(logger.LevelWarning, "server", "%s", err)
						continue
					}

					// Get provider name
					providerName := metric.GetConnector().(connector.Connector).GetName()

					// Initialize provider query if needed
					if _, ok := providerQueries[providerName]; !ok {
						providerQueries[providerName] = &providerQuery{
							query: plot.Query{
								Requestor: plotReq.requestor,
								StartTime: plotReq.startTime,
								EndTime:   plotReq.endTime,
								Sample:    plotReq.Sample,
								Series:    make([]plot.QuerySeries, 0),
							},
							queryMap:  make([]providerQueryMap, 0),
							connector: metric.GetConnector().(connector.Connector),
						}
					}

					// Append metric to provider query
					providerQueries[providerName].query.Series = append(
						providerQueries[providerName].query.Series,
						plot.QuerySeries{
							Name:   fmt.Sprintf("series%d", len(providerQueries[providerName].query.Series)),
							Origin: metric.GetSource().GetOrigin().OriginalName,
							Source: metric.GetSource().OriginalName,
							Metric: metric.OriginalName,
						},
					)

					// Keep track of user-defined series name and source/metric information
					providerQueries[providerName].queryMap = append(
						providerQueries[providerName].queryMap,
						providerQueryMap{
							seriesName:      seriesItem.Name,
							sourceName:      metric.GetSource().Name,
							metricName:      metric.Name,
							fromSourceGroup: strings.HasPrefix(seriesItem.Source, library.LibraryGroupPrefix),
						},
					)
				}
			}
		}
	}

	return providerQueries, nil
}

func parsePlotRequest(request *http.Request) (*PlotRequest, error) {
	var err error

	plotReq := &PlotRequest{}

	// Parse input JSON for plots request
	body, _ := ioutil.ReadAll(request.Body)
	if err = json.Unmarshal(body, plotReq); err != nil {
		return nil, err
	}

	// Check plots request parameters
	if plotReq.Time.IsZero() {
		plotReq.endTime = time.Now()
	} else if strings.HasPrefix(strings.Trim(plotReq.Range, " "), "-") {
		plotReq.endTime = plotReq.Time
	} else {
		plotReq.startTime = plotReq.Time
	}

	if plotReq.startTime.IsZero() {
		if plotReq.startTime, err = utils.TimeApplyRange(plotReq.endTime, plotReq.Range); err != nil {
			return nil, err
		}
	} else if plotReq.endTime, err = utils.TimeApplyRange(plotReq.startTime, plotReq.Range); err != nil {
		return nil, err
	}

	if plotReq.Sample == 0 {
		plotReq.Sample = config.DefaultPlotSample
	}

	// Append plot requestor identifier
	plotReq.requestor = request.Header.Get("X-Facette-Requestor")

	return plotReq, nil
}

func executeQueries(queries map[string]*providerQuery) (map[string][]plot.Series, error) {
	plotSeries := make(map[string][]plot.Series)

	for _, providerQuery := range queries {
		plots, err := providerQuery.connector.GetPlots(&providerQuery.query)
		if err != nil {
			logger.Log(logger.LevelError, "server", "%s", err)
			continue
		}

		// Re-arrange internal plot results according to original queries
		for plotsIndex, plotsItem := range plots {
			// Add metric name detail to series name is a source/metric group
			if providerQuery.queryMap[plotsIndex].fromSourceGroup ||
				strings.HasPrefix(providerQuery.queryMap[plotsIndex].seriesName, library.LibraryGroupPrefix) {

				plotsItem.Name = fmt.Sprintf(
					"%s (%s)",
					providerQuery.queryMap[plotsIndex].sourceName,
					providerQuery.queryMap[plotsIndex].metricName,
				)
			} else {
				plotsItem.Name = providerQuery.queryMap[plotsIndex].seriesName
			}

			if _, ok := plotSeries[providerQuery.queryMap[plotsIndex].seriesName]; !ok {
				plotSeries[providerQuery.queryMap[plotsIndex].seriesName] = make([]plot.Series, 0)
			}

			plotSeries[providerQuery.queryMap[plotsIndex].seriesName] = append(
				plotSeries[providerQuery.queryMap[plotsIndex].seriesName],
				plotsItem,
			)
		}
	}

	return plotSeries, nil
}

func makePlotsResponse(plotSeries map[string][]plot.Series, plotReq *PlotRequest,
	graph *library.Graph) (*PlotResponse, error) {

	response := &PlotResponse{
		ID:          graph.ID,
		Start:       plotReq.startTime.Format(time.RFC3339),
		End:         plotReq.endTime.Format(time.RFC3339),
		Name:        graph.Name,
		Description: graph.Description,
		Title:       graph.Title,
		Type:        graph.Type,
		StackMode:   graph.StackMode,
		UnitType:    graph.UnitType,
		UnitLegend:  graph.UnitLegend,
		Modified:    graph.Modified,
	}

	for _, groupItem := range graph.Groups {
		var (
			groupConsolidate int
			groupSeries      []plot.Series
			err              error
		)

		seriesOptions := make(map[string]map[string]interface{})

		for _, seriesItem := range groupItem.Series {
			if _, ok := plotSeries[seriesItem.Name]; !ok {
				return nil, fmt.Errorf("unable to find plots for `%s' series", seriesItem.Name)
			}

			for _, plotItem := range plotSeries[seriesItem.Name] {
				var optionKey string

				// Apply series scale if any
				if scale, _ := config.GetFloat(seriesItem.Options, "scale", false); scale != 0 {
					plotItem.Scale(plot.Value(scale))
				}

				// Merge options from group and series
				if groupItem.Type == plot.OperTypeAverage || groupItem.Type == plot.OperTypeSum {
					optionKey = groupItem.Name
				} else {
					optionKey = seriesItem.Name
				}

				seriesOptions[optionKey] = make(map[string]interface{})
				for key, value := range groupItem.Options {
					seriesOptions[optionKey][key] = value
				}
				if groupItem.Type != plot.OperTypeAverage && groupItem.Type != plot.OperTypeSum {
					for key, value := range seriesItem.Options {
						seriesOptions[optionKey][key] = value
					}
				}

				groupSeries = append(groupSeries, plotItem)
			}
		}

		if len(groupSeries) == 0 {
			continue
		}

		// Normalize all series plots on the same time step
		groupConsolidate, err = config.GetInt(groupItem.Options, "consolidate", true)
		if err != nil {
			groupConsolidate = plot.ConsolidateAverage
		}

		groupSeries, err = plot.Normalize(
			groupSeries,
			plotReq.startTime,
			plotReq.endTime,
			plotReq.Sample,
			groupConsolidate,
		)
		if err != nil {
			return nil, fmt.Errorf("unable to consolidate series: %s", err)
		}

		// Perform requested series operations
		if groupItem.Type == plot.OperTypeAverage || groupItem.Type == plot.OperTypeSum {
			var (
				operSeries plot.Series
				err        error
			)

			if groupItem.Type == plot.OperTypeAverage {
				operSeries, err = plot.AverageSeries(groupSeries)
				if err != nil {
					return nil, fmt.Errorf("unable to average series: %s", err)
				}
			} else {
				operSeries, err = plot.SumSeries(groupSeries)
				if err != nil {
					return nil, fmt.Errorf("unable to sum series: %s", err)
				}
			}

			operSeries.Name = groupItem.Name

			groupSeries = []plot.Series{operSeries}
		}

		// Apply group scale if any
		if scale, _ := config.GetFloat(groupItem.Options, "scale", false); scale != 0 {
			for _, seriesItem := range groupSeries {
				seriesItem.Scale(plot.Value(scale))
			}
		}

		for _, seriesItem := range groupSeries {
			// Summarize each series (compute min/max/avg/last values)
			seriesItem.Summarize(plotReq.Percentiles)

			response.Series = append(response.Series, &SeriesResponse{
				Name:    seriesItem.Name,
				StackID: groupItem.StackID,
				Plots:   seriesItem.Plots,
				Summary: seriesItem.Summary,
				Options: seriesOptions[seriesItem.Name],
			})
		}
	}

	return response, nil
}

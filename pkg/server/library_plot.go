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

	"github.com/facette/facette/pkg/backend"
	"github.com/facette/facette/pkg/library"
	"github.com/facette/facette/pkg/types"
	"github.com/facette/facette/pkg/utils"
)

const (
	defaultPlotSample = 400
)

func (server *Server) plotHandle(writer http.ResponseWriter, request *http.Request) {
	var (
		graph     *library.Graph
		endTime   time.Time
		err       error
		startTime time.Time
	)

	if request.Method != "POST" && request.Method != "HEAD" {
		server.handleResponse(writer, http.StatusMethodNotAllowed)
		return
	} else if utils.RequestGetContentType(request) != "application/json" {
		server.handleResponse(writer, http.StatusUnsupportedMediaType)
		return
	}

	// Parse input JSON for graph data
	body, _ := ioutil.ReadAll(request.Body)

	plotReq := &types.PlotRequest{}

	if err := json.Unmarshal(body, plotReq); err != nil {
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
	} else if item, err := server.Library.GetItem(plotReq.Graph, library.LibraryItemGraph); err == nil {
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

	step := endTime.Sub(startTime) / time.Duration(plotReq.Sample)

	// Get plots data
	groupOptions := make(map[string]map[string]interface{})

	data := []map[string]*backend.PlotResult{}

	for _, stackItem := range graph.Stacks {
		for _, groupItem := range stackItem.Groups {
			query, originBackend, err := server.plotPrepareQuery(plotReq, groupItem)
			if err != nil {
				log.Println("ERROR: " + err.Error())
				server.handleResponse(writer, http.StatusBadRequest)
				return
			}

			groupOptions[groupItem.Name] = groupItem.Options

			plotResult, err := originBackend.GetPlots(query, startTime, endTime, step, plotReq.Percentiles)
			if err != nil {
				log.Println("ERROR: " + err.Error())
				server.handleResponse(writer, http.StatusInternalServerError)
				return
			}

			data = append(data, plotResult)
		}
	}

	response := &types.PlotResponse{
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

	plotMax := 0

	for _, stackItem := range graph.Stacks {
		stack := &types.StackResponse{Name: stackItem.Name}

		for _, groupItem := range stackItem.Groups {
			var plotResult map[string]*backend.PlotResult

			plotResult, data = data[0], data[1:]

			for serieName, serieResult := range plotResult {
				if len(serieResult.Plots) > plotMax {
					plotMax = len(serieResult.Plots)
				}

				stack.Series = append(stack.Series, &types.SerieResponse{
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

func (server *Server) plotPrepareQuery(plotReq *types.PlotRequest,
	groupItem *library.OperGroup) (*backend.GroupQuery, backend.BackendHandler, error) {

	var originBackend backend.BackendHandler

	query := &backend.GroupQuery{
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

		serieSources := []string{}

		if plotReq.Template != "" {
			serieSources = []string{plotReq.Source}
		} else if strings.HasPrefix(serieItem.Source, "group:") {
			serieSources = server.Library.ExpandGroup(serieItem.Source[6:], library.LibraryItemSourceGroup)
		} else {
			serieSources = []string{serieItem.Source}
		}

		index := 0

		for _, serieSource := range serieSources {
			if strings.HasPrefix(serieItem.Metric, "group:") {
				for _, serieChunk := range server.Library.ExpandGroup(serieItem.Metric[6:],
					library.LibraryItemMetricGroup) {
					metric := server.Catalog.GetMetric(
						serieItem.Origin,
						serieSource,
						serieChunk,
					)

					if metric == nil {
						log.Printf("unknown `%s' metric for source `%s' (origin: %s)", serieChunk, serieSource,
							serieItem.Origin)
					}

					query.Series = append(query.Series, &backend.SerieQuery{
						Name:   fmt.Sprintf("%s-%d", serieItem.Name, index),
						Metric: metric,
						Scale:  serieItem.Scale,
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

				serie := &backend.SerieQuery{
					Metric: metric,
					Scale:  serieItem.Scale,
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

	return query, originBackend, nil
}

func (server *Server) plotValues(writer http.ResponseWriter, request *http.Request) {
	var (
		err     error
		graph   *library.Graph
		item    interface{}
		refTime time.Time
	)

	if request.Method != "POST" && request.Method != "HEAD" {
		server.handleResponse(writer, http.StatusMethodNotAllowed)
		return
	} else if utils.RequestGetContentType(request) != "application/json" {
		server.handleResponse(writer, http.StatusUnsupportedMediaType)
		return
	}

	// Parse input JSON for graph data
	body, _ := ioutil.ReadAll(request.Body)

	plotReq := &types.PlotRequest{}

	if err := json.Unmarshal(body, &plotReq); err != nil {
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
	response := make(map[string]map[string]types.PlotValue)

	for _, stackItem := range graph.Stacks {
		for _, groupItem := range stackItem.Groups {
			query, originBackend, err := server.plotPrepareQuery(plotReq, groupItem)
			if err != nil {
				log.Println("ERROR: " + err.Error())
				server.handleResponse(writer, http.StatusBadRequest)
				return
			}

			values, err := originBackend.GetValue(query, refTime, plotReq.Percentiles)
			if err != nil {
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

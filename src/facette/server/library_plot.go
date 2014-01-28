package server

import (
	"encoding/json"
	"facette/backend"
	"facette/common"
	"facette/library"
	"facette/utils"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	defaultPlotSample = 400
)

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

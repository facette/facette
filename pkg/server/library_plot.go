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

	"github.com/facette/facette/pkg/connector"
	"github.com/facette/facette/pkg/library"
	"github.com/facette/facette/pkg/types"
	"github.com/facette/facette/pkg/utils"
)

const (
	defaultPlotSample = 400
)

// PlotRequest represents a eplot request struct in the server library.
type PlotRequest struct {
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

// SerieResponse represents a serie response struct in the server library.
type SerieResponse struct {
	Name    string                     `json:"name"`
	Plots   []types.PlotValue          `json:"plots"`
	Info    map[string]types.PlotValue `json:"info"`
	Options map[string]interface{}     `json:"options"`
}

// StackResponse represents a stack response struct in the server library.
type StackResponse struct {
	Name   string           `json:"name"`
	Series []*SerieResponse `json:"series"`
}

// PlotResponse represents a plot response struct in the server library.
type PlotResponse struct {
	ID          string           `json:"id"`
	Start       string           `json:"start"`
	End         string           `json:"end"`
	Step        float64          `json:"step"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Type        int              `json:"type"`
	StackMode   int              `json:"stack_mode"`
	Stacks      []*StackResponse `json:"stacks"`
	Modified    time.Time        `json:"modified"`
}

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

	plotReq := &PlotRequest{}

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

	data := []map[string]*connector.PlotResult{}

	for _, stackItem := range graph.Stacks {
		for _, groupItem := range stackItem.Groups {
			query, originConnector, err := server.plotPrepareQuery(plotReq, groupItem)
			if err != nil {
				log.Println("ERROR: " + err.Error())
				server.handleResponse(writer, http.StatusBadRequest)
				return
			}

			groupOptions[groupItem.Name] = groupItem.Options

			plotResult, err := originConnector.GetPlots(query, startTime, endTime, step, plotReq.Percentiles)
			if err != nil {
				log.Println("ERROR: " + err.Error())
				server.handleResponse(writer, http.StatusInternalServerError)
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
		server.handleJSON(writer, statusResponse{"No data"})
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

	server.handleJSON(writer, response)
}

func (server *Server) plotPrepareQuery(plotReq *PlotRequest, groupItem *library.OperGroup) (*connector.GroupQuery,
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

		serieSources := []string{}

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

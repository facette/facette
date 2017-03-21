package main

import (
	"context"
	"net/http"
	"strings"
	"time"

	"facette/backend"
	"facette/connector"
	"facette/plot"
	"facette/template"
	"facette/timerange"

	"github.com/facette/httputil"
)

const (
	defaultTimeRange = "-1h"
)

type plotRequest struct {
	plot.Request
	startTime time.Time
	endTime   time.Time
}

type plotQuery struct {
	query     plot.Query
	queryMap  [][2]int
	connector connector.Connector
}

func (w *httpWorker) httpHandlePlots(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	var err error

	defer r.Body.Close()

	// Check for request content type
	if ct, _ := httputil.GetContentType(r); ct != "application/json" {
		httputil.WriteJSON(rw, httpBuildMessage(ErrUnsupportedType), http.StatusUnsupportedMediaType)
		return
	}

	// Get plot request from received data
	req := &plotRequest{}
	if err := httputil.BindJSON(r, req); err != nil {
		w.log.Error("unable to unmarshal JSON data: %s", err)
		httputil.WriteJSON(rw, httpBuildMessage(ErrInvalidParameter), http.StatusBadRequest)
		return
	}

	// Request item from backend
	if req.ID != "" {
		req.Graph = &backend.Graph{}

		if err := w.service.backend.Get(req.ID, req.Graph); err == backend.ErrItemNotExist {
			httputil.WriteJSON(rw, httpBuildMessage(err), http.StatusNotFound)
			return
		} else if err != nil {
			w.log.Error("failed to fetch item: %s", err)
			httputil.WriteJSON(rw, httpBuildMessage(ErrUnhandledError), http.StatusInternalServerError)
			return
		} else if req.Graph.Template && len(req.Attributes) == 0 {
			w.log.Debug("requested item %s is a template, skipping", req.ID)
			httputil.WriteJSON(rw, httpBuildMessage(ErrInvalidParameter), http.StatusBadRequest)
			return
		}
	} else if req.Graph == nil {
		httputil.WriteJSON(rw, httpBuildMessage(ErrInvalidParameter), http.StatusBadRequest)
		return
	}

	// Expand graph template if linked
	if req.Graph.LinkID != "" || req.Graph.ID == "" && len(req.Graph.Attributes) > 0 || len(req.Attributes) > 0 {
		// Back up template attributes for later expansion
		attrs := req.Graph.Attributes
		if len(req.Attributes) > 0 {
			attrs.Merge(req.Attributes, true)
		}

		if req.Graph.LinkID != "" {
			if err := w.service.backend.Get(req.Graph.LinkID, req.Graph); err == backend.ErrItemNotExist {
				httputil.WriteJSON(rw, httpBuildMessage(err), http.StatusNotFound)
				return
			} else if err != nil {
				w.log.Error("failed to fetch linked item: %s", err)
				httputil.WriteJSON(rw, httpBuildMessage(ErrUnhandledError), http.StatusInternalServerError)
				return
			}
		}

		if err := req.Graph.Expand(attrs); err != nil {
			w.log.Warning("%s", err)
			httputil.WriteJSON(rw, httpBuildMessage(template.ErrInvalidTemplate), http.StatusBadRequest)
			return
		}
	}

	// Initialize request time boundaries and range
	if req.Time.IsZero() {
		req.Time = time.Now().UTC()
	}

	if req.Range == "" {
		if value, ok := req.Graph.Options["range"].(string); ok {
			req.Range = value
		} else {
			req.Range = defaultTimeRange
		}
	}

	if strings.HasPrefix(req.Range, "-") {
		req.endTime = req.Time
		if req.startTime, err = timerange.Apply(req.Time, req.Range); err != nil {
			w.log.Warning("unable to apply time range: %s", err)
			httputil.WriteJSON(rw, httpBuildMessage(ErrInvalidParameter), http.StatusBadRequest)
			return
		}
	} else {
		req.startTime = req.Time
		if req.endTime, err = timerange.Apply(req.Time, req.Range); err != nil {
			w.log.Warning("unable to apply time range: %s", err)
			httputil.WriteJSON(rw, httpBuildMessage(ErrInvalidParameter), http.StatusBadRequest)
			return
		}
	}

	// Set default plot sample if none provided
	if req.Sample == 0 {
		req.Sample = plot.DefaultSample
	}

	// Execute plots request
	plots := plot.Response{
		Start:   req.startTime.Format(time.RFC3339),
		End:     req.endTime.Format(time.RFC3339),
		Series:  w.executeRequest(req),
		Options: req.Graph.Options,
	}

	// Set fallback title to graph name if none provided
	if plots.Options == nil {
		plots.Options = make(map[string]interface{})
	}

	if _, ok := plots.Options["title"]; !ok {
		plots.Options["title"] = req.Graph.Name
	}

	httputil.WriteJSON(rw, plots, http.StatusOK)
}

func (w *httpWorker) executeRequest(req *plotRequest) []plot.SeriesResponse {
	// Expand groups series
	for i, group := range req.Graph.Groups {
		expandedSeries := []backend.Series{}
		for _, series := range group.Series {
			expandedSeries = append(expandedSeries, w.expandSeries(series, true)...)
		}
		req.Graph.Groups[i].Series = expandedSeries
	}

	// Dispatch plot queries among providers
	data := make([][]plot.Series, len(req.Graph.Groups))
	for i, group := range req.Graph.Groups {
		data[i] = make([]plot.Series, len(group.Series))
	}

	for _, q := range w.dispatchQueries(req) {
		series, err := q.connector.Plots(&q.query)
		if err != nil {
			w.log.Error("unable to fetch plots: %s", err)
			continue
		}

		count := len(series)
		expected := len(q.query.Series)
		if count != expected {
			w.log.Error("unable to fetch plots: expected %d series but got %d", expected, count)
			continue
		}

		// Put back series to its original indexes
		for i, s := range series {
			data[q.queryMap[i][0]][q.queryMap[i][1]] = s
		}
	}

	// Generate plots series
	result := []plot.SeriesResponse{}
	for i, group := range req.Graph.Groups {
		var (
			consolidate int
			err         error
		)

		// Skip processing if no data
		if len(data[i]) == 0 {
			goto finalize
		}

		// Apply series scale if any
		for j, series := range group.Series {
			if v, ok := series.Options["scale"].(float64); ok {
				data[i][j].Scale(plot.Value(v))
			}
		}

		// Skip operations if none requested
		if group.Operator == plot.OperatorNone {
			goto finalize
		}

		// Get group consolidation mode option
		consolidate = plot.ConsolidateAverage
		if v, ok := group.Options["consolidate"].(int); ok {
			consolidate = v
		}

		// Normalize series and apply operations
		data[i], err = plot.Normalize(data[i], req.startTime, req.endTime, req.Sample, consolidate)
		if err != nil {
			w.log.Error("failed to normalize series: %s", err)
			continue
		}

		switch group.Operator {
		case plot.OperatorAverage, plot.OperatorSum:
			var (
				series plot.Series
				err    error
			)

			if group.Operator == plot.OperatorAverage {
				series, err = plot.Average(data[i])
			} else {
				series, err = plot.Sum(data[i])
			}

			if err != nil {
				w.log.Error("failed to apply series operation: %s", err)
				continue
			}

			// Set series name to group name
			req.Graph.Groups[i].Series[0].Name = group.Name

			// Replace group series with operation result
			data[i] = []plot.Series{series}

		case plot.OperatorNormalize:
			// noop

		default:
			w.log.Warning("unknown %d operation type", group.Operator)
			continue
		}

	finalize:
		// Get group scale value
		scale, _ := group.Options["scale"].(float64)

		for j, series := range data[i] {
			// Apply group scale if any
			if scale != 0 {
				series.Scale(plot.Value(scale))
			}

			// Summarize series
			percentiles := []float64{}
			if slice, ok := req.Graph.Options["percentiles"].([]interface{}); ok {
				for _, entry := range slice {
					if val, ok := entry.(float64); ok {
						percentiles = append(percentiles, val)
					}
				}
			}

			series.Summarize(percentiles)

			result = append(result, plot.SeriesResponse{
				Series:  series,
				Name:    group.Series[j].Name,
				Options: group.Series[j].Options,
			})
		}
	}

	return result
}

func (w *httpWorker) dispatchQueries(req *plotRequest) []plotQuery {
	providers := make(map[string]*plotQuery)

	for i, group := range req.Graph.Groups {
		for j, series := range group.Series {
			if !series.IsValid() {
				w.log.Warning("invalid series metric: %s", series)
				continue
			}

			search := w.service.searcher.Metrics(series.Origin, series.Source, series.Metric, 1)
			if len(search) == 0 {
				w.log.Warning("unable to find series metric: %s", series)
				continue
			}

			// Get series connector and provider name
			c := search[0].Connector().(connector.Connector)
			provName := c.Name()

			// Initialize provider-specific plot query
			if _, ok := providers[provName]; !ok {
				providers[provName] = &plotQuery{
					query: plot.Query{
						StartTime: req.startTime,
						EndTime:   req.endTime,
						Sample:    req.Sample,
						Series:    []plot.QuerySeries{},
					},
					queryMap:  [][2]int{},
					connector: c,
				}
			}

			// Append new series to plot query and save series index
			providers[provName].query.Series = append(providers[provName].query.Series, plot.QuerySeries{
				Origin: search[0].Source().Origin().OriginalName,
				Source: search[0].Source().OriginalName,
				Metric: search[0].OriginalName,
			})

			providers[provName].queryMap = append(providers[provName].queryMap, [2]int{i, j})
		}
	}

	result := []plotQuery{}
	for _, q := range providers {
		result = append(result, *q)
	}

	return result
}

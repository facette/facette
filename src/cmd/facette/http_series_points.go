package main

import (
	"net/http"
	"strings"
	"time"

	"facette/backend"
	"facette/connector"
	"facette/series"
	"facette/template"
	"facette/timerange"

	"github.com/facette/httputil"
	"github.com/facette/sqlstorage"
	"github.com/hashicorp/go-uuid"
)

type pointQuery struct {
	query     series.Query
	queryMap  [][2]int
	connector connector.Connector
}

// api:method POST /api/v1/series/points "Retrieve series data points"
//
// This endpoint retrieves data points for all of a graph's series based on a points query specifying either one of the
// following elements:
//
//   * `id` (type _string_): ID of an existing graph
//   * `graph` (type _object_): graph object definition
//
// Optional elements:
//
//   * `time` (type _string_, default `"now"`): reference time for setting the time span (format: RFC 3339)
//   * `range` (type _string_, default `"-1h"`): time offset relative to the reference `time` option
//   * `start_time` (type _string_): absolute time start bound (format: RFC 3339)
//   * `end_time` (type _string_): absolute time end bound (format: RFC 3339)
//   * `sample` (type _integer_): data points sampling size
//   * `attributes` (type _object_): graph template attributes
//
// Note: for absolute time span selection, both `start_end` and `end_time` values must be specified.
//
// The response is an array of graph series and their data points for the requested time span.
//
// ---
// section: series
// request:
//   type: object
//   examples:
//   - format: javascript
//     headers:
//       Content-Type: application/json
//     body: |
//       {
//         "id": "c5e5faf1-dda1-50b3-abcb-4a5bdae7328e",
//         "sample": 10,
//         "range": "-60s"
//       }
// responses:
//   200:
//     type: object
//     examples:
//     - format: javascript
//       body: |
//         {
//           "start": "2017-06-07T12:28:08Z",
//           "end": "2017-06-07T12:29:08Z",
//           "series": [
//             {
//               "points": [
//                 [1496838488, 673],
//                 [1496838494, 576],
//                 [1496838500, 585.5],
//                 [1496838506, 595],
//                 [1496838512, 648],
//                 [1496838518, 678],
//                 [1496838524, 708],
//                 [1496838530, 716],
//                 [1496838536, 724],
//                 [1496838542, 733]
//               ],
//               "summary": {
//                 "avg": 662.6111111111111,
//                 "last": 733,
//                 "max": 733,
//                 "min": 576
//               },
//               "name": "lb1_example_net.current_connections",
//               "options": null
//             }
//           ],
//           "options": {
//             "title": "lb1.example.net - Current connections",
//             "type": "line",
//             "yaxis_unit": "metric"
//           }
//         }
func (w *httpWorker) httpHandleSeriesPoints(rw http.ResponseWriter, r *http.Request) {
	var err error

	defer r.Body.Close()

	// Get point request from received data
	req := &series.Request{}
	if err := httputil.BindJSON(r, req); err == httputil.ErrInvalidContentType {
		httputil.WriteJSON(rw, httpBuildMessage(err), http.StatusUnsupportedMediaType)
		return
	} else if err != nil {
		w.log.Error("unable to unmarshal JSON data: %s", err)
		httputil.WriteJSON(rw, httpBuildMessage(ErrInvalidParameter), http.StatusBadRequest)
		return
	}

	// Request item from backend
	if req.ID != "" {
		req.Graph = w.service.backend.NewGraph()

		// Check for aliased item if identifier value isn't valid
		column := "id"
		if _, err := uuid.ParseUUID(req.ID); err != nil {
			column = "alias"
		}

		if err := w.service.backend.Storage().Get(column, req.ID, req.Graph, false); err == sqlstorage.ErrItemNotFound {
			httputil.WriteJSON(rw, httpBuildMessage(err), http.StatusNotFound)
			return
		} else if err != nil {
			w.log.Error("failed to fetch item: %s", err)
			httputil.WriteJSON(rw, httpBuildMessage(ErrUnhandledError), http.StatusInternalServerError)
			return
		}
	} else if req.Graph != nil {
		// Register back-end (needed for graph expansion)
		req.Graph.Item.SetBackend(w.service.backend)
	} else {
		httputil.WriteJSON(rw, httpBuildMessage(ErrInvalidParameter), http.StatusBadRequest)
		return
	}

	// Expand graph template if linked
	if err := req.Graph.Expand(req.Attributes); req.ID == "" && err == template.ErrInvalidTemplate {
		httputil.WriteJSON(rw, httpBuildMessage(err), http.StatusBadRequest)
		return
	} else if err != nil {
		w.log.Error("%s", err)
		httputil.WriteJSON(rw, httpBuildMessage(ErrUnhandledError), http.StatusInternalServerError)
		return
	}

	// Set request time boundaries and range
	// * both start and end time must be provided, or none
	// * range can't be specified if start and end are
	if req.StartTime.IsZero() && req.EndTime.IsZero() {
		if req.Time.IsZero() {
			req.Time = time.Now().UTC()
		}

		if req.Range == "" {
			if value, ok := req.Graph.Options["range"].(string); ok {
				req.Range = value
			} else {
				req.Range = w.service.config.DefaultTimeRange
			}
		}

		if strings.HasPrefix(req.Range, "-") {
			req.EndTime = req.Time
			if req.StartTime, err = timerange.Apply(req.Time, req.Range); err != nil {
				w.log.Warning("unable to apply time range: %s", err)
				httputil.WriteJSON(rw, httpBuildMessage(ErrInvalidParameter), http.StatusBadRequest)
				return
			}
		} else {
			req.StartTime = req.Time
			if req.EndTime, err = timerange.Apply(req.Time, req.Range); err != nil {
				w.log.Warning("unable to apply time range: %s", err)
				httputil.WriteJSON(rw, httpBuildMessage(ErrInvalidParameter), http.StatusBadRequest)
				return
			}
		}
	} else if (req.StartTime.IsZero() || req.EndTime.IsZero()) || req.Range != "" {
		httputil.WriteJSON(rw, httpBuildMessage(ErrInvalidTimerange), http.StatusBadRequest)
		return
	}

	// Set default point sample if none provided
	if req.Sample == 0 {
		req.Sample = series.DefaultSample
	}

	// Execute points request
	points := series.Response{
		Start:   req.StartTime.Format(time.RFC3339),
		End:     req.EndTime.Format(time.RFC3339),
		Series:  w.executeRequest(req, httpGetBoolParam(r, "normalize")),
		Options: req.Graph.Options,
	}

	// Set fallback title to graph name if none provided
	if points.Options == nil {
		points.Options = make(map[string]interface{})
	}

	if _, ok := points.Options["title"]; !ok {
		points.Options["title"] = req.Graph.Name
	}

	httputil.WriteJSON(rw, points, http.StatusOK)
}

func (w *httpWorker) executeRequest(req *series.Request, forceNormalize bool) []series.SeriesResponse {
	// Expand groups series
	for _, group := range req.Graph.Groups {
		expandedSeries := []*backend.Series{}
		for _, s := range group.Series {
			expandedSeries = append(expandedSeries, w.expandSeries(s, true)...)
		}
		group.Series = expandedSeries
	}

	// Dispatch point queries among providers
	data := make([][]series.Series, len(req.Graph.Groups))
	for i, group := range req.Graph.Groups {
		data[i] = make([]series.Series, len(group.Series))
	}

	for _, q := range w.dispatchQueries(req) {
		points, err := q.connector.Points(&q.query)
		if err != nil {
			w.log.Error("unable to fetch points: %s", err)
			continue
		}

		count := len(points)
		expected := len(q.query.Series)
		if count != expected {
			w.log.Error("unable to fetch points: expected %d series but got %d", expected, count)
			continue
		}

		// Put back series to its original indexes
		for i, p := range points {
			data[q.queryMap[i][0]][q.queryMap[i][1]] = p
		}
	}

	// Lower sample size if too few points available
	maxPoints := 0
	for i, group := range req.Graph.Groups {
		for j := range group.Series {
			if n := len(data[i][j].Points); n > maxPoints {
				maxPoints = n
			}
		}
	}

	if req.Sample > maxPoints && maxPoints > 0 {
		req.Sample = maxPoints
	}

	// Generate points series
	result := []series.SeriesResponse{}
	for i, group := range req.Graph.Groups {
		var (
			consolidate int
			interpolate bool
			err         error
		)

		// Skip processing if no data
		if len(data[i]) == 0 {
			goto finalize
		}

		// Apply series scale if any
		for j, s := range group.Series {
			if v, ok := s.Options["scale"].(float64); ok {
				data[i][j].Scale(series.Value(v))
			}
		}

		// Skip normalization if operator is not set and not forced
		if group.Operator == series.OperatorNone && !forceNormalize {
			goto finalize
		}

		// Get group consolidation mode and interpolation options
		consolidate = series.ConsolidateAverage
		if v, ok := group.Options["consolidate"].(int); ok {
			consolidate = v
		}

		interpolate = true
		if v, ok := group.Options["interpolate"].(bool); ok {
			interpolate = v
		}

		// Normalize series and apply operations
		data[i], err = series.Normalize(data[i], req.StartTime, req.EndTime, req.Sample, consolidate, interpolate)
		if err != nil {
			w.log.Error("failed to normalize series: %s", err)
			continue
		}

		switch group.Operator {
		case series.OperatorAverage, series.OperatorSum:
			var (
				s   series.Series
				err error
			)

			if group.Operator == series.OperatorAverage {
				s, err = series.Average(data[i])
			} else {
				s, err = series.Sum(data[i])
			}

			if err != nil {
				w.log.Error("failed to apply series operation: %s", err)
				continue
			}

			// Set series name to group name
			group.Series[0].Name = group.Name

			// Replace group series with operation result
			data[i] = []series.Series{s}

		case series.OperatorNone:
			// noop

		default:
			w.log.Warning("unknown %d operation type", group.Operator)
			continue
		}

	finalize:
		// Get group scale value
		scale, _ := group.Options["scale"].(float64)

		for j, s := range data[i] {
			// Apply group scale if any
			if scale != 0 {
				s.Scale(series.Value(scale))
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

			s.Summarize(percentiles)

			result = append(result, series.SeriesResponse{
				Series:  s,
				Name:    group.Series[j].Name,
				Options: group.Series[j].Options,
			})
		}
	}

	return result
}

func (w *httpWorker) dispatchQueries(req *series.Request) []pointQuery {
	providers := make(map[string]*pointQuery)

	for i, group := range req.Graph.Groups {
		for j, s := range group.Series {
			if !s.IsValid() {
				w.log.Warning("invalid series metric: %s", s)
				continue
			}

			search := w.service.searcher.Metrics(s.Origin, s.Source, s.Metric, 1)
			if len(search) == 0 {
				w.log.Warning("unable to find series metric: %s", s)
				continue
			}

			// Get series connector and provider name
			c := search[0].Connector().(connector.Connector)
			provName := c.Name()

			// Initialize provider-specific point query
			if _, ok := providers[provName]; !ok {
				providers[provName] = &pointQuery{
					query: series.Query{
						StartTime: req.StartTime,
						EndTime:   req.EndTime,
						Sample:    req.Sample,
						Series:    []series.QuerySeries{},
					},
					queryMap:  [][2]int{},
					connector: c,
				}
			}

			// Append new series to point query and save series index
			providers[provName].query.Series = append(providers[provName].query.Series, series.QuerySeries{
				Origin: search[0].Source().Origin().OriginalName,
				Source: search[0].Source().OriginalName,
				Metric: search[0].OriginalName,
			})

			providers[provName].queryMap = append(providers[provName].queryMap, [2]int{i, j})
		}
	}

	result := []pointQuery{}
	for _, q := range providers {
		result = append(result, *q)
	}

	return result
}

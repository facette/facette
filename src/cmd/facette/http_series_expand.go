package main

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"facette/backend"

	"github.com/facette/httputil"
	"github.com/fatih/set"
)

// api:section series "Series"

// api:method POST /api/v1/series/expand "Expand groups in series"
//
// This endpoint performs source/metric group expansion for a specific origin. The input format is a list of series
// element (`origin`/`source`/`metric`), where the both of the `source` and `metric` field value can be a reference to
// an existing source/metric group ID.
//
// The format for describing a series in a expansion list is:
//
// ```javascript
// {
//   "origin": "<origin name>",
//   "source": "<source name or source group identifier (format: `group:ID`)>",
//   "metric": "<metric name or metric group identifier (format: `group:ID`)>"
// }
// ```
//
// The response is a list of series (origin/source/metric, and a pre-formatted `name` field for display purposes).
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
//       [
//         {
//           "origin": "kairosdb",
//           "source": "host1.example.net",
//           "metric": "group:118e864e-d880-5499-864b-06dedfd9f9ef"
//         }
//       ]
// responses:
//   200:
//     type: array
//     examples:
//     - format: javascript
//       body: |
//         [
//             {
//               "name": "host1.example.net (load.shortterm)",
//               "origin": "kairosdb",
//               "source": "host1.example.net",
//               "metric": "load.shortterm"
//             },
//             {
//               "name": "host1.example.net (load.midterm)",
//               "origin": "kairosdb",
//               "source": "host1.example.net",
//               "metric": "load.midterm"
//             },
//             {
//               "name": "host1.example.net (load.longterm)",
//               "origin": "kairosdb",
//               "source": "host1.example.net",
//               "metric": "load.longterm"
//             }
//           ]
//         ]
func (w *httpWorker) httpHandleSeriesExpand(rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Get expand request from received data
	series := []*backend.Series{}
	if err := httputil.BindJSON(r, &series); err == httputil.ErrInvalidContentType {
		httputil.WriteJSON(rw, httpBuildMessage(err), http.StatusUnsupportedMediaType)
		return
	} else if err != nil {
		w.log.Error("unable to unmarshal JSON data: %s", err)
		httputil.WriteJSON(rw, httpBuildMessage(ErrInvalidJSON), http.StatusBadRequest)
		return
	}

	result := make([][]*backend.Series, len(series))

	for i, s := range series {
		result[i] = w.expandSeries(s, false)
	}

	httputil.WriteJSON(rw, result, http.StatusOK)
}

func (w *httpWorker) expandSeries(series *backend.Series, existOnly bool) []*backend.Series {
	var hasGroup bool

	out := []*backend.Series{}

	sourcesSet := set.New()
	if strings.HasPrefix(series.Source, backend.GroupPrefix) {
		id := strings.TrimPrefix(series.Source, backend.GroupPrefix)

		// Request source group from back-end
		group := backend.SourceGroup{}
		if err := w.service.backend.Storage().Get("id", id, &group, false); err != nil {
			w.log.Warning("unable to expand %s source group: %s", id, err)
			return nil
		}

		// Loop through sources checking for patterns matching
		for _, s := range w.service.searcher.Sources(series.Origin, "", -1) {
			for _, p := range group.Patterns {
				if match, err := filterMatch(p, s.Name); err != nil {
					w.log.Error("failed to match filter: %s", err)
					return nil
				} else if match {
					sourcesSet.Add(s.Name)
				}
			}
		}

		hasGroup = true
	} else {
		sourcesSet.Add(series.Source)
	}

	metricsSet := set.New()
	if strings.HasPrefix(series.Metric, backend.GroupPrefix) {
		id := strings.TrimPrefix(series.Metric, backend.GroupPrefix)

		// Request metric group from back-end
		group := backend.MetricGroup{}
		if err := w.service.backend.Storage().Get("id", id, &group, false); err != nil {
			w.log.Warning("unable to expand %s metric group: %s", id, err)
			return nil
		}

		// Loop through metrics checking for patterns matching
		for _, m := range w.service.searcher.Metrics(series.Origin, "", "", -1) {
			// Skip if metric source does not match an existing metric
			if existOnly && !sourcesSet.Has(m.Source().Name) {
				continue
			}

			for _, p := range group.Patterns {
				if match, err := filterMatch(p, m.Name); err != nil {
					w.log.Error("failed to match filter: %s", err)
					return nil
				} else if match {
					metricsSet.Add(m.Name)
				}
			}
		}

		hasGroup = true
	} else {
		metricsSet.Add(series.Metric)
	}

	count := 0

	sources := set.StringSlice(sourcesSet)
	metrics := set.StringSlice(metricsSet)
	sort.Strings(sources)
	sort.Strings(metrics)

	for _, source := range sources {
		for _, metric := range metrics {
			var name string

			// Override name if source/series has been expanded
			if hasGroup {
				name = fmt.Sprintf("%s (%s)", source, metric)
				count++
			} else {
				name = series.Name
			}

			out = append(out, &backend.Series{
				Name:    name,
				Origin:  series.Origin,
				Source:  source,
				Metric:  metric,
				Options: series.Options,
			})
		}
	}

	return out
}

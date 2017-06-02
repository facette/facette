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

// api:section expand "Expand"

// api:method POST /expand/ "Expand source/metric group in graph series"
//
// This endpoint performs source/metric group expansion for a specific origin. The input format is a list of series
// element (`origin`/`source`/`metric`), where the both of the `source` and `metric` field value can be a reference to
// an existing source/metric group ID.
//
// The format for describing a series in a expansion list is:
//
// ```
// {
//   "origin": "< Origin name >",
//   "source": "< Source name or source group ID (format: `group:ID`) >",
//   "metric": "< Metric name or metric group ID (format: `group:ID`) >"
// }
// ```
//
// Here is an example of an expansion request body:
//
// ```
// [
//   {
//     "origin": "kairosdb",
//     "source": "host1.example.net",
//     "metric": "group:118e864e-d880-5499-864b-06dedfd9f9ef"
//   }
// ]
// ```
//
// The response is a list of series (origin/source/metric, and a pre-formatted `name` field for display purposes).
//
// ---
// section: expand
// content_types:
// - application/json
// responses:
//   200:
//     type: array
//     example:
//       format: json
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
func (w *httpWorker) httpHandleExpand(rw http.ResponseWriter, r *http.Request) {
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
	} else {
		metricsSet.Add(series.Metric)
	}

	multiple := sourcesSet.Size() > 1 || metricsSet.Size() > 1
	count := 0

	sources := set.StringSlice(sourcesSet)
	metrics := set.StringSlice(metricsSet)
	sort.Strings(sources)
	sort.Strings(metrics)

	for _, source := range sources {
		for _, metric := range metrics {
			var name string

			// Override name if source/series has been expanded
			if multiple {
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

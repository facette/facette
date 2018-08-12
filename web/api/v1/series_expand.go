package v1

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"facette.io/facette/pattern"
	"facette.io/facette/set"
	"facette.io/facette/storage"
	"facette.io/httputil"
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
func (a *API) seriesExpand(rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Get expand request from received data
	series := []*storage.Series{}
	if err := httputil.BindJSON(r, &series); err == httputil.ErrInvalidContentType {
		httputil.WriteJSON(rw, newMessage(err), http.StatusUnsupportedMediaType)
		return
	} else if err != nil {
		a.logger.Error("unable to unmarshal JSON data: %s", err)
		httputil.WriteJSON(rw, newMessage(errInvalidJSON), http.StatusBadRequest)
		return
	}

	result := make([][]*storage.Series, len(series))

	for i, s := range series {
		result[i] = a.expandSeries(s, false)
	}

	httputil.WriteJSON(rw, result, http.StatusOK)
}

func (a *API) expandSeries(series *storage.Series, existOnly bool) []*storage.Series {
	var hasGroup bool

	out := []*storage.Series{}

	sourcesSet := set.New()
	if strings.HasPrefix(series.Source, storage.GroupPrefix) {
		id := strings.TrimPrefix(series.Source, storage.GroupPrefix)

		// Request source group from storage
		group := storage.SourceGroup{}
		if err := a.storage.SQL().Get("id", id, &group, false); err != nil {
			a.logger.Warning("unable to expand %s source group: %s", id, err)
			return nil
		}

		// Loop through sources checking for patterns matching
		for _, s := range a.searcher.Sources(series.Origin, "", -1) {
			for _, p := range group.Patterns {
				if match, err := pattern.Match(p, s.Name); err != nil {
					a.logger.Error("failed to match filter: %s", err)
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
	if strings.HasPrefix(series.Metric, storage.GroupPrefix) {
		id := strings.TrimPrefix(series.Metric, storage.GroupPrefix)

		// Request metric group from storage
		group := storage.MetricGroup{}
		if err := a.storage.SQL().Get("id", id, &group, false); err != nil {
			a.logger.Warning("unable to expand %s metric group: %s", id, err)
			return nil
		}

		// Loop through metrics checking for patterns matching
		for _, m := range a.searcher.Metrics(series.Origin, "", "", -1) {
			// Skip if metric source does not match an existing metric
			if existOnly && !sourcesSet.Has(m.Source().Name) {
				continue
			}

			for _, p := range group.Patterns {
				if match, err := pattern.Match(p, m.Name); err != nil {
					a.logger.Error("failed to match filter: %s", err)
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

			out = append(out, &storage.Series{
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

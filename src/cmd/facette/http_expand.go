package main

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"facette/backend"

	"github.com/facette/httputil"
	"github.com/fatih/set"
)

type expandListEntry [3]string

func (w *httpWorker) httpHandleExpand(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Get expand request from received data
	list := []expandListEntry{}
	if err := httputil.BindJSON(r, &list); err == httputil.ErrInvalidContentType {
		httputil.WriteJSON(rw, httpBuildMessage(err), http.StatusUnsupportedMediaType)
		return
	} else if err != nil {
		w.log.Error("unable to unmarshal JSON data: %s", err)
		httputil.WriteJSON(rw, httpBuildMessage(ErrInvalidJSON), http.StatusBadRequest)
		return
	}

	result := make([][]expandListEntry, len(list))

	for i, entry := range list {
		s := &backend.Series{
			Origin: entry[0],
			Source: entry[1],
			Metric: entry[2],
		}

		result[i] = []expandListEntry{}
		for _, s := range w.expandSeries(s, false) {
			result[i] = append(result[i], expandListEntry{s.Origin, s.Source, s.Metric})
		}
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
		if err := w.service.backend.Storage().Get("id", id, &group); err != nil {
			w.log.Warning("unable to expand %s source group: %s", id, err)
			return nil
		}

		// Loop through sources checking for patterns matching
		for _, s := range w.service.searcher.Sources(series.Origin, "", -1) {
			if group.Patterns == nil {
				continue
			}

			for _, p := range *group.Patterns {
				if filterMatch(p, s.Name) {
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
		if err := w.service.backend.Storage().Get("id", id, &group); err != nil {
			w.log.Warning("unable to expand %s metric group: %s", id, err)
			return nil
		}

		// Loop through metrics checking for patterns matching
		for _, m := range w.service.searcher.Metrics(series.Origin, "", "", -1) {
			// Skip if metric source does not match an existing metric or if no pattern
			if existOnly && !sourcesSet.Has(m.Source().Name) || group.Patterns == nil {
				continue
			}

			for _, p := range *group.Patterns {
				if filterMatch(p, m.Name) {
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

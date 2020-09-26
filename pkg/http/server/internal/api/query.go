// Copyright (c) 2020, The Facette Authors
//
// Licensed under the terms of the BSD 3-Clause License; a copy of the license
// is available at: https://opensource.org/licenses/BSD-3-Clause

package api

import (
	"context"
	"net/http"

	"go.uber.org/zap"

	"facette.io/facette/pkg/api"
	"facette.io/facette/pkg/catalog"
	"facette.io/facette/pkg/connector"
	"facette.io/facette/pkg/errors"
	httpjson "facette.io/facette/pkg/http/json"
	"facette.io/facette/pkg/labels"
	"facette.io/facette/pkg/series"
)

func (h handler) ExecQuery(rw http.ResponseWriter, r *http.Request) {
	q := &series.Query{}

	err := httpjson.Unmarshal(r, q)
	if err != nil {
		h.WriteError(rw, err)
		return
	}

	metrics := map[*catalog.Metric][]series.Point{}

	for _, cq := range dispatchQuery(q, h.catalog) {
		if cq.Connector == nil {
			h.log.Error("invalid connector association", zap.Any("metrics", cq.Query.Metrics))
			continue
		}

		result, err := cq.Connector.Query(r.Context(), cq.Query)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				h.log.Error("cannot fetch points", zap.Error(err))
			}

			continue
		}

		for idx := range result {
			metrics[&result[idx].Metric] = result[idx].Points
		}
	}

	httpjson.Write(rw, api.Response{Data: series.Result{
		From:   q.From,
		To:     q.To,
		Step:   q.Step,
		Series: series.Render(q, metrics),
	}}, http.StatusOK)
}

func dispatchQuery(q *series.Query, cat *catalog.Catalog) []connectorQuery {
	dispatch := map[string]connectorQuery{}

	for _, matcher := range series.MatchersFromExprs(q.Exprs...) {
		for _, metric := range cat.Metrics(&catalog.ListOptions{Matcher: matcher}) {
			provider := metric.Labels.Get(labels.Provider)

			_, ok := dispatch[provider]
			if !ok {
				dispatch[provider] = connectorQuery{
					Connector: metric.Connector().(connector.Connector),
					Query: &connector.Query{
						From: q.From.Time,
						To:   q.To.Time,
						Step: q.Step.Duration,
					},
				}
			}

			dispatch[provider].Query.Metrics = append(dispatch[provider].Query.Metrics, metric)
		}
	}

	queries := []connectorQuery{}
	for _, q := range dispatch {
		queries = append(queries, q)
	}

	return queries
}

type connectorQuery struct {
	Connector connector.Connector
	Query     *connector.Query
}

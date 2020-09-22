// Copyright (c) 2020, The Facette Authors
//
// Licensed under the terms of the BSD 3-Clause License; a copy of the license
// is available at: https://opensource.org/licenses/BSD-3-Clause

package api

import (
	"net/http"

	"batou.dev/httprouter"
	"facette.io/facette/pkg/api"
	"facette.io/facette/pkg/errors"
	httpjson "facette.io/facette/pkg/http/json"
)

func (h handler) ListLabels(rw http.ResponseWriter, r *http.Request) {
	opts, err := listOptionsFromRequest(r)
	if err != nil {
		h.WriteError(rw, err)
		return
	}

	filter := httprouter.QueryParam(r, "filter")

	matcher, err := matcherFromRequest(r)
	if err != nil {
		h.WriteError(rw, errors.Wrap(api.ErrInvalid, err.Error()))
		return
	}

	labels := h.catalog.Labels(matcher, filter)
	total := int64(len(labels))

	if opts.Limit > 0 {
		applyCatalogPagination(&labels, total, opts.Offset, opts.Limit)
	}

	httpjson.Write(rw, api.Response{Data: labels, Total: total}, http.StatusOK)
}

func (h handler) ListMetrics(rw http.ResponseWriter, r *http.Request) {
	opts, err := listOptionsFromRequest(r)
	if err != nil {
		h.WriteError(rw, err)
		return
	}

	matcher, err := matcherFromRequest(r)
	if err != nil {
		h.WriteError(rw, errors.Wrap(api.ErrInvalid, err.Error()))
		return
	}

	metrics := []string{}
	for _, metric := range h.catalog.Metrics(matcher) {
		metrics = append(metrics, metric.String())
	}

	total := int64(len(metrics))

	if opts.Limit > 0 {
		applyCatalogPagination(&metrics, total, opts.Offset, opts.Limit)
	}

	httpjson.Write(rw, api.Response{Data: metrics, Total: total}, http.StatusOK)
}

func (h handler) ListValues(rw http.ResponseWriter, r *http.Request) {
	opts, err := listOptionsFromRequest(r)
	if err != nil {
		h.WriteError(rw, err)
		return
	}

	filter := httprouter.QueryParam(r, "filter")

	matcher, err := matcherFromRequest(r)
	if err != nil {
		h.WriteError(rw, errors.Wrap(api.ErrInvalid, err.Error()))
		return
	}

	var names []string

	v := httprouter.QueryParam(r, "name")
	if v != "" {
		names = append(names, v)
	} else {
		names = h.catalog.Labels(matcher, "")
	}

	values := map[string]api.LabelValues{}

	for _, name := range names {
		subValues := h.catalog.Values(name, matcher, filter)

		total := int64(len(subValues))
		if total == 0 {
			continue
		}

		if opts.Limit > 0 {
			applyCatalogPagination(&subValues, total, opts.Offset, opts.Limit)
		}

		values[name] = api.LabelValues{
			Values: subValues,
			Total:  total,
		}
	}

	httpjson.Write(rw, api.Response{Data: values}, http.StatusOK)
}

func applyCatalogPagination(entries *[]string, total, offset, limit int64) {
	if offset < total {
		end := offset + limit
		if limit > 0 && total > end {
			*entries = (*entries)[offset:end]
		} else if offset > 0 {
			*entries = (*entries)[offset:]
		}
	} else {
		*entries = []string{}
	}
}

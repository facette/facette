package main

import (
	"context"
	"fmt"
	"net/http"
	"reflect"

	"facette/backend"

	"github.com/facette/httputil"
)

type searchRequest struct {
	Types []string               `json:"types"`
	Terms map[string]interface{} `json:"terms"`
}

type searchRecord struct {
	Type  string       `json:"type"`
	Value backend.Item `json:"value"`
}

func (w *httpWorker) httpHandleLibrarySearch(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Check for request content type
	if ct, _ := httputil.GetContentType(r); ct != "application/json" {
		httputil.WriteJSON(rw, httpBuildMessage(ErrUnsupportedType), http.StatusUnsupportedMediaType)
		return
	}

	// Get search request from received data
	req := searchRequest{}
	if err := httputil.BindJSON(r, &req); err != nil {
		w.log.Error("unable to unmarshal JSON data: %s", err)
		httputil.WriteJSON(rw, httpBuildMessage(ErrInvalidParameter), http.StatusBadRequest)
		return
	}

	if req.Terms == nil {
		httputil.WriteJSON(rw, httpBuildMessage(ErrInvalidParameter), http.StatusBadRequest)
		return
	}

	// Get requested types for request or fallback to 'all'
	if len(req.Types) == 0 {
		for typ := range backendTypes {
			req.Types = append(req.Types, typ)
		}
	}

	types := []interface{}{}
	for _, typ := range req.Types {
		if rt, ok := backendTypes[typ]; ok {
			rv := reflect.New(rt)
			types = append(types, reflect.Indirect(rv).Interface())
		}
	}

	offset, err := httpGetIntParam(r, "offset")
	if err != nil || offset < 0 {
		httputil.WriteJSON(rw, httpBuildMessage(ErrInvalidParameter), http.StatusBadRequest)
		return
	}

	limit, err := httpGetIntParam(r, "limit")
	if err != nil || limit < 0 {
		httputil.WriteJSON(rw, httpBuildMessage(ErrInvalidParameter), http.StatusBadRequest)
		return
	}

	sort := httpGetListParam(r, "sort", []string{"name"})

	// Execute search request
	items, count, err := w.service.backend.Search(types, req.Terms, sort, offset, limit)
	if err != nil {
		w.log.Error("failed to perform search: %s", err)
		httputil.WriteJSON(rw, httpBuildMessage(ErrUnhandledError), http.StatusInternalServerError)
		return
	}

	result := []searchRecord{}
	for _, entry := range items {
		result = append(result, searchRecord{
			Type:  entry.Type,
			Value: entry.Item,
		})
	}

	rw.Header().Set("X-Total-Records", fmt.Sprintf("%d", count))
	httputil.WriteJSON(rw, result, http.StatusOK)
}

package main

import (
	"fmt"
	"net/http"

	"facette/backend"

	"github.com/facette/httputil"
)

type searchRequest struct {
	Types []string               `json:"types"`
	Terms map[string]interface{} `json:"terms"`
}

func (w *httpWorker) httpHandleLibrarySearch(rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Get search request from received data
	req := searchRequest{}
	if err := httputil.BindJSON(r, &req); err == httputil.ErrInvalidContentType {
		httputil.WriteJSON(rw, httpBuildMessage(err), http.StatusUnsupportedMediaType)
		return
	} else if err != nil {
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
		for _, typ := range backendTypes {
			req.Types = append(req.Types, typ)
		}
	}

	types := []interface{}{}
	for _, typ := range req.Types {
		if item, ok := w.httpBackendNewItem(typ); ok {
			types = append(types, item)
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

	// Apply back-end storage modifiers
	for k, v := range req.Terms {
		if s, ok := v.(string); ok {
			req.Terms[k] = filterApplyModifier(s)
		}
	}

	// Execute search request
	result := []*backend.Item{}

	count, err := w.service.backend.Storage().Search(types, &result, req.Terms, sort, offset, limit)
	if err != nil {
		w.log.Error("failed to perform search: %s", err)
		httputil.WriteJSON(rw, httpBuildMessage(ErrUnhandledError), http.StatusInternalServerError)
		return
	}

	rw.Header().Set("X-Total-Records", fmt.Sprintf("%d", count))
	httputil.WriteJSON(rw, result, http.StatusOK)
}

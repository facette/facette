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

// api:method POST /api/v1/library/search "Search library for items"
//
// This endpoint searches library for items matching a set of types and terms.
//
// This endpoint supports pagination through the `offset` and `limit` query parameters and sorting using `sort` query
// parameter (separated by commas; prefix field name with "-" to reverse sort order).
//
// ---
// section: library
// parameters:
// - name: sort
//   type: string
//   description: fields to sort results on
//   in: query
// - name: offset
//   type: integer
//   description: offset to return items from
//   in: query
// - name: limit
//   type: integer
//   description: number of items to return
//   in: query
// request:
//   type: object
//   examples:
//   - format: javascript
//     headers:
//       Content-Type: application/json
//     body: |
//       {
//         "types": ["collections", "graphs"],
//         "terms": {
//           "name": "glob:*test*",
//           "template": false
//         }
//       }
// responses:
//   200:
//     type: array
//     headers:
//       X-Total-Records: total number of library items found
//     examples:
//     - headers:
//         X-Total-Records: 2
//       format: javascript
//       body: |
//         [
//           {
//             "type": "collections",
//             "id": "0f660bc7-c8d7-4beb-497e-f1fdbf14092a",
//             "name": "collection1",
//             "description": null,
//             "created": "2017-05-27T11:36:00Z",
//             "modified": "2017-06-12T06:18:48Z"
//           },
//           {
//             "type": "graphs",
//             "id": "b3233810-ceb2-5e7a-17df-336b2710eef2",
//             "name": "graph3",
//             "description": null,
//             "created": "2017-05-27T11:35:43Z",
//             "modified": "2017-06-12T06:18:48Z"
//           }
//         ]
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

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/facette/httputil"
)

// api:section bulk "Bulk"

type bulkRequest []bulkRequestEntry

type bulkRequestEntry struct {
	Endpoint string                 `json:"endpoint"`
	Method   string                 `json:"method"`
	Params   map[string]interface{} `json:"params"`
	Data     json.RawMessage        `json:"data"`
}

type bulkResponse []bulkResponseEntry

type bulkResponseEntry struct {
	Status int         `json:"status"`
	Data   interface{} `json:"data"`
}

func init() {
}

// api:method POST /api/v1/bulk/ "Bulk API requests execution"
//
// This endpoint expects a request providing as body a list of API requests to execute in bulk, and returns a list of
// API responses corresponding to the requests. The format for describing an API request in a bulk list is:
//
// ```
// {
//   "endpoint": "< API endpoint relative to prefix /api/v1/ >",
//   "method": "< HTTP Mmethod >",
//   "params": {
//     < Query string parameters >
// }
// ```
//
// Here is an example of a bulk request body listing several requests to be performed in bulk:
//
// ```
// [
//   {
//     "endpoint": "library/graphs/9084083e-312f-55cf-9bd6-57406cfad22a",
//     "method": "GET",
//     "params": {
//       "fields": "id,name"
//     }
//   },
//   {
//     "endpoint": "library/graphs/65f812e1-9856-5a2c-8f1a-8e349f8945f0",
//     "method": "GET",
//     "params": {
//       "fields": "id,name"
//     }
//   },
//   {
//     "endpoint": "library/graphs/36bdae08-8d4e-51cb-87d1-f016bed65864",
//     "method": "GET",
//     "params": {
//       "fields": "id,name"
//     }
//   }
// ]
// ```
//
// ---
// section: bulk
// content_types:
// - application/json
// responses:
//   200:
//     type: array
//     example:
//       format: json
//       body: |
//         [
//           {
//             "status": 200,
//             "data": {
//               "id": "9084083e-312f-55cf-9bd6-57406cfad22a",
//               "name": "www_facette_io.request.latency"
//             }
//           },
//           {
//             "status": 200,
//             "data": {
//               "id": "65f812e1-9856-5a2c-8f1a-8e349f8945f0",
//               "name": "docs_facette_io.request.latency"
//             }
//           },
//           {
//             "status": 200,
//             "data": {
//               "id": "36bdae08-8d4e-51cb-87d1-f016bed65864",
//               "name": "blog_facette_io.request.latency"
//             }
//           }
//         ]
func (w *httpWorker) httpHandleBulk(rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Get search request from received data
	req := bulkRequest{}
	if err := httputil.BindJSON(r, &req); err == httputil.ErrInvalidContentType {
		httputil.WriteJSON(rw, httpBuildMessage(err), http.StatusUnsupportedMediaType)
		return
	} else if err != nil {
		w.log.Error("unable to unmarshal JSON data: %s", err)
		httputil.WriteJSON(rw, httpBuildMessage(ErrInvalidParameter), http.StatusBadRequest)
		return
	}

	result := make(bulkResponse, len(req))
	for idx, entry := range req {
		// Prepare sub-request
		rec := httptest.NewRecorder()

		r, err := http.NewRequest(entry.Method, w.prefix+"/"+strings.TrimLeft(entry.Endpoint, "/"),
			bytes.NewReader(entry.Data))
		if err != nil {
			w.log.Error("unable to generate bulk sub-request: %s", err)
			result[idx].Status = http.StatusInternalServerError
			continue
		}

		switch entry.Method {
		case "PATCH", "POST", "PUT":
			r.Header.Set("Content-Type", "application/json")
		}

		// Generate query string form parameters
		q := r.URL.Query()
		for key, value := range entry.Params {
			q.Set(key, fmt.Sprintf("%v", value))
		}
		r.URL.RawQuery = q.Encode()

		// Set remote address to internal (displayed in debugging logs)
		r.RemoteAddr = "<internal>"

		w.router.ServeHTTP(rec, r)

		// Generate response entry
		result[idx] = bulkResponseEntry{
			Status: rec.Code,
		}

		json.Unmarshal(rec.Body.Bytes(), &result[idx].Data)
	}

	httputil.WriteJSON(rw, result, http.StatusOK)
}

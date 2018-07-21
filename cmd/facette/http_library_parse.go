package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"facette.io/facette/backend"
	"facette.io/facette/template"
	"facette.io/httputil"
)

type parseRequest struct {
	ID   string          `json:"id"`
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// api:method POST /api/v1/library/parse "Retrieve template keys"
//
// This endpoint parses requested library item or received data and returns the template keys.
//
// | Name | Type | Description |
// | --- | --- | --- |
// | `id` | string | identifier of the item |
// | `type`| string | type of the item |
// | `data` | string | arbitrary data to parse |
//
// Note: you should either specify `id` and `type` or `data` but not both.
//
// ---
// section: library
// request:
//   type: object
//   examples:
//   - format: javascript
//     headers:
//       Content-Type: application/json
//     body: |
//       {
//         "id": "368b62f2-873d-580c-ba24-440325af0582",
//         "type": "collections"
//       }
//   - format: javascript
//     headers:
//       Content-Type: application/json
//     body: |
//       {
//         "data": "{\"description\":\"A test string with {{ .key1 }}.\"}"
//       }
// responses:
//   200:
//     type: array
//     examples:
//     - format: javascript
//       body: |
//         [
//           "key1",
//           "key2"
//         ]
func (w *httpWorker) httpHandleLibraryParse(rw http.ResponseWriter, r *http.Request) {
	var data string

	defer r.Body.Close()

	// Get parse request from received data
	req := &parseRequest{}
	if err := httputil.BindJSON(r, req); err == httputil.ErrInvalidContentType {
		httputil.WriteJSON(rw, httpBuildMessage(err), http.StatusUnsupportedMediaType)
		return
	} else if err != nil {
		w.log.Error("unable to unmarshal JSON data: %s", err)
		httputil.WriteJSON(rw, httpBuildMessage(ErrInvalidParameter), http.StatusBadRequest)
		return
	}

	if req.ID != "" && req.Type != "" && len(req.Data) == 0 {
		// Check if requested type is valid
		if req.Type != "collections" && req.Type != "graphs" {
			httputil.WriteJSON(rw, httpBuildMessage(ErrInvalidParameter), http.StatusBadRequest)
			return
		}

		// Make internal request to retrieve item
		paths := []string{w.prefix + "/library/" + req.Type + "/" + req.ID}

		if req.Type == "collections" {
			collection := backend.Collection{}
			if err := w.service.backend.Storage().Get("id", req.ID, &collection, true); err == nil {
				for _, entry := range collection.Entries {
					paths = append(paths, w.prefix+"/library/graphs/"+entry.GraphID)
				}
			}
		}

		for _, path := range paths {
			rec := httptest.NewRecorder()

			r, err := http.NewRequest("GET", path, bytes.NewReader(nil))
			if err != nil {
				w.log.Error("unable to generate parse sub-request: %s", err)
				httputil.WriteJSON(rw, httpBuildMessage(ErrUnhandledError), http.StatusInternalServerError)
				return
			}

			// Set remote address to internal (displayed in debugging logs)
			r.RemoteAddr = "<internal>"

			w.router.ServeHTTP(rec, r)

			data += rec.Body.String()
		}
	} else if len(req.Data) > 0 {
		data = string(req.Data)
	} else {
		httputil.WriteJSON(rw, httpBuildMessage(ErrInvalidParameter), http.StatusBadRequest)
		return
	}

	result, err := template.Parse(data)
	if err != nil {
		w.log.Error("failed to parse template data: %s", err)
		httputil.WriteJSON(rw, httpBuildMessage(template.ErrInvalidTemplate), http.StatusBadRequest)
		return
	}

	httputil.WriteJSON(rw, result, http.StatusOK)
}

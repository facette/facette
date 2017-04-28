package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"facette/backend"
	"facette/template"

	"github.com/facette/httputil"
)

type parseRequest struct {
	ID   string          `json:"id"`
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func (w *httpWorker) httpHandleLibraryParse(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	var data string

	defer r.Body.Close()

	// Check for request content type
	if ct, _ := httputil.GetContentType(r); ct != "application/json" {
		httputil.WriteJSON(rw, httpBuildMessage(ErrUnsupportedType), http.StatusUnsupportedMediaType)
		return
	}

	// Get parse request from received data
	req := &parseRequest{}
	if err := httputil.BindJSON(r, req); err != nil {
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
			if err := w.service.backend.Get(req.ID, &collection); err == nil {
				for _, entry := range collection.Entries {
					paths = append(paths, w.prefix+"/library/graphs/"+entry.ID)
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

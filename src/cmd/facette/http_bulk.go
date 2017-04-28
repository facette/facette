package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/facette/httputil"
)

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

func (w *httpWorker) httpHandleBulk(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Check for request content type
	if ct, _ := httputil.GetContentType(r); ct != "application/json" {
		httputil.WriteJSON(rw, httpBuildMessage(ErrUnsupportedType), http.StatusUnsupportedMediaType)
		return
	}

	// Get search request from received data
	req := bulkRequest{}
	if err := httputil.BindJSON(r, &req); err != nil {
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

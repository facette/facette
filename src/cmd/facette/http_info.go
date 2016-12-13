package main

import (
	"context"
	"fmt"
	"net/http"
	"runtime"

	"github.com/facette/httputil"

	"facette/backend"
	"facette/connector"
)

type httpInfo struct {
	Version    string   `json:"version"`
	BuildDate  string   `json:"build_date"`
	BuildHash  string   `json:"build_hash"`
	Compiler   string   `json:"compiler"`
	Drivers    []string `json:"drivers"`
	Connectors []string `json:"connectors"`
}

func (w *httpWorker) httpHandleInfo(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Get service information
	result := httpInfo{
		Version:    version,
		BuildDate:  buildDate,
		BuildHash:  buildHash,
		Compiler:   fmt.Sprintf("%s (%s)", runtime.Version(), runtime.Compiler),
		Drivers:    backend.Drivers(),
		Connectors: connector.Connectors(),
	}

	httputil.WriteJSON(rw, result, http.StatusOK)
}

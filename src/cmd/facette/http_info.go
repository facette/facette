package main

import (
	"fmt"
	"net/http"
	"runtime"

	"facette/connector"

	"github.com/facette/httputil"
	"github.com/facette/sqlstorage"
)

type httpInfo struct {
	Version    string   `json:"version,omitempty"`
	BuildDate  string   `json:"build_date,omitempty"`
	BuildHash  string   `json:"build_hash,omitempty"`
	Compiler   string   `json:"compiler,omitempty"`
	Drivers    []string `json:"drivers"`
	Connectors []string `json:"connectors"`
	ReadOnly   bool     `json:"read_only,omitempty"`
}

func (w *httpWorker) httpHandleInfo(rw http.ResponseWriter, r *http.Request) {
	var result httpInfo

	defer r.Body.Close()

	// Get service information
	if w.service.config.HideBuildDetails {
		result = httpInfo{
			Drivers:    sqlstorage.Drivers(),
			Connectors: connector.Connectors(),
			ReadOnly:   w.service.config.ReadOnly,
		}
	} else {
		result = httpInfo{
			Version:    version,
			BuildDate:  buildDate,
			BuildHash:  buildHash,
			Compiler:   fmt.Sprintf("%s (%s)", runtime.Version(), runtime.Compiler),
			Drivers:    sqlstorage.Drivers(),
			Connectors: connector.Connectors(),
			ReadOnly:   w.service.config.ReadOnly,
		}
	}

	httputil.WriteJSON(rw, result, http.StatusOK)
}

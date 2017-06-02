package main

import (
	"fmt"
	"net/http"
	"runtime"

	"facette/connector"

	"github.com/facette/httputil"
	"github.com/facette/sqlstorage"
)

// api:section info "Information"

type httpInfo struct {
	Version    string   `json:"version,omitempty"`
	BuildDate  string   `json:"build_date,omitempty"`
	BuildHash  string   `json:"build_hash,omitempty"`
	Compiler   string   `json:"compiler,omitempty"`
	Drivers    []string `json:"drivers"`
	Connectors []string `json:"connectors"`
	ReadOnly   bool     `json:"read_only,omitempty"`
}

// api:method GET /api/v1/ "Get service version and supported features"
//
// This endpoint returns the SQL storage drivers and catalog connectors supported by the Facette back-end.
//
// If the back-end is not configured to hide build information details, it will also return the detailed build
// information.
//
// ---
// section: info
// responses:
//   200:
//     type: object
//     example:
//       body: |
//         {
//           "version": "0.4.0",
//           "build_date": "2017-06-06",
//           "build_hash": "08794ed",
//           "compiler": "go1.8.3 (gc)",
//           "drivers": [
//             "mysql",
//             "pgsql",
//             "sqlite"
//           ],
//           "connectors": [
//             "facette",
//             "graphite",
//             "influxdb",
//             "kairosdb",
//             "rrd"
//           ]
//         }
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

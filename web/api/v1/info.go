package v1

import (
	"net/http"

	"facette.io/facette/connector"
	"facette.io/facette/version"
	"facette.io/httputil"
	"facette.io/jsonutil"
	"facette.io/sqlstorage"
)

// Info represents an API information instance.
type Info struct {
	Version    jsonutil.NullString `json:"version"`
	Branch     jsonutil.NullString `json:"branch"`
	Revision   jsonutil.NullString `json:"revision"`
	Compiler   string              `json:"compiler"`
	BuildDate  jsonutil.NullString `json:"build_date"`
	Drivers    []string            `json:"drivers"`
	Connectors []string            `json:"connectors"`
	ReadOnly   bool                `json:"read_only,omitempty"`
}

// api:section info "Information"

// api:method GET /api/v1 "Get service version and supported features"
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
//     examples:
//     - format: javascript
//       body: |
//         {
//           "version": "0.5.0",
//           "branch": "master",
//           "revision": "a1ad755fb940223d23098e7a71ca9a5252f87d9a",
//           "compiler": "go1.10.3 (gc)",
//           "build_date": "2018-07-22",
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
func (a *API) infoGet(rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	httputil.WriteJSON(rw, Info{
		Version:    jsonutil.NullString(version.Version),
		Branch:     jsonutil.NullString(version.Branch),
		Revision:   jsonutil.NullString(version.Revision),
		Compiler:   version.Compiler,
		BuildDate:  jsonutil.NullString(version.BuildDate),
		Drivers:    sqlstorage.Drivers(),
		Connectors: connector.Connectors(),
		ReadOnly:   a.config.ReadOnly,
	}, http.StatusOK)
}

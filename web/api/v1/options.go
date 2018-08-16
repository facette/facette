package v1

import (
	"net/http"

	"facette.io/facette/connector"
	"facette.io/httputil"
)

// Options represents an API options instance.
type Options struct {
	Connectors []string `json:"connectors"`
	ReadOnly   bool     `json:"read_only"`
}

// api:section options "Options"

// api:method OPTIONS /api/v1 "Get service options"
//
// This endpoint returns the options associated with the service instance.
//
// ---
// section: options
// responses:
//   200:
//     type: object
//     examples:
//     - format: javascript
//       body: |
//         {
//           "connectors": [
//             "facette",
//             "graphite",
//             "influxdb",
//             "kairosdb",
//             "rrd"
//           ],
//           "read_only": false
//         }
func (a *API) optionsGet(rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	httputil.WriteJSON(rw, Options{
		Connectors: connector.Connectors(),
		ReadOnly:   a.config.HTTP.ReadOnly,
	}, http.StatusOK)
}

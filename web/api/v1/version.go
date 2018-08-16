package v1

import (
	"net/http"

	"facette.io/facette/version"
	"facette.io/httputil"
	"facette.io/jsonutil"
)

// Version represents an API version information instance.
type Version struct {
	Version   jsonutil.NullString `json:"version"`
	Branch    jsonutil.NullString `json:"branch"`
	Revision  jsonutil.NullString `json:"revision"`
	Compiler  string              `json:"compiler"`
	BuildDate jsonutil.NullString `json:"build_date"`
}

// api:section version "Version"

// api:method GET /api/v1/version "Get service version information"
//
// This endpoint returns the Facette service version information.
//
// If the service is configured to not expose the version information, the request will be rejected with
// `403 Forbidden`.
//
// ---
// section: version
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
//         }
func (a *API) versionGet(rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if !a.config.HTTP.ExposeVersion {
		rw.WriteHeader(http.StatusForbidden)
		return
	}

	httputil.WriteJSON(rw, Version{
		Version:   jsonutil.NullString(version.Version),
		Branch:    jsonutil.NullString(version.Branch),
		Revision:  jsonutil.NullString(version.Revision),
		Compiler:  version.Compiler,
		BuildDate: jsonutil.NullString(version.BuildDate),
	}, http.StatusOK)
}

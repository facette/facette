package v1

import (
	"net/http"

	"facette.io/httputil"
)

// api:section library "Library"

var libraryTypes = []string{
	"collections",
	"graphs",
	"sourcegroups",
	"metricgroups",
}

// api:method GET /api/v1/library "Get library summary"
//
// This endpoint returns library items count per type.
//
// ---
// section: library
// responses:
//   200:
//     type: object
//     examples:
//     - format: javascript
//       body: |
//         {
//           "collections": 1,
//           "graphs": 7,
//           "sourcegroups": 3,
//           "metricgroups": 42
//         }
func (a *API) librarySummary(rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Get item types list and information
	result := map[string]int{}
	for _, typ := range libraryTypes {
		item, _ := a.storageItem(typ)

		count, err := a.storage.SQL().Count(item)
		if err != nil {
			a.logger.Error("failed to fetch count: %s", err)
			httputil.WriteJSON(rw, newMessage(errUnhandledError), http.StatusInternalServerError)
			return
		}

		result[typ] = count
	}

	httputil.WriteJSON(rw, result, http.StatusOK)
}

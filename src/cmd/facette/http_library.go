package main

import (
	"net/http"

	"github.com/facette/httputil"
)

// api:section library "Library"

var libraryTypes = []string{
	"collections",
	"graphs",
	"sourcegroups",
	"metricgroups",
}

// api:method GET /api/v1/library/ "Get library summary"
//
// This endpoint returns library items count per type.
//
// ---
// section: library
// responses:
//   200:
//     type: object
//     example:
//       body: |
//         {
//           "collections": 1,
//           "graphs": 7,
//           "sourcegroups": 3,
//           "metricgroups": 42
//         }
func (w *httpWorker) httpHandleLibrarySummary(rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Get item types list and information
	result := map[string]int{}
	for _, typ := range libraryTypes {
		item, _ := w.httpBackendNewItem(typ)

		count, err := w.service.backend.Storage().Count(item)
		if err != nil {
			w.log.Error("failed to fetch count: %s", err)
			httputil.WriteJSON(rw, httpBuildMessage(ErrUnhandledError), http.StatusInternalServerError)
			return
		}

		result[typ] = count
	}

	httputil.WriteJSON(rw, result, http.StatusOK)
}

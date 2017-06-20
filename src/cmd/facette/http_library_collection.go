package main

import (
	"net/http"

	"github.com/facette/httproute"
	"github.com/facette/httputil"
)

// api:method GET /api/v1/library/collections/tree "Get collections tree"
//
// This endpoint renders the library collections tree.
//
// ---
// section: library
// parameters:
// - name: parent
//   type: string
//   description: parent node to generate the tree from
//   in: query
func (w *httpWorker) httpHandleLibraryCollectionTree(rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	tree, err := w.service.backend.NewCollectionTree(httproute.QueryParam(r, "parent"))
	if err != nil {
		w.log.Error("unable to get collections tree: %s", err)
		httputil.WriteJSON(rw, httpBuildMessage(ErrUnhandledError), http.StatusInternalServerError)
		return
	}

	httputil.WriteJSON(rw, tree, http.StatusOK)
}

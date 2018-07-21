package main

import (
	"net/http"

	"facette.io/facette/backend"
	"facette.io/httproute"
	"facette.io/httputil"
	"facette.io/sqlstorage"
	"github.com/hashicorp/go-uuid"
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

	root := httproute.QueryParam(r, "parent")
	if root != "" {
		var c backend.Collection

		// If provided parent value is not a valid UUID it is probably an alias, resolve it to get the actual UUID value
		if _, err := uuid.ParseUUID(root); err != nil {
			if err := w.service.backend.Storage().Get("alias", root, &c, false); err == sqlstorage.ErrItemNotFound {
				w.log.Error("unable to get collections tree: %s", err)
				httputil.WriteJSON(rw, httpBuildMessage(backend.ErrInvalidAlias), http.StatusBadRequest)
				return
			}

			root = c.ID
		}
	}

	tree, err := w.service.backend.NewCollectionTree(root)
	if err != nil {
		w.log.Error("unable to get collections tree: %s", err)
		httputil.WriteJSON(rw, httpBuildMessage(ErrUnhandledError), http.StatusInternalServerError)
		return
	}

	httputil.WriteJSON(rw, tree, http.StatusOK)
}

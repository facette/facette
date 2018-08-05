package v1

import (
	"net/http"

	"facette.io/facette/backend"
	"facette.io/httputil"
	"facette.io/sqlstorage"
	"github.com/hashicorp/go-uuid"
	"github.com/vbatoufflet/httproute"
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
func (a *API) libraryCollectionTree(rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	root := httproute.QueryParam(r, "parent")
	if root != "" {
		var c backend.Collection

		// If provided parent value is not a valid UUID it is probably an alias, resolve it to get the actual UUID value
		if _, err := uuid.ParseUUID(root); err != nil {
			if err := a.backend.Storage().Get("alias", root, &c, false); err == sqlstorage.ErrItemNotFound {
				a.logger.Error("unable to get collections tree: %s", err)
				httputil.WriteJSON(rw, newMessage(backend.ErrInvalidAlias), http.StatusBadRequest)
				return
			}

			root = c.ID
		}
	}

	tree, err := a.backend.NewCollectionTree(root)
	if err != nil {
		a.logger.Error("unable to get collections tree: %s", err)
		httputil.WriteJSON(rw, newMessage(errUnhandledError), http.StatusInternalServerError)
		return
	}

	httputil.WriteJSON(rw, tree, http.StatusOK)
}

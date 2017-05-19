package main

import (
	"net/http"

	"github.com/facette/httputil"
)

func (w *httpWorker) httpHandleLibraryCollectionTree(rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	tree, err := w.service.backend.NewCollectionTree(r.URL.Query().Get("parent"))
	if err != nil {
		w.log.Error("unable to get collections tree: %s", err)
		httputil.WriteJSON(rw, httpBuildMessage(ErrUnhandledError), http.StatusInternalServerError)
		return
	}

	httputil.WriteJSON(rw, tree, http.StatusOK)
}

package main

import (
	"net/http"

	"facette/backend"

	"github.com/facette/httproute"
	"github.com/facette/httputil"
	"github.com/facette/sqlstorage"
)

func (w *httpWorker) httpHandleProviderRefresh(rw http.ResponseWriter, r *http.Request) {
	id := httproute.ContextParam(r, "id").(string)

	provider := backend.Provider{}

	// Request item from back-end
	if err := w.service.backend.Storage().Get("id", id, &provider); err == sqlstorage.ErrItemNotFound {
		httputil.WriteJSON(rw, httpBuildMessage(err), http.StatusNotFound)
		return
	} else if err != nil {
		w.log.Error("failed to fetch item: %s", err)
		httputil.WriteJSON(rw, httpBuildMessage(ErrUnhandledError), http.StatusInternalServerError)
		return
	}

	w.service.poller.Refresh(provider)

	httputil.WriteJSON(rw, nil, http.StatusNoContent)
}

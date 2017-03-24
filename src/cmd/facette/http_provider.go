package main

import (
	"context"
	"net/http"

	"facette/backend"

	"github.com/facette/httputil"
)

func (w *httpWorker) httpHandleProviderRefresh(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	id := ctx.Value("id").(string)

	provider := backend.Provider{}

	// Request item from backend
	if err := w.service.backend.Get(id, &provider); err == backend.ErrItemNotExist {
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

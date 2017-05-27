package main

import (
	"context"
	"net/http"

	"facette/backend"

	"github.com/facette/httproute"
	"github.com/facette/httputil"
	"github.com/facette/sqlstorage"
)

func (w *httpWorker) httpHandleProviderCreate(rw http.ResponseWriter, r *http.Request) {
	w.httpHandleBackendCreate(rw, r.WithContext(context.WithValue(r.Context(), "type", "providers")))
}

func (w *httpWorker) httpHandleProviderGet(rw http.ResponseWriter, r *http.Request) {
	w.httpHandleBackendGet(rw, r.WithContext(context.WithValue(r.Context(), "type", "providers")))
}

func (w *httpWorker) httpHandleProviderUpdate(rw http.ResponseWriter, r *http.Request) {
	w.httpHandleBackendUpdate(rw, r.WithContext(context.WithValue(r.Context(), "type", "providers")))
}

func (w *httpWorker) httpHandleProviderDelete(rw http.ResponseWriter, r *http.Request) {
	w.httpHandleBackendDelete(rw, r.WithContext(context.WithValue(r.Context(), "type", "providers")))
}

func (w *httpWorker) httpHandleProviderDeleteAll(rw http.ResponseWriter, r *http.Request) {
	w.httpHandleBackendDeleteAll(rw, r.WithContext(context.WithValue(r.Context(), "type", "providers")))
}

func (w *httpWorker) httpHandleProviderList(rw http.ResponseWriter, r *http.Request) {
	w.httpHandleBackendList(rw, r.WithContext(context.WithValue(r.Context(), "type", "providers")))
}

func (w *httpWorker) httpHandleProviderRefresh(rw http.ResponseWriter, r *http.Request) {
	id := httproute.ContextParam(r, "id").(string)

	provider := backend.Provider{}

	// Request item from back-end
	if err := w.service.backend.Storage().Get("id", id, &provider, false); err == sqlstorage.ErrItemNotFound {
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

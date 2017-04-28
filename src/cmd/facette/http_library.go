package main

import (
	"context"
	"net/http"

	"github.com/facette/httputil"
)

var libraryTypes = []string{
	"collections",
	"graphs",
	"sourcegroups",
	"metricgroups",
}

func (w *httpWorker) httpHandleLibraryRoot(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
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

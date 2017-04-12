package main

import (
	"context"
	"net/http"
	"reflect"
	"sort"

	"github.com/facette/httputil"
	"github.com/fatih/set"
)

var libraryTypes = set.New(
	"collections",
	"graphs",
	"sourcegroups",
	"metricgroups",
)

func (w *httpWorker) httpHandleLibraryRoot(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Get item types list and information
	result := httpTypeList{}
	for typ, rt := range backendTypes {
		if !libraryTypes.Has(typ) {
			continue
		}

		rv := reflect.New(rt)

		count, err := w.service.backend.Count(rv.Interface(), nil)
		if err != nil {
			w.log.Error("failed to fetch count: %s", err)
			httputil.WriteJSON(rw, httpBuildMessage(ErrUnhandledError), http.StatusInternalServerError)
			return
		}

		result = append(result, httpTypeRecord{
			Name:  typ,
			Count: count,
		})
	}

	sort.Sort(result)

	httputil.WriteJSON(rw, result, http.StatusOK)
}

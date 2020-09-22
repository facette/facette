// Copyright (c) 2020, The Facette Authors
//
// Licensed under the terms of the BSD 3-Clause License; a copy of the license
// is available at: https://opensource.org/licenses/BSD-3-Clause

// Package api provides version 1 of the API endpoints.
package api

import (
	"net/http"

	"batou.dev/httprouter"
	"go.uber.org/zap"

	"facette.io/facette/pkg/api"
	"facette.io/facette/pkg/catalog"
	"facette.io/facette/pkg/errors"
	httpjson "facette.io/facette/pkg/http/json"
	"facette.io/facette/pkg/poller"
	"facette.io/facette/pkg/store"
)

// Register initializes and registers API version 1 endpoints.
func Register(
	router *httprouter.Router,
	catalog *catalog.Catalog,
	store *store.Store,
	poller *poller.Poller,
) {
	h := handler{
		router:  router,
		catalog: catalog,
		store:   store,
		poller:  poller,
		log:     zap.L().Named("http/server"),
	}

	root := router.Endpoint(api.Prefix).
		Use(noCache).
		Options(h.getOptions)

	root.Endpoint("/bulk").
		Post(h.ExecBulk)

	endpoint := root.Endpoint("/charts").Use(withType("charts"))
	endpoint.
		Get(h.ListObjects).
		Post(h.SaveObject)
	endpoint.Endpoint("/:id").
		Delete(h.DeleteObject).
		Get(h.GetObject).
		Patch(h.SaveObject).
		Put(h.SaveObject)
	endpoint.Endpoint("/:id/resolve").
		Post(h.ResolveObject)
	endpoint.Endpoint("/:id/vars").
		Get(h.GetObjectVars)

	endpoint = root.Endpoint("/dashboards").Use(withType("dashboards"))
	endpoint.
		Get(h.ListObjects).
		Post(h.SaveObject)
	endpoint.Endpoint("/:id").
		Delete(h.DeleteObject).
		Get(h.GetObject).
		Patch(h.SaveObject).
		Put(h.SaveObject)
	endpoint.Endpoint("/:id/resolve").
		Post(h.ResolveObject)
	endpoint.Endpoint("/:id/vars").
		Get(h.GetObjectVars)

	endpoint = root.Endpoint("/labels")
	endpoint.
		Get(h.ListLabels)
	endpoint.Endpoint("/values").
		Get(h.ListValues)

	root.Endpoint("/metrics").
		Get(h.ListMetrics)

	endpoint = root.Endpoint("/providers").Use(withType("providers"))
	endpoint.
		Get(h.ListObjects).
		Post(h.SaveObject)
	endpoint.Endpoint("/poll").
		Post(h.PollProvider)
	endpoint.Endpoint("/test").
		Post(h.TestProvider)
	endpoint.Endpoint("/:id").
		Delete(h.DeleteObject).
		Get(h.GetObject).
		Patch(h.SaveObject).
		Put(h.SaveObject)
	endpoint.Endpoint("/:id/poll").
		Post(h.PollProvider)

	root.Endpoint("/query").
		Post(h.ExecQuery)

	endpoint = root.Endpoint("/store")
	endpoint.
		Get(h.DumpStore).
		Post(h.RestoreStore)

	root.Endpoint("/version").
		Get(h.GetVersion)

	root.Endpoint("/*").
		Any(notFound)
}

func noCache(h http.Handler) http.Handler {
	// Ensure that HTTP API calls are not cached by clients
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
		h.ServeHTTP(rw, r)
	})
}

func notFound(rw http.ResponseWriter, r *http.Request) {
	httpjson.Write(rw, responseFromError(api.ErrNotFound), http.StatusNotFound)
}

func withType(name string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(rw, httprouter.SetContextParam(r, "type", name))
		})
	}
}

type handler struct {
	router  *httprouter.Router
	catalog *catalog.Catalog
	store   *store.Store
	poller  *poller.Poller
	log     *zap.Logger
}

func (h handler) WriteError(rw http.ResponseWriter, err error) {
	v := errors.Unwrap(err)
	if v == nil {
		v = err
	}

	switch v {
	case api.ErrConflict:
		httpjson.Write(rw, responseFromError(err), http.StatusConflict)

	case api.ErrInvalid:
		httpjson.Write(rw, responseFromError(err), http.StatusBadRequest)

	case api.ErrNotFound:
		httpjson.Write(rw, responseFromError(err), http.StatusNotFound)

	case api.ErrUnsupportedType:
		httpjson.Write(rw, responseFromError(err), http.StatusUnsupportedMediaType)

	default:
		zap.L().WithOptions(zap.AddCallerSkip(1)).Error(err.Error())
		httpjson.Write(rw, responseFromError(api.ErrUnhandled), http.StatusInternalServerError)
	}
}

func responseFromError(err error) api.Response {
	return api.Response{Error: err.Error()}
}

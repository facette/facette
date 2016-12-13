package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"text/template/parse"
	"time"

	"facette/backend"

	"github.com/facette/httputil"
	"github.com/facette/jsonutil"
	uuid "github.com/hashicorp/go-uuid"
)

var backendTypes = map[string]reflect.Type{
	"providers":    reflect.TypeOf(backend.Provider{}),
	"collections":  reflect.TypeOf(backend.Collection{}),
	"graphs":       reflect.TypeOf(backend.Graph{}),
	"sourcegroups": reflect.TypeOf(backend.SourceGroup{}),
	"metricgroups": reflect.TypeOf(backend.MetricGroup{}),
	"scales":       reflect.TypeOf(backend.Scale{}),
	"units":        reflect.TypeOf(backend.Unit{}),
}

func (w *httpWorker) httpHandleBackendCreate(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	if ct, _ := httputil.GetContentType(r); ct != "application/json" {
		httputil.WriteJSON(rw, httpBuildMessage(ErrUnsupportedType), http.StatusUnsupportedMediaType)
		return
	}

	typ := ctx.Value("type").(string)

	// Get backend item type
	rt, ok := backendTypes[typ]
	if !ok {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	rv := reflect.New(rt)

	// Check for existing item properties inheritance
	if id := r.URL.Query().Get("inherit"); id != "" {
		err := w.service.backend.Get(id, rv.Interface())
		if err == backend.ErrItemNotExist {
			httputil.WriteJSON(rw, httpBuildMessage(err), http.StatusNotFound)
			return
		} else if err != nil {
			w.log.Error("failed to fetch item: %s", err)
			httputil.WriteJSON(rw, httpBuildMessage(ErrUnhandledError), http.StatusInternalServerError)
			return
		}

		reflect.Indirect(rv).FieldByName("ID").SetString("")
	}

	// Fill item with data received from request
	if err := httputil.BindJSON(r, rv.Interface()); err != nil {
		w.log.Error("unable to unmarshal JSON data: %s", err)
		httputil.WriteJSON(rw, httpBuildMessage(ErrInvalidJSON), http.StatusBadRequest)
		return
	}

	// Parse body for template keys potential errors
	if typ == "collections" || typ == "graphs" {
		if reflect.Indirect(rv).FieldByName("Template").Bool() {
			data, _ := json.Marshal(rv.Interface())

			if _, err := parse.Parse("inline", string(data), "", ""); err != nil {
				w.log.Error("failed to parse template: %s", err)
				httputil.WriteJSON(rw, httpBuildMessage(ErrInvalidTemplate), http.StatusBadRequest)
				return
			}
		}
	}

	// Insert item into backend
	id, err := uuid.GenerateUUID()
	if err != nil {
		w.log.Error("failed to generate identifier: %s", err)
		httputil.WriteJSON(rw, httpBuildMessage(ErrUnhandledError), http.StatusInternalServerError)
	}

	reflect.Indirect(rv).FieldByName("ID").SetString(id)
	reflect.Indirect(rv).FieldByName("Created").Set(reflect.ValueOf(time.Now().UTC()))

	if typ == "providers" {
		// Set provider enabled by default
		reflect.Indirect(rv).FieldByName("Enabled").SetBool(true)
	}

	if err := w.service.backend.Add(rv.Interface()); err != nil {
		switch err {
		case backend.ErrResourceConflict:
			httputil.WriteJSON(rw, httpBuildMessage(err), http.StatusConflict)

		case backend.ErrEmptyGraph, backend.ErrEmptyGroup, backend.ErrExtraAttributes, backend.ErrInvalidName,
			backend.ErrInvalidParent, backend.ErrInvalidScale, backend.ErrInvalidUnit,
			backend.ErrResourceMissingData:
			httputil.WriteJSON(rw, httpBuildMessage(err), http.StatusBadRequest)

		case backend.ErrResourceMissingDependency:
			httputil.WriteJSON(rw, httpBuildMessage(err), http.StatusNotFound)

		default:
			w.log.Error("failed to insert item: %s", err)
			httputil.WriteJSON(rw, httpBuildMessage(ErrUnhandledError), http.StatusInternalServerError)
		}

		return
	}

	w.log.Debug("inserted %s item into backend", id)

	// Start new provider on new registration
	if typ == "providers" {
		go w.service.poller.StartProvider(reflect.Indirect(rv).Interface().(backend.Provider))
	}

	http.Redirect(rw, r, strings.TrimRight(r.URL.Path, "/")+"/"+id, http.StatusCreated)
}

func (w *httpWorker) httpHandleBackendDelete(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	typ := ctx.Value("type").(string)
	id := ctx.Value("id").(string)

	// Get backend item type
	rt, ok := backendTypes[typ]
	if !ok {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	rv := reflect.New(rt)
	reflect.Indirect(rv).FieldByName("ID").SetString(id)

	// Delete item from backend
	v := reflect.Indirect(rv).Interface()

	err := w.service.backend.Delete(v)
	if err == backend.ErrItemNotExist {
		httputil.WriteJSON(rw, httpBuildMessage(err), http.StatusNotFound)
		return
	} else if err != nil {
		w.log.Error("failed to delete item: %s", err)
		httputil.WriteJSON(rw, httpBuildMessage(ErrUnhandledError), http.StatusInternalServerError)
		return
	}

	w.log.Debug("deleted %s item from backend", id)

	// Stop provider on deletion
	if typ == "providers" {
		go w.service.poller.StopProvider(reflect.Indirect(rv).Interface().(backend.Provider), false)
	}

	rw.WriteHeader(http.StatusNoContent)
}

func (w *httpWorker) httpHandleBackendDeleteAll(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	typ := ctx.Value("type").(string)

	// Check backend item type
	rt, ok := backendTypes[typ]
	if !ok {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	// Check for confirmation header
	if r.Header.Get("X-Confirm-Action") != "1" {
		rw.WriteHeader(http.StatusForbidden)
		return
	}

	w.service.backend.Reset(reflect.New(rt).Interface())

	w.log.Debug("deleted %s from backend", typ)

	rw.WriteHeader(http.StatusNoContent)
}

func (w *httpWorker) httpHandleBackendGet(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	typ := ctx.Value("type").(string)
	id := ctx.Value("id").(string)

	fields := httpGetListParam(r, "fields", nil)

	// Get backend item type
	rt, ok := backendTypes[typ]
	if !ok {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	rv := reflect.New(rt)

	// Request item from backend
	if err := w.service.backend.Get(id, rv.Interface()); err == backend.ErrItemNotExist {
		httputil.WriteJSON(rw, httpBuildMessage(err), http.StatusNotFound)
		return
	} else if err != nil {
		w.log.Error("failed to fetch item: %s", err)
		httputil.WriteJSON(rw, httpBuildMessage(ErrUnhandledError), http.StatusInternalServerError)
		return
	}

	// Handle collection expansion request
	if typ == "collections" && r.URL.Query().Get("expand") == "1" {
		collection := rv.Interface().(*backend.Collection)

		if collection.Link != nil && len(collection.Attributes) > 0 {
			if err := collection.Link.Expand(collection.Attributes, w.service.backend); err != nil {
				w.log.Warning("failed to expand template: %s", err)
				httputil.WriteJSON(rw, httpBuildMessage(ErrUnhandledError), http.StatusInternalServerError)
				return
			}

			rv = reflect.ValueOf(collection.Link)

			collection = rv.Interface().(*backend.Collection)
			collection.Attributes = nil
		} else {
			for _, entry := range collection.Entries {
				entry.Attributes.Merge(collection.Attributes, true)
			}
		}
	}

	result := jsonutil.FilterStruct(rv.Interface(), fields)

	httputil.WriteJSON(rw, result, http.StatusOK)
}

func (w *httpWorker) httpHandleBackendList(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	typ := ctx.Value("type").(string)

	// Check backend item type
	rt, ok := backendTypes[typ]
	if !ok {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	// Check for list filter
	filters := make(map[string]interface{})

	if v := r.URL.Query().Get("filter"); v != "" {
		filters["name"] = v
	}

	if typ == "collections" || typ == "graphs" {
		switch r.URL.Query().Get("kind") {
		case "raw":
			filters["template"] = false

		case "template":
			filters["template"] = true

		case "all", "":
			// no filtering

		default:
			httputil.WriteJSON(rw, httpBuildMessage(ErrInvalidParameter), http.StatusBadRequest)
			return
		}

		if typ == "collections" {
			if p := r.URL.Query().Get("parent"); p != "" {
				filters["parent"] = p
			}
		}
	}

	// Request items list from backend
	rv := reflect.New(reflect.MakeSlice(reflect.SliceOf(rt), 0, 0).Type())

	offset, err := httpGetIntParam(r, "offset")
	if err != nil {
		httputil.WriteJSON(rw, httpBuildMessage(ErrInvalidParameter), http.StatusBadRequest)
		return
	}

	limit, err := httpGetIntParam(r, "limit")
	if err != nil {
		httputil.WriteJSON(rw, httpBuildMessage(ErrInvalidParameter), http.StatusBadRequest)
		return
	}

	sort := httpGetListParam(r, "sort", []string{"name"})

	count, err := w.service.backend.List(rv.Interface(), filters, sort, offset, limit)
	if err == backend.ErrUnknownColumn {
		httputil.WriteJSON(rw, httpBuildMessage(ErrInvalidParameter), http.StatusBadRequest)
		return
	} else if err != nil {
		w.log.Error("failed to fetch list: %s", err)
		httputil.WriteJSON(rw, httpBuildMessage(ErrUnhandledError), http.StatusInternalServerError)
		return
	}

	// Parse requested fields list or set defaults
	fields := httpGetListParam(r, "fields", nil)
	if fields == nil {
		fields = []string{"id", "name", "description", "created", "modified"}
		if typ == "providers" {
			fields = append(fields, "enabled")
		}
	}

	// Fill items list
	result := []map[string]interface{}{}

	n := reflect.Indirect(rv).Len()
	for i := 0; i < n; i++ {
		result = append(result, jsonutil.FilterStruct(reflect.Indirect(rv).Index(i).Interface(), fields))
	}

	rw.Header().Set("X-Total-Records", fmt.Sprintf("%d", count))
	httputil.WriteJSON(rw, result, http.StatusOK)
}

func (w *httpWorker) httpHandleBackendUpdate(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	if ct, _ := httputil.GetContentType(r); ct != "application/json" {
		httputil.WriteJSON(rw, httpBuildMessage(ErrUnsupportedType), http.StatusUnsupportedMediaType)
		return
	}

	typ := ctx.Value("type").(string)
	id := ctx.Value("id").(string)

	// Get backend item type
	rt, ok := backendTypes[typ]
	if !ok {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	rv := reflect.New(rt)

	// Retrieve existing element if patching
	if r.Method == "PATCH" {
		if err := w.service.backend.Get(id, rv.Interface()); err == backend.ErrItemNotExist {
			httputil.WriteJSON(rw, httpBuildMessage(err), http.StatusNotFound)
			return
		} else if err != nil {
			w.log.Error("failed to fetch item: %s", err)
			httputil.WriteJSON(rw, httpBuildMessage(ErrUnhandledError), http.StatusInternalServerError)
			return
		}
	}

	// Fill item with data received from request
	if err := httputil.BindJSON(r, rv.Interface()); err != nil {
		w.log.Error("unable to unmarshal JSON data: %s", err)
		httputil.WriteJSON(rw, httpBuildMessage(ErrInvalidJSON), http.StatusBadRequest)
		return
	}

	// Parse body for template keys potential errors
	if typ == "collections" || typ == "graphs" {
		if reflect.Indirect(rv).FieldByName("Template").Bool() {
			data, _ := json.Marshal(rv.Interface())

			if _, err := parse.Parse("inline", string(data), "", ""); err != nil {
				w.log.Error("failed to parse template: %s", err)
				httputil.WriteJSON(rw, httpBuildMessage(ErrInvalidTemplate), http.StatusBadRequest)
				return
			}
		}
	}

	// Set minimal fields values
	now := time.Now().UTC()

	reflect.Indirect(rv).FieldByName("ID").SetString(id)
	reflect.Indirect(rv).FieldByName("Modified").Set(reflect.ValueOf(&now))

	// Update item in backend
	if err := w.service.backend.Add(rv.Interface()); err != nil {
		switch err {
		case backend.ErrItemNotExist:
			httputil.WriteJSON(rw, httpBuildMessage(err), http.StatusNotFound)

		case backend.ErrEmptyGraph, backend.ErrEmptyGroup, backend.ErrExtraAttributes, backend.ErrInvalidName,
			backend.ErrInvalidParent, backend.ErrInvalidScale, backend.ErrInvalidUnit,
			backend.ErrResourceMissingData:
			httputil.WriteJSON(rw, httpBuildMessage(err), http.StatusBadRequest)

		case backend.ErrResourceConflict, backend.ErrResourceMissingDependency:
			httputil.WriteJSON(rw, httpBuildMessage(err), http.StatusConflict)

		default:
			w.log.Error("failed to update item: %s", err)
			httputil.WriteJSON(rw, httpBuildMessage(ErrUnhandledError), http.StatusInternalServerError)
		}

		return
	}

	w.log.Debug("updated %s item from backend", id)

	// Restart provider on update
	if typ == "providers" {
		if err := w.service.backend.Get(id, rv.Interface()); err == nil {
			go w.service.poller.StopProvider(reflect.Indirect(rv).Interface().(backend.Provider), true)
		}
	}

	rw.WriteHeader(http.StatusNoContent)
}

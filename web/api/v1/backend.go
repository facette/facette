package v1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"facette.io/facette/backend"
	"facette.io/facette/template"
	"facette.io/httputil"
	"facette.io/jsonutil"
	"facette.io/sqlstorage"
	"github.com/hashicorp/go-uuid"
	"github.com/vbatoufflet/httproute"
)

var backendTypes = []string{
	"providers",
	"collections",
	"graphs",
	"sourcegroups",
	"metricgroups",
}

// api:method POST /api/v1/library/:type "Create a library item"
//
// This endpoint creates a new item and stores it to the back-end database.
//
// The `inherit` query parameter can be used to inherit fields from an existing item, then applying new values with
// received body payload.
//
// If the instance is *read-only* the operation will be rejected with `403 Forbidden`.
//
// ---
// section: library
// parameters:
// - name: type
//   type: string
//   description: type of library items
//   required: true
//   in: path
// - name: inherit
//   type: string
//   description: identifier of the item to inherit from
//   in: query
// responses:
//   201:
func (a *API) backendCreate(rw http.ResponseWriter, r *http.Request) {
	if a.config.ReadOnly {
		httputil.WriteJSON(rw, newMessage(errReadOnly), http.StatusForbidden)
		return
	}

	typ := httproute.ContextParam(r, "type").(string)

	// Initialize new back-end item
	item, ok := a.backendItem(typ)
	if !ok {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	// Retrieve existing item data from back-end if inheriting
	rv := reflect.ValueOf(item)

	if id := httproute.QueryParam(r, "inherit"); id != "" {
		if err := a.backend.Storage().Get("id", id, rv.Interface(), false); err == sqlstorage.ErrItemNotFound {
			httputil.WriteJSON(rw, newMessage(err), http.StatusNotFound)
			return
		} else if err != nil {
			a.logger.Error("failed to fetch item for deletion: %s", err)
			httputil.WriteJSON(rw, newMessage(errUnhandledError), http.StatusInternalServerError)
			return
		}

		for _, name := range []string{"ID", "Created", "Modifed", "Alias"} {
			if f := reflect.Indirect(rv).FieldByName(name); f.IsValid() {
				f.Set(reflect.Zero(f.Type()))
			}
		}
	}

	// Fill item with data received from request
	if err := httputil.BindJSON(r, rv.Interface()); err == httputil.ErrInvalidContentType {
		httputil.WriteJSON(rw, newMessage(err), http.StatusUnsupportedMediaType)
		return
	} else if err != nil {
		a.logger.Error("unable to unmarshal JSON data: %s", err)
		httputil.WriteJSON(rw, newMessage(errInvalidJSON), http.StatusBadRequest)
		return
	}

	// Parse body for template keys potential errors
	if typ == "collections" || typ == "graphs" {
		if reflect.Indirect(rv).FieldByName("Template").Bool() {
			data, _ := json.Marshal(rv.Interface())

			if _, err := template.Parse(string(data)); err != nil {
				a.logger.Error("failed to parse template: %s", err)
				httputil.WriteJSON(rw, newMessage(template.ErrInvalidTemplate), http.StatusBadRequest)
				return
			}
		}
	}

	// Set provider enabled by default
	if typ == "providers" {
		reflect.Indirect(rv).FieldByName("Enabled").SetBool(true)
	}

	// Insert item into back-end
	if err := a.backend.Storage().Save(rv.Interface()); err != nil {
		switch err {
		case sqlstorage.ErrItemConflict:
			httputil.WriteJSON(rw, newMessage(err), http.StatusConflict)

		case backend.ErrInvalidAlias, backend.ErrInvalidID, backend.ErrInvalidName, backend.ErrInvalidPattern,
			sqlstorage.ErrMissingField, sqlstorage.ErrUnknownReference:
			httputil.WriteJSON(rw, newMessage(err), http.StatusBadRequest)

		default:
			a.logger.Error("failed to insert item: %s", err)
			httputil.WriteJSON(rw, newMessage(errUnhandledError), http.StatusInternalServerError)
		}

		return
	}

	id := reflect.Indirect(rv).FieldByName("ID").String()

	a.logger.Debug("inserted %q item into backend", id)

	// Start new provider upon creation
	if typ == "providers" {
		go a.poller.StartWorker(rv.Interface().(*backend.Provider))
	}

	http.Redirect(rw, r, strings.TrimRight(r.URL.Path, "/")+"/"+id, http.StatusCreated)
}

// api:method GET /api/v1/library/:type/:id "Get a library item"
//
// This endpoint returns a library item given its type and identifier.
//
// The `expand` query parameter _(available for collections and graphs)_ can be set to request item expansion. If the
// item is an instance of a template, all internal references will be resolved.
//
// ---
// section: library
// parameters:
// - name: type
//   type: string
//   description: type of library items
//   required: true
//   in: path
// - name: id
//   type: string
//   description: identifier of the item
//   required: true
//   in: path
// - name: expand
//   type: boolean
//   description: item expansion flag
//   in: query
// responses:
//   200:
//     type: object
//     examples:
//     - format: javascript
//       body: |
//         {
//           "id": "eccd09c3-aaa9-592b-ad55-3d92b4acf119",
//           "name": "load",
//           "description": "Load average for \"{{ .source }}\"",
//           "created": "2017-05-19T15:08:39Z",
//           "modified": "2017-06-14T06:17:46Z",
//           "groups": [
//             {
//               "name": "",
//               "operator": 0,
//               "consolidate": 1,
//               "series": [
//                 {
//                   "name": "shortterm",
//                   "origin": "{{ .origin }}",
//                   "source": "{{ .source }}",
//                   "metric": "load.shortterm",
//                   "options": {
//                     "color": "#fff726"
//                   }
//                 }
//               ]
//             },
//             {
//               "name": "",
//               "operator": 0,
//               "consolidate": 1,
//               "series": [
//                 {
//                   "name": "midterm",
//                   "origin": "{{ .origin }}",
//                   "source": "{{ .source }}",
//                   "metric": "load.midterm",
//                   "options": {
//                     "color": "#ff602a"
//                   }
//                 }
//               ]
//             },
//             {
//               "name": "",
//               "operator": 0,
//               "consolidate": 1,
//               "series": [
//                 {
//                   "name": "longterm",
//                   "origin": "{{ .origin }}",
//                   "source": "{{ .source }}",
//                   "metric": "load.longterm",
//                   "options": {
//                     "color": "#be1732"
//                   }
//                 }
//               ]
//             }
//           ],
//           "options": {
//             "title": "{{ .source }} - Load Average",
//             "type": "line",
//             "yaxis_unit": "fixed"
//           },
//           "template": true
//         }
func (a *API) backendGet(rw http.ResponseWriter, r *http.Request) {
	var result interface{}

	typ := httproute.ContextParam(r, "type").(string)
	id := httproute.ContextParam(r, "id").(string)

	// Initialize new back-end item
	item, ok := a.backendItem(typ)
	if !ok {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	// Check for aliased item if identifier value isn't valid
	column := "id"
	if typ == "collections" || typ == "graphs" {
		if _, err := uuid.ParseUUID(id); err != nil {
			column = "alias"
		}
	}

	// Request item from back-end
	rv := reflect.ValueOf(item)

	if err := a.backend.Storage().Get(column, id, rv.Interface(), true); err == sqlstorage.ErrItemNotFound {
		httputil.WriteJSON(rw, newMessage(err), http.StatusNotFound)
		return
	} else if err != nil {
		a.logger.Error("failed to fetch item: %s", err)
		httputil.WriteJSON(rw, newMessage(errUnhandledError), http.StatusInternalServerError)
		return
	}

	// Handle collection expansion request
	if parseBoolParam(r, "expand") {
		if typ == "collections" {
			c := rv.Interface().(*backend.Collection)
			c.Expand(nil)
		} else if typ == "graphs" {
			g := rv.Interface().(*backend.Graph)
			g.Expand(nil)
		}
	}

	if fields := parseListParam(r, "fields", nil); fields != nil {
		result = jsonutil.FilterStruct(rv.Interface(), fields)
	} else {
		result = rv.Interface()
	}

	httputil.WriteJSON(rw, result, http.StatusOK)
}

// api:method PUT /api/v1/library/:type/:id "Update a library item"
//
// This endpoint updates a library item given its identifier. The request body is similar to the _Create a new library
// item_ endpoint.
//
// If the instance is *read-only* the operation will be rejected with `403 Forbidden`.
//
// ---
// section: library
// parameters:
// - name: type
//   type: string
//   description: type of library items
//   required: true
//   in: path
// - name: id
//   type: string
//   description: identifier of the item
//   required: true
//   in: path
// responses:
//   204:

// api:method PATCH /api/v1/library/:type/:id "Partially update a library item"
//
// This endpoint partially updates a library item given its identifier. The request body is similar to the _Update a
// library item_ endpoint, but only specified fields will be modified.
//
// If the instance is *read-only* the operation will be rejected with `403 Forbidden`.
//
// ---
// section: library
// parameters:
// - name: id
//   type: string
//   description: identifier of the provider
//   required: true
//   in: path
// responses:
//   204:
func (a *API) backendUpdate(rw http.ResponseWriter, r *http.Request) {
	if a.config.ReadOnly {
		httputil.WriteJSON(rw, newMessage(errReadOnly), http.StatusForbidden)
		return
	}

	typ := httproute.ContextParam(r, "type").(string)
	id := httproute.ContextParam(r, "id").(string)

	// Initialize new back-end item
	item, ok := a.backendItem(typ)
	if !ok {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	// Retrieve existing item data from back-end if patching
	rv := reflect.ValueOf(item)

	if r.Method == "PATCH" {
		if err := a.backend.Storage().Get("id", id, rv.Interface(), true); err == sqlstorage.ErrItemNotFound {
			httputil.WriteJSON(rw, newMessage(err), http.StatusNotFound)
			return
		} else if err != nil {
			a.logger.Error("failed to fetch item for deletion: %s", err)
			httputil.WriteJSON(rw, newMessage(errUnhandledError), http.StatusInternalServerError)
			return
		}
	} else {
		reflect.Indirect(rv).FieldByName("ID").SetString(id)
	}

	// Fill item with data received from request
	if err := httputil.BindJSON(r, rv.Interface()); err == httputil.ErrInvalidContentType {
		httputil.WriteJSON(rw, newMessage(err), http.StatusUnsupportedMediaType)
		return
	} else if err != nil {
		a.logger.Error("unable to unmarshal JSON data: %s", err)
		httputil.WriteJSON(rw, newMessage(errInvalidJSON), http.StatusBadRequest)
		return
	}

	// Parse body for template keys potential errors
	if typ == "collections" || typ == "graphs" {
		if reflect.Indirect(rv).FieldByName("Template").Bool() {
			data, _ := json.Marshal(rv.Interface())

			if _, err := template.Parse(string(data)); err != nil {
				a.logger.Error("failed to parse template: %s", err)
				httputil.WriteJSON(rw, newMessage(template.ErrInvalidTemplate), http.StatusBadRequest)
				return
			}
		}
	}

	// Update item in back-end
	if err := a.backend.Storage().Save(rv.Interface()); err != nil {
		switch err {
		case sqlstorage.ErrItemConflict:
			httputil.WriteJSON(rw, newMessage(err), http.StatusConflict)

		case backend.ErrInvalidAlias, backend.ErrInvalidID, backend.ErrInvalidName, backend.ErrInvalidPattern,
			sqlstorage.ErrMissingField, sqlstorage.ErrUnknownReference:
			httputil.WriteJSON(rw, newMessage(err), http.StatusBadRequest)

		default:
			a.logger.Error("failed to update item: %s", err)
			httputil.WriteJSON(rw, newMessage(errUnhandledError), http.StatusInternalServerError)
		}

		return
	}

	a.logger.Debug("updated %s item from back-end", id)

	// Restart provider on update
	if typ == "providers" {
		if err := a.backend.Storage().Get("id", id, rv.Interface(), false); err == nil {
			go a.poller.StopWorker(rv.Interface().(*backend.Provider), true)
		}
	}

	rw.WriteHeader(http.StatusNoContent)
}

// api:method DELETE /api/v1/library/:type/:id "Delete a library item"
//
// This endpoint deletes a library item given its type and identifier.
//
// If the instance is *read-only* the operation will be rejected with `403 Forbidden`.
//
// ---
// section: library
// parameters:
// - name: type
//   type: string
//   description: type of library items
//   required: true
//   in: path
// - name: id
//   type: string
//   description: identifier of the item
//   required: true
//   in: path
// responses:
//   204:
func (a *API) backendDelete(rw http.ResponseWriter, r *http.Request) {
	if a.config.ReadOnly {
		httputil.WriteJSON(rw, newMessage(errReadOnly), http.StatusForbidden)
		return
	}

	typ := httproute.ContextParam(r, "type").(string)
	id := httproute.ContextParam(r, "id").(string)

	// Initialize new back-end item
	item, ok := a.backendItem(typ)
	if !ok {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	// Request item from back-end
	rv := reflect.ValueOf(item)

	if err := a.backend.Storage().Get("id", id, rv.Interface(), false); err == sqlstorage.ErrItemNotFound {
		httputil.WriteJSON(rw, newMessage(err), http.StatusNotFound)
		return
	} else if err != nil {
		a.logger.Error("failed to fetch item for deletion: %s", err)
		httputil.WriteJSON(rw, newMessage(errUnhandledError), http.StatusInternalServerError)
		return
	}

	// Delete item from back-end
	err := a.backend.Storage().Delete(rv.Interface())
	if err == sqlstorage.ErrItemNotFound {
		httputil.WriteJSON(rw, newMessage(err), http.StatusNotFound)
		return
	} else if err != nil {
		a.logger.Error("failed to delete item: %s", err)
		httputil.WriteJSON(rw, newMessage(errUnhandledError), http.StatusInternalServerError)
		return
	}

	a.logger.Debug("deleted %s item from back-end", id)

	// Stop provider upon deletion
	if typ == "providers" {
		go a.poller.StopWorker(rv.Interface().(*backend.Provider), false)
	}

	rw.WriteHeader(http.StatusNoContent)
}

// api:method DELETE /api/v1/library/:type "Delete library items of a given type"
//
// This endpoint deletes all items of a given type.
//
// If the request header `X-Confirm-Action` is not present or if the instance is *read-only* the operation will be
// rejected with `403 Forbidden`.
//
// ---
// section: library
// parameters:
// - name: type
//   type: string
//   description: type of library items
//   required: true
//   in: path
// responses:
//   204:
func (a *API) backendDeleteAll(rw http.ResponseWriter, r *http.Request) {
	var rv reflect.Value

	if a.config.ReadOnly {
		httputil.WriteJSON(rw, newMessage(errReadOnly), http.StatusForbidden)
		return
	}

	typ := httproute.ContextParam(r, "type").(string)

	// Initialize new back-end item
	item, ok := a.backendItem(typ)
	if !ok {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	// Check for confirmation header
	if r.Header.Get("X-Confirm-Action") != "1" {
		rw.WriteHeader(http.StatusForbidden)
		return
	}

	// Request items list from back-end
	if typ == "providers" {
		rv = reflect.New(reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(item)), 0, 0).Type())

		_, err := a.backend.Storage().List(rv.Interface(), nil, nil, 0, 0, false)
		if err == sqlstorage.ErrUnknownColumn {
			httputil.WriteJSON(rw, newMessage(err), http.StatusBadRequest)
			return
		} else if err != nil {
			a.logger.Error("failed to fetch items for deletion: %s", err)
			httputil.WriteJSON(rw, newMessage(errUnhandledError), http.StatusInternalServerError)
			return
		}
	}

	a.backend.Storage().Delete(reflect.ValueOf(item).Interface())

	a.logger.Debug("deleted %s from back-end", typ)

	// Stop provider upon deletion
	if typ == "providers" {
		for i, n := 0, reflect.Indirect(rv).Len(); i < n; i++ {
			go a.poller.StopWorker(reflect.Indirect(rv).Index(i).Interface().(*backend.Provider), false)
		}
	}

	rw.WriteHeader(http.StatusNoContent)
}

// api:method GET /api/v1/library/:type "List library items of a given type"
//
// This endpoint returns library items of a given type. If a `filter` query parameter is given, only items having
// their name matching the filter will be returned.
//
// This endpoint supports pagination through the `offset` and `limit` query parameters and sorting using `sort` query
// parameter (separated by commas; prefix field name with "-" to reverse sort order).
//
// The `kind` query parameter _(available for collections and graphs)_ can be set in order to target or exclude
// templates from result:
//
//  * `all`: return all kind of items (default)
//  * `raw`: only return raw items, thus removing templates from result
//  * `template`: only return templates
//
// The `link` parameter _(available for collection and graphs)_ can be set in order to only return items having the
// given item as template reference.
//
// The `parent` query parameter _(only available for collections)_ can be set in order to only return items having the
// given collection for parent.
//
// ---
// section: library
// parameters:
// - name: type
//   type: string
//   description: type of library items
//   required: true
//   in: path
// - name: filter
//   type: string
//   description: term to filter names on
//   in: query
// - name: sort
//   type: string
//   description: fields to sort results on
//   in: query
// - name: offset
//   type: integer
//   description: offset to return items from
//   in: query
// - name: limit
//   type: integer
//   description: number of items to return
//   in: query
// - name: kind
//   type: string
//   description: kind of item to return
//   in: query
// - name: link
//   type: string
//   description: identifier of the linked item
//   in: query
// - name: parent
//   type: string
//   description: identifier of the parent item
//   in: query
// responses:
//   200:
//     type: array
//     examples:
//     - format: javascript
//       headers:
//         X-Total-Records: 3
//       body: |
//         [
//           {
//             "created": "2017-05-19T15:08:40Z",
//             "description": "CPU usage for \"{{ .source }}\"",
//             "id": "c1c5ba71-428a-565e-94e3-304c16e9a92f",
//             "modified": "2017-06-14T06:17:46Z",
//             "name": "cpu"
//           },
//           {
//             "created": "2017-05-19T15:08:39Z",
//             "description": "Disk usage for \"{{ .volume }}\" on \"{{ .source }}\"",
//             "id": "c77c2dae-b37f-5210-80b5-5d44ce5f7a97",
//             "modified": "2017-06-14T06:17:46Z",
//             "name": "df.bytes"
//           },
//           {
//             "created": "2017-05-19T15:08:39Z",
//             "description": "Load average for \"{{ .source }}\"",
//             "id": "eccd09c3-aaa9-592b-ad55-3d92b4acf119",
//             "modified": "2017-06-14T06:17:46Z",
//             "name": "load"
//           }
//         ]
func (a *API) backendList(rw http.ResponseWriter, r *http.Request) {
	typ := httproute.ContextParam(r, "type").(string)

	// Initialize new back-end item
	item, ok := a.backendItem(typ)
	if !ok {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	// Check for list filter
	filters := make(map[string]interface{})

	if v := httproute.QueryParam(r, "filter"); v != "" {
		filters["name"] = applyModifier(v)
	}

	if typ == "collections" || typ == "graphs" {
		switch httproute.QueryParam(r, "kind") {
		case "raw":
			filters["template"] = false

		case "template":
			filters["template"] = true

		case "all", "":
			// no filtering

		default:
			httputil.WriteJSON(rw, newMessage(errInvalidParameter), http.StatusBadRequest)
			return
		}

		if v := httproute.QueryParam(r, "link"); v != "" {
			filters["link"] = v
		}

		if typ == "collections" {
			if v := httproute.QueryParam(r, "parent"); v != "" {
				filters["parent"] = v
			}
		}
	}

	// Request items list from back-end
	rv := reflect.New(reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(item)), 0, 0).Type())

	offset, err := parseIntParam(r, "offset")
	if err != nil || offset < 0 {
		httputil.WriteJSON(rw, newMessage(errInvalidParameter), http.StatusBadRequest)
		return
	}

	limit, err := parseIntParam(r, "limit")
	if err != nil || limit < 0 {
		httputil.WriteJSON(rw, newMessage(errInvalidParameter), http.StatusBadRequest)
		return
	}

	sort := parseListParam(r, "sort", []string{"name"})

	count, err := a.backend.Storage().List(rv.Interface(), filters, sort, offset, limit, true)
	if err == sqlstorage.ErrUnknownColumn {
		httputil.WriteJSON(rw, newMessage(err), http.StatusBadRequest)
		return
	} else if err != nil {
		a.logger.Error("failed to fetch items: %s", err)
		httputil.WriteJSON(rw, newMessage(errUnhandledError), http.StatusInternalServerError)
		return
	}

	// Parse requested fields list or set defaults
	fields := parseListParam(r, "fields", nil)
	if fields == nil {
		fields = []string{"id", "name", "description", "created", "modified"}
		if typ == "providers" {
			fields = append(fields, "enabled")
		}
	}

	// Fill items list
	result := []map[string]interface{}{}

	for i, n := 0, reflect.Indirect(rv).Len(); i < n; i++ {
		if typ == "collections" && parseBoolParam(r, "expand") {
			collection := reflect.Indirect(rv).Index(i).Interface().(*backend.Collection)
			collection.Expand(nil)

			result = append(result, jsonutil.FilterStruct(collection, fields))
		} else {
			result = append(result, jsonutil.FilterStruct(reflect.Indirect(rv).Index(i).Interface(), fields))
		}
	}

	rw.Header().Set("X-Total-Records", fmt.Sprintf("%d", count))
	httputil.WriteJSON(rw, result, http.StatusOK)
}

func (a *API) backendItem(typ string) (interface{}, bool) {
	switch typ {
	case "providers":
		return a.backend.NewProvider(), true

	case "collections":
		return a.backend.NewCollection(), true

	case "graphs":
		return a.backend.NewGraph(), true

	case "sourcegroups":
		return a.backend.NewSourceGroup(), true

	case "metricgroups":
		return a.backend.NewMetricGroup(), true

	}

	return nil, false
}

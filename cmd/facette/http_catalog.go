package main

import (
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"strings"

	"facette.io/facette/catalog"
	"facette.io/facette/pattern"
	"facette.io/httputil"
	"github.com/fatih/set"
	"github.com/vbatoufflet/httproute"
)

// api:section catalog "Catalog"
//
// The catalog contains entries of the following types:
//
//  * `origins`
//  * `sources`
//  * `metrics`
//

// api:method GET /api/v1/catalog/ "Get catalog summary"
//
// This endpoint returns catalog entries count per type.
//
// ---
// section: catalog
// responses:
//   200:
//     type: object
//     examples:
//     - format: javascript
//       body: |
//         {
//           "origins": 1,
//           "sources": 3,
//           "metrics": 42
//         }
func (w *httpWorker) httpHandleCatalogSummary(rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Get item types list and information
	result := map[string]int{
		"origins": len(w.httpCatalogSearch("origins", "", r)),
		"sources": len(w.httpCatalogSearch("sources", "", r)),
		"metrics": len(w.httpCatalogSearch("metrics", "", r)),
	}

	httputil.WriteJSON(rw, result, http.StatusOK)
}

// api:method GET /api/v1/catalog/:type/ "List catalog entries of a given type"
//
// This endpoint returns catalog entries of a given type. If a `filter` query parameter is given, only entries having
// their name matching the filter will be returned.
//
// This endpoint supports pagination through the `offset` and `limit` query parameters.
//
// ---
// section: catalog
// parameters:
// - name: type
//   type: string
//   description: type of catalog entries
//   in: path
//   required: true
// - name: filter
//   type: string
//   description: term to filter names on
//   in: query
// - name: offset
//   type: integer
//   description: offset to return items from
//   in: query
// - name: limit
//   type: integer
//   description: number of items to return
//   in: query
// responses:
//   200:
//     type: array
//     headers:
//       X-Total-Records: total number of catalog records for this type
//     examples:
//     - headers:
//         X-Total-Records: 3
//       format: javascript
//       body: |
//         [
//           "metric1",
//           "metric2",
//           "metric3"
//         ]
func (w *httpWorker) httpHandleCatalogType(rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	typ := httproute.ContextParam(r, "type").(string)

	search := w.httpCatalogSearch(typ, "", r)
	if search == nil {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	// Fill result list
	filter := httproute.QueryParam(r, "filter")

	s := set.New()
	for _, item := range search {
		name := reflect.Indirect(reflect.ValueOf(item)).FieldByName("Name").String()
		if filter == "" {
			s.Add(name)
		} else if match, err := pattern.Match(filter, name); err != nil {
			httputil.WriteJSON(rw, httpBuildMessage(ErrInvalidFilter), http.StatusBadRequest)
			return
		} else if match {
			s.Add(name)
		}
	}

	total := s.Size()

	// Apply items list offset and limit
	result := set.StringSlice(s)
	sort.Strings(result)

	offset, err := httpGetIntParam(r, "offset")
	if err != nil || offset < 0 {
		httputil.WriteJSON(rw, httpBuildMessage(ErrInvalidParameter), http.StatusBadRequest)
		return
	}

	if offset < total {
		limit, err := httpGetIntParam(r, "limit")
		if err != nil || limit < 0 {
			httputil.WriteJSON(rw, httpBuildMessage(ErrInvalidParameter), http.StatusBadRequest)
			return
		}

		if limit != 0 && total > offset+limit {
			result = result[offset : offset+limit]
		} else if offset > 0 {
			result = result[offset:total]
		}
	} else {
		result = []string{}
	}

	rw.Header().Set("X-Total-Records", fmt.Sprintf("%d", total))
	httputil.WriteJSON(rw, result, http.StatusOK)
}

// api:method GET /api/v1/catalog/:type/:name "Get catalog entry information"
//
// This endpoint returns the information associated with a catalog entry given its type and name.
//
// ---
// section: catalog
// parameters:
// - name: type
//   type: string
//   description: type of catalog items
//   in: path
//   required: true
// - name: name
//   type: string
//   description: name of the catalog item
//   in: path
//   required: true
// responses:
//   200:
//     type: object
//     examples:
//     - format: javascript
//       body: |
//         {
//           "name": "metric3",
//           "origins": [
//             "provider1",
//           ],
//           "sources": [
//             "host1.example.net"
//           ],
//           "providers": [
//             "provider1",
//           ]
//         }
func (w *httpWorker) httpHandleCatalogEntry(rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var result interface{}

	typ := httproute.ContextParam(r, "type").(string)
	name := strings.TrimPrefix(r.URL.Path, w.prefix+"/catalog/"+typ+"/")

	search := w.httpCatalogSearch(typ, name, r)
	if search == nil || len(search) == 0 {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	switch typ {
	case "origins":
		item := struct {
			Name      string   `json:"name"`
			Providers []string `json:"providers"`
		}{}

		providers := set.New()
		for i, entry := range search {
			o := entry.(*catalog.Origin)
			if i == 0 {
				item.Name = o.Name
			}
			providers.Add(o.Catalog().Name())
		}

		item.Providers = set.StringSlice(providers)
		sort.Strings(item.Providers)

		result = item

	case "sources":
		item := struct {
			Name      string   `json:"name"`
			Origins   []string `json:"origins"`
			Providers []string `json:"providers"`
		}{}

		origins := set.New()
		providers := set.New()

		for i, entry := range search {
			s := entry.(*catalog.Source)
			if i == 0 {
				item.Name = s.Name
			}
			origins.Add(s.Origin().Name)
			providers.Add(s.Origin().Catalog().Name())
		}

		item.Origins = set.StringSlice(origins)
		item.Providers = set.StringSlice(providers)
		sort.Strings(item.Origins)
		sort.Strings(item.Providers)

		result = item

	case "metrics":
		item := struct {
			Name      string   `json:"name"`
			Origins   []string `json:"origins"`
			Sources   []string `json:"sources"`
			Providers []string `json:"providers"`
		}{}

		sources := set.New()
		origins := set.New()
		providers := set.New()

		for i, entry := range search {
			m := entry.(*catalog.Metric)
			if i == 0 {
				item.Name = m.Name
			}
			sources.Add(m.Source().Name)
			origins.Add(m.Source().Origin().Name)
			providers.Add(m.Source().Origin().Catalog().Name())
		}

		item.Sources = set.StringSlice(sources)
		item.Origins = set.StringSlice(origins)
		item.Providers = set.StringSlice(providers)
		sort.Strings(item.Sources)
		sort.Strings(item.Origins)
		sort.Strings(item.Providers)

		result = item
	}

	httputil.WriteJSON(rw, result, http.StatusOK)
}

func (w *httpWorker) httpCatalogSearch(typ, name string, r *http.Request) []interface{} {
	search := []interface{}{}

	switch typ {
	case "origins":
		for _, o := range w.service.searcher.Origins(
			name,
			-1,
		) {
			search = append(search, o)
		}

	case "sources":
		for _, s := range w.service.searcher.Sources(
			httproute.QueryParam(r, "origin"),
			name,
			-1,
		) {
			search = append(search, s)
		}

	case "metrics":
		for _, m := range w.service.searcher.Metrics(
			httproute.QueryParam(r, "origin"),
			httproute.QueryParam(r, "source"),
			name,
			-1,
		) {
			search = append(search, m)
		}

	default:
		return nil
	}

	return search
}

package main

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"sort"

	"facette/backend"
	"facette/catalog"

	"github.com/facette/httputil"
	"github.com/fatih/set"
)

func (w *httpWorker) httpHandleCatalogRoot(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Get item types list and information
	result := httpTypeList{
		httpTypeRecord{Name: "metrics", Count: len(w.httpCatalogSearch("metrics", "", r))},
		httpTypeRecord{Name: "origins", Count: len(w.httpCatalogSearch("origins", "", r))},
		httpTypeRecord{Name: "sources", Count: len(w.httpCatalogSearch("sources", "", r))},
	}

	httputil.WriteJSON(rw, result, http.StatusOK)
}

func (w *httpWorker) httpHandleCatalogType(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	typ := ctx.Value("type").(string)

	search := w.httpCatalogSearch(typ, "", r)
	if search == nil {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	// Fill result list
	filter := r.URL.Query().Get("filter")

	s := set.New()
	for _, item := range search {
		name := reflect.Indirect(reflect.ValueOf(item)).FieldByName("Name").String()
		if filter == "" || backend.FilterMatch(filter, name) {
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

func (w *httpWorker) httpHandleCatalogEntry(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var result interface{}

	typ := ctx.Value("type").(string)
	name := ctx.Value("name").(string)

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

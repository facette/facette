package main

import "net/http"

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
			r.URL.Query().Get("origin"),
			name,
			-1,
		) {
			search = append(search, s)
		}

	case "metrics":
		for _, m := range w.service.searcher.Metrics(
			r.URL.Query().Get("origin"),
			r.URL.Query().Get("source"),
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

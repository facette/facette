package httproute

import "net/http"

// ContextParam returns a request context parameter given its name.
func ContextParam(r *http.Request, key string) interface{} {
	return r.Context().Value(key)
}

// QueryParam returns a request query parameter given its name.
func QueryParam(r *http.Request, key string) string {
	return r.URL.Query().Get(key)
}

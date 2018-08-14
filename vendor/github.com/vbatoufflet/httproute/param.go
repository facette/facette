package httproute

import (
	"context"
	"net/http"
)

// ContextParam returns a request context parameter given its name.
func ContextParam(r *http.Request, key string) interface{} {
	return r.Context().Value(contextKey{key})
}

// QueryParam returns a request query parameter given its name.
func QueryParam(r *http.Request, key string) string {
	return r.URL.Query().Get(key)
}

// SetContextParam sets a new request context parameter.
func SetContextParam(r *http.Request, key string, value interface{}) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), contextKey{key}, value))
}

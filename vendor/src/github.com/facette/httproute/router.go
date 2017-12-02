package httproute

import (
	"net/http"
	"strings"
)

// Router represents an HTTP router instance.
type Router struct {
	endpoints       []*Endpoint
	middlewares     []func(http.Handler) http.Handler
	endpointHandler *endpointHandler
	chain           http.Handler
}

// NewRouter creates a new HTTP router instance.
func NewRouter() *Router {
	rt := &Router{
		endpoints:   []*Endpoint{},
		middlewares: []func(http.Handler) http.Handler{},
	}

	rt.endpointHandler = newEndpointHandler(rt)
	rt.chain = rt.endpointHandler

	return rt
}

// Endpoint creates a new HTTP router endpoint.
func (rt *Router) Endpoint(pattern string) *Endpoint {
	e := newEndpoint(pattern, rt)
	rt.endpoints = append(rt.endpoints, e)

	return e
}

// ServeHTTP satisfies 'http.Handler' interface requirements.
func (rt *Router) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path != "/" && strings.HasSuffix(path, "/") {
		path = strings.TrimSuffix(path, "/")
	}

	rt.chain.ServeHTTP(rw, r)
}

// Use registers a new middleware in the HTTP handlers chain.
func (rt *Router) Use(f func(http.Handler) http.Handler) *Router {
	rt.middlewares = append(rt.middlewares, f)
	rt.updateChain()

	return rt
}

// updateChain updates the middleware HTTP handlers chain.
func (rt *Router) updateChain() {
	rt.chain = rt.endpointHandler
	for i := len(rt.middlewares) - 1; i >= 0; i-- {
		rt.chain = rt.middlewares[i](rt.chain)
	}
}

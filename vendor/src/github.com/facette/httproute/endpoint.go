package httproute

import (
	"net/http"
	"sort"
	"strings"

	"context"
)

// Handler represents an HTTP request handler.
type Handler func(http.ResponseWriter, *http.Request)

// Endpoint represents an HTTP router endpoint.
type Endpoint struct {
	pattern  *pattern
	handlers map[string]Handler
	context  context.Context
	router   *Router
}

// newEndpoint creates a new HTTP enpdpoint instance.
func newEndpoint(pattern string, rt *Router) *Endpoint {
	return &Endpoint{
		pattern:  newPattern(pattern),
		handlers: make(map[string]Handler),
		router:   rt,
	}
}

// Any registers a handler for any method.
func (e *Endpoint) Any(h Handler) *Endpoint {
	return e.register("", h)
}

// Delete registers a 'DELETE' method handler.
func (e *Endpoint) Delete(h Handler) *Endpoint {
	return e.register("DELETE", h)
}

// Get registers a 'GET' method handler.
func (e *Endpoint) Get(h Handler) *Endpoint {
	return e.register("GET", h)
}

// Head registers a 'HEAD' method handler.
func (e *Endpoint) Head(h Handler) *Endpoint {
	return e.register("HEAD", h)
}

// Methods returns the list of methods available from the HTTP router endpoint.
func (e *Endpoint) Methods() []string {
	var hasGet, hasHead bool

	methods := []string{"OPTIONS"}
	for method := range e.handlers {
		methods = append(methods, method)

		if method == "GET" {
			hasGet = true
		} else if method == "HEAD" {
			hasHead = true
		}
	}

	if hasGet && !hasHead {
		methods = append(methods, "HEAD")
	}

	sort.Strings(methods)

	return methods
}

// Patch registers a 'PATCH' method handler.
func (e *Endpoint) Patch(h Handler) *Endpoint {
	return e.register("PATCH", h)
}

// Post registers a 'POST' method handler.
func (e *Endpoint) Post(h Handler) *Endpoint {
	return e.register("POST", h)
}

// Put registers a 'PUT' method handler.
func (e *Endpoint) Put(h Handler) *Endpoint {
	return e.register("PUT", h)
}

// handle wraps HTTP router endpoint handler with common internal logic.
func (e *Endpoint) handle(rw http.ResponseWriter, r *http.Request) {
	// Handle slash redirections
	if !e.pattern.hasWildcard {
		if e.pattern.hasSlash && !strings.HasSuffix(r.URL.Path, "/") {
			r.URL.Path += "/"
			http.Redirect(rw, r, r.URL.String(), http.StatusPermanentRedirect)
			return
		} else if !e.pattern.hasSlash && strings.HasSuffix(r.URL.Path, "/") {
			r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
			http.Redirect(rw, r, r.URL.String(), http.StatusPermanentRedirect)
			return
		}
	}

	// Check for requested method and handle defaults
	handler, ok := e.handlers[r.Method]
	if !ok {
		if _, ok = e.handlers[""]; ok {
			// Use 'Any' handler
			handler = e.handlers[""]
		} else {
			switch r.Method {
			case "HEAD":
				handler, ok = e.handlers["GET"]

			case "OPTIONS":
				rw.Header().Add("Allow", strings.Join(e.Methods(), ", "))
				rw.WriteHeader(http.StatusNoContent)
				return
			}
		}
	}

	if !ok {
		rw.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	handler(rw, r)
}

// register registers a new HTTP handler for a given method.
func (e *Endpoint) register(method string, handler Handler) *Endpoint {
	e.handlers[method] = handler
	return e
}

// endpointHandler represents an HTTP endpoint handler.
type endpointHandler struct {
	router *Router
}

// newEndpointHandler creates a new enpoint handler instance.
func newEndpointHandler(rt *Router) *endpointHandler {
	return &endpointHandler{rt}
}

// ServeHTTP satisfies 'http.Handler' interface requirements for the endpoint handler.
func (h *endpointHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	for _, endpoint := range h.router.endpoints {
		endpointCtx := endpoint.context
		if endpointCtx == nil {
			endpointCtx = context.Background()
		}

		if ctx, ok := endpoint.pattern.match(endpointCtx, r.URL.Path); ok {
			endpoint.handle(rw, r.WithContext(ctx))
			return
		}
	}

	rw.WriteHeader(http.StatusNotFound)
}

package httprouter

import (
	"net/http"
	"sort"
	"strings"
)

// Endpoint represents an HTTP router endpoint.
type Endpoint struct {
	pattern     *pattern
	handlers    map[string]http.HandlerFunc
	middlewares []func(http.Handler) http.Handler
	handler     *handler
	chain       http.Handler
	endpoints   []*Endpoint
	root        *Endpoint
}

func newEndpoint(pattern string) *Endpoint {
	e := &Endpoint{
		pattern:  newPattern(pattern),
		handlers: map[string]http.HandlerFunc{},
	}
	e.handler = newHandler(e)
	e.chain = e.handler

	if pattern == "" {
		e.root = e
	}

	return e
}

// Any registers a handler for any method.
func (e *Endpoint) Any(f http.HandlerFunc) *Endpoint {
	return e.register("", f)
}

// Delete registers a DELETE method handler.
func (e *Endpoint) Delete(f http.HandlerFunc) *Endpoint {
	return e.register("DELETE", f)
}

// Endpoint creates a new HTTP router endpoint.
func (e *Endpoint) Endpoint(pattern string) *Endpoint {
	endpoint := newEndpoint(strings.TrimRight(e.pattern.value, "/") + "/" + strings.TrimLeft(pattern, "/"))
	endpoint.root = e.root
	endpoint.middlewares = e.middlewares[:]
	endpoint.updateChain()

	e.root.endpoints = append(e.root.endpoints, endpoint)

	return endpoint
}

// Get registers a GET method handler.
func (e *Endpoint) Get(f http.HandlerFunc) *Endpoint {
	return e.register("GET", f)
}

// Head registers a HEAD method handler.
func (e *Endpoint) Head(f http.HandlerFunc) *Endpoint {
	return e.register("HEAD", f)
}

// Methods returns the list of methods available from the HTTP router endpoint.
func (e *Endpoint) Methods() []string {
	var hasGet, hasHead bool

	// Return all methods if a handler has been registered for any method.
	if _, ok := e.handlers[""]; ok {
		return []string{"DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"}
	}

	methods := []string{}
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
	if _, ok := e.handlers["OPTIONS"]; !ok {
		methods = append(methods, "OPTIONS")
	}
	sort.Strings(methods)

	return methods
}

// Options registers a OPTIONS method handler.
func (e *Endpoint) Options(f http.HandlerFunc) *Endpoint {
	return e.register("OPTIONS", f)
}

// Patch registers a PATCH method handler.
func (e *Endpoint) Patch(f http.HandlerFunc) *Endpoint {
	return e.register("PATCH", f)
}

// Post registers a POST method handler.
func (e *Endpoint) Post(f http.HandlerFunc) *Endpoint {
	return e.register("POST", f)
}

// Put registers a PUT method handler.
func (e *Endpoint) Put(f http.HandlerFunc) *Endpoint {
	return e.register("PUT", f)
}

// Use registers a new middleware in the HTTP handlers chain.
func (e *Endpoint) Use(f func(http.Handler) http.Handler) *Endpoint {
	e.middlewares = append(e.middlewares, f)
	e.updateChain()
	return e
}

func (e *Endpoint) register(method string, f http.HandlerFunc) *Endpoint {
	e.handlers[method] = f
	return e
}

func (e *Endpoint) serve(rw http.ResponseWriter, r *http.Request) {
	// Handle slash redirections
	if !e.pattern.wildcard {
		if e.pattern.slash && !strings.HasSuffix(r.URL.Path, "/") {
			r.URL.Path += "/"
			http.Redirect(rw, r, r.URL.String(), http.StatusPermanentRedirect)
			return
		} else if !e.pattern.slash && strings.HasSuffix(r.URL.Path, "/") {
			r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
			http.Redirect(rw, r, r.URL.String(), http.StatusPermanentRedirect)
			return
		}
	}

	handler, ok := e.handlers[r.Method]
	if !ok {
		if r.Method == "OPTIONS" {
			rw.Header().Add("Allow", strings.Join(e.Methods(), ", "))
			rw.WriteHeader(http.StatusNoContent)
			return
		} else if _, ok = e.handlers[""]; ok {
			// Use "Any" handler
			handler = e.handlers[""]
		} else if r.Method == "HEAD" {
			// Use GET method handler when HEAD is requested
			handler, ok = e.handlers["GET"]
		}

		if !ok {
			rw.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	} else if r.Method == "OPTIONS" {
		rw.Header().Add("Allow", strings.Join(e.Methods(), ", "))
	}

	handler(rw, r)
}

func (e *Endpoint) updateChain() {
	e.chain = e.handler
	for i := len(e.middlewares) - 1; i >= 0; i-- {
		e.chain = e.middlewares[i](e.chain)
	}
}

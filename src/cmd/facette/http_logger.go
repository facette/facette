package main

import (
	"net/http"

	"github.com/facette/logger"
)

type responseWriter struct {
	http.ResponseWriter
	request *http.Request
	log     *logger.Logger
}

func (rw responseWriter) WriteHeader(status int) {
	rw.ResponseWriter.WriteHeader(status)

	rw.log.Debug("received \"%s %s %s\" from %s, returned: %d", rw.request.Method, rw.request.URL,
		rw.request.Proto, rw.request.RemoteAddr, status)
}

func (w *httpWorker) httpHandleLogger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if w.service.config.Username != "" && w.service.config.Password != "" {
			user, pass, _ := r.BasicAuth()
			if user != w.service.config.Username || pass != w.service.config.Password {
				rw.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
				http.Error(rw, "Unauthorized.", 401)
				return
			}
		}
		h.ServeHTTP(responseWriter{rw, r, w.log}, r)
	})
}

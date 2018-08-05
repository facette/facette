// +build !builtin_assets

package web

import (
	"net/http"

	"facette.io/logger"
)

type responseWriter struct {
	http.ResponseWriter
	r      *http.Request
	logger *logger.Logger
}

func (rw responseWriter) WriteHeader(status int) {
	rw.ResponseWriter.WriteHeader(status)

	rw.logger.Debug(
		"received \"%s %s %s\" from %s, returned: %d",
		rw.r.Method,
		rw.r.URL,
		rw.r.Proto,
		rw.r.RemoteAddr,
		status,
	)
}

func (h *Handler) handleLog(hh http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		hh.ServeHTTP(responseWriter{rw, r, h.logger}, r)
	})
}

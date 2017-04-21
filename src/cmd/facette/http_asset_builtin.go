// +build builtin_assets

package main

import (
	"context"
	"net/http"
	"path"
	"strings"
)

const (
	httpDefaultPath = "html/index.html"
)

func (w *httpWorker) httpHandleAsset(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	var ct string

	// Stop handling assets if frontend is disabled
	if !w.service.config.Frontend.Enabled {
		rw.WriteHeader(http.StatusForbidden)
		return
	}

	// Get file data from built-in assets
	filePath := strings.TrimPrefix(r.URL.Path, "/assets/")
	if strings.HasSuffix(filePath, "/") || filepath.Ext(filePath) == "" {
		filePath = httpDefaultPath
	}

	data, err := Asset(filePath)
	if err != nil {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	// Get asset content type
	switch path.Ext(filePath) {
	case ".css":
		ct = "text/css"

	case ".js":
		ct = "text/javascript"

	default:
		ct = http.DetectContentType(data)
	}

	rw.Header().Set("Content-Type", ct)
	rw.WriteHeader(http.StatusOK)
	rw.Write(data)
}

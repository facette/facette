// +build builtin_assets

package main

import (
	"context"
	"net/http"
	"path"
	"strings"
)

func (w *httpWorker) httpHandleAsset(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	var (
		isDefault bool
		ct        string
	)

	// Stop handling assets if frontend is disabled
	if !w.service.config.Frontend.Enabled {
		rw.WriteHeader(http.StatusForbidden)
		return
	}

	// Get file data from built-in assets
	filePath := strings.TrimPrefix(r.URL.Path, w.service.config.RootPath+"/assets/")
	if strings.HasSuffix(filePath, "/") || filepath.Ext(filePath) == "" {
		filePath = httpDefaultPath
	}

	if filePath == httpDefaultPath {
		isDefault = true
	}

	data, err := Asset(filePath)
	if err != nil {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	// Handle default file path
	if isDefault {
		w.httpServeDefault(rw, string(data))
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

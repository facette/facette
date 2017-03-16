// +build !builtin_assets

package main

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	httpDefaultPath = "html/index.html"
)

func (w *httpWorker) httpHandleAsset(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	// Stop handling assets if frontend is disabled
	if !w.service.config.Frontend.Enabled {
		rw.WriteHeader(http.StatusForbidden)
		return
	}

	// Strip assets prefix and handle default path
	filePath := strings.TrimPrefix(r.URL.Path, "/assets")
	if strings.HasSuffix(filePath, "/") || filepath.Ext(filePath) == "" {
		filePath = httpDefaultPath
	}

	filePath = filepath.Join(w.service.config.Frontend.AssetsDir, filePath)

	// Check for existing asset file
	if fi, err := os.Stat(filePath); err != nil {
		http.Error(rw, "", http.StatusNotFound)
		return
	} else if fi.IsDir() && filePath != httpDefaultPath {
		// Prevent directory listing
		http.Error(rw, "", http.StatusForbidden)
		return
	}

	// Serve static file
	http.ServeFile(rw, r, filePath)
}

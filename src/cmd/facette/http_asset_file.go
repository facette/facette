// +build !builtin_assets

package main

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func (w *httpWorker) httpHandleAsset(rw http.ResponseWriter, r *http.Request) {
	var (
		isAsset   bool
		isDefault bool
		filePath  string
	)

	// Stop handling assets if frontend is disabled
	if !w.service.config.Frontend.Enabled {
		rw.WriteHeader(http.StatusForbidden)
		return
	}

	// Strip assets prefix and handle default path
	if strings.HasPrefix(r.URL.Path, w.service.config.RootPath+"/assets") {
		filePath = strings.TrimPrefix(r.URL.Path, w.service.config.RootPath+"/assets")
		isAsset = true
	}

	if strings.HasSuffix(filePath, "/") || !isAsset {
		filePath = httpDefaultPath
	}

	if filePath == httpDefaultPath {
		isDefault = true
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

	// Handle default file path
	if isDefault {
		file, err := os.OpenFile(filePath, os.O_RDONLY, 0644)
		if err != nil {
			w.service.log.Error("failed to open %q file: %s", filePath, err)
			http.Error(rw, "", http.StatusInternalServerError)
			return
		}

		data, err := ioutil.ReadAll(file)
		if err != nil {
			w.service.log.Error("failed to read %q file: %s", filePath, err)
			http.Error(rw, "", http.StatusInternalServerError)
			return
		}

		w.httpServeDefault(rw, string(data))
		return
	}

	// Serve static file
	http.ServeFile(rw, r, filePath)
}

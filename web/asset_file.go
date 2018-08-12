// +build !builtin_assets

package web

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var assetsDir = filepath.Join(filepath.Dir(os.Args[0]), "../dist/assets")

func (h *Handler) handleAsset(rw http.ResponseWriter, r *http.Request) {
	var (
		isAsset   bool
		isDefault bool
		filePath  string
	)

	// Stop handling assets if frontend is disabled
	if !h.config.Frontend.Enabled {
		rw.WriteHeader(http.StatusForbidden)
		return
	}

	// Strip assets prefix and handle default path
	if strings.HasPrefix(r.URL.Path, h.config.RootPath+"/assets") {
		filePath = strings.TrimPrefix(r.URL.Path, h.config.RootPath+"/assets")
		isAsset = true
	}

	if strings.HasSuffix(filePath, "/") || !isAsset {
		filePath = httpDefaultPath
	}

	if filePath == httpDefaultPath {
		isDefault = true
	}

	filePath = filepath.Join(assetsDir, filePath)

	// Check for existing asset file
	if fi, err := os.Stat(filePath); err != nil {
		rw.WriteHeader(http.StatusNotFound)
		return
	} else if fi.IsDir() && filePath != httpDefaultPath {
		// Prevent directory listing
		rw.WriteHeader(http.StatusForbidden)
		return
	}

	// Handle default file path
	if isDefault {
		file, err := os.OpenFile(filePath, os.O_RDONLY, 0600)
		if err != nil {
			h.logger.Error("failed to open %q file: %s", filePath, err)
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		data, err := ioutil.ReadAll(file)
		if err != nil {
			h.logger.Error("failed to read %q file: %s", filePath, err)
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		h.serveDefault(rw, string(data))
		return
	}

	// Serve static file
	http.ServeFile(rw, r, filePath)
}

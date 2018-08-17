// +build builtin_assets
//
//go:generate go-bindata -pkg web -prefix ../dist/assets -tags builtin_assets -o bindata.go ../dist/assets/...

package web

import (
	"net/http"
	"path"
	"strings"
)

func (h *Handler) handleAsset(rw http.ResponseWriter, r *http.Request) {
	var (
		isAsset   bool
		isDefault bool
		filePath  string
		ct        string
	)

	// Stop handling assets if frontend is disabled
	if !h.config.HTTP.EnableUI {
		rw.WriteHeader(http.StatusForbidden)
		return
	}

	// Get file data from built-in assets
	if strings.HasPrefix(r.URL.Path, h.config.HTTP.BasePath+"/assets/") {
		filePath = strings.TrimPrefix(r.URL.Path, h.config.HTTP.BasePath+"/assets/")
		isAsset = true
	}

	if strings.HasSuffix(filePath, "/") || !isAsset {
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
		h.serveDefault(rw, string(data))
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

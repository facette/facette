package server

import (
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/facette/facette/pkg/library"
	"github.com/facette/facette/pkg/logger"
)

func (server *Server) serveAdmin(writer http.ResponseWriter, request *http.Request) {
	var err error

	if request.Method != "GET" && request.Method != "HEAD" {
		server.serveResponse(writer, nil, http.StatusMethodNotAllowed)
		return
	}

	setHTTPCacheHeaders(writer)

	if strings.HasPrefix(request.URL.Path, urlAdminPath+"sourcegroups/") ||
		strings.HasPrefix(request.URL.Path, urlAdminPath+"metricgroups/") {
		err = server.serveAdminGroup(writer, request)
	} else if strings.HasPrefix(request.URL.Path, urlAdminPath+"graphs/") {
		err = server.serveAdminGraph(writer, request)
	} else if strings.HasPrefix(request.URL.Path, urlAdminPath+"collections/") {
		err = server.serveAdminCollection(writer, request)
	} else if request.URL.Path == urlAdminPath+"origins/" || request.URL.Path == urlAdminPath+"sources/" ||
		request.URL.Path == urlAdminPath+"metrics/" {
		err = server.serveAdminCatalog(writer, request)
	} else if strings.HasPrefix(request.URL.Path, urlAdminPath+"scales/") {
		err = server.serveAdminScale(writer, request)
	} else if request.URL.Path == urlAdminPath {
		err = server.serveAdminIndex(writer, request)
	} else {
		err = os.ErrNotExist
	}

	if os.IsNotExist(err) {
		server.serveError(writer, http.StatusNotFound)
	} else if err != nil {
		logger.Log(logger.LevelError, "server", "%s", err)
		server.serveError(writer, http.StatusInternalServerError)
	}
}

func (server *Server) serveAdminCatalog(writer http.ResponseWriter, request *http.Request) error {
	return server.execTemplate(
		writer,
		struct {
			URLPrefix string
			Section   string
		}{
			URLPrefix: server.Config.URLPrefix,
			Section:   strings.TrimRight(strings.TrimPrefix(request.URL.Path, urlAdminPath), "/"),
		},
		path.Join(server.Config.BaseDir, "html", "layout.html"),
		path.Join(server.Config.BaseDir, "html", "common", "element.html"),
		path.Join(server.Config.BaseDir, "html", "admin", "layout.html"),
		path.Join(server.Config.BaseDir, "html", "admin", "catalog_list.html"),
	)
}

func (server *Server) serveAdminCollection(writer http.ResponseWriter, request *http.Request) error {
	var tmplFile string

	data := struct {
		URLPrefix string
		Section   string
		Path      string
	}{
		URLPrefix: server.Config.URLPrefix,
	}

	data.Section, data.Path = splitAdminURLPath(request.URL.Path)

	if data.Path != "" && (data.Path == "add" || server.Library.ItemExists(data.Path, library.LibraryItemCollection)) {
		tmplFile = "collection_edit.html"
	} else if data.Path == "" {
		tmplFile = "collection_list.html"
	}

	if tmplFile == "" {
		return os.ErrNotExist
	}

	return server.execTemplate(
		writer,
		data,
		path.Join(server.Config.BaseDir, "html", "layout.html"),
		path.Join(server.Config.BaseDir, "html", "common", "element.html"),
		path.Join(server.Config.BaseDir, "html", "admin", "layout.html"),
		path.Join(server.Config.BaseDir, "html", "admin", tmplFile),
	)
}

func (server *Server) serveAdminGraph(writer http.ResponseWriter, request *http.Request) error {
	var tmplFile string

	data := struct {
		URLPrefix        string
		Section          string
		Path             string
		GraphTypeArea    int
		GraphTypeLine    int
		StackModeNone    int
		StackModeNormal  int
		StackModePercent int
	}{
		URLPrefix: server.Config.URLPrefix,
	}

	data.Section, data.Path = splitAdminURLPath(request.URL.Path)

	if data.Path != "" && (data.Path == "add" || server.Library.ItemExists(data.Path, library.LibraryItemGraph)) {
		tmplFile = "graph_edit.html"

		data.GraphTypeArea = library.GraphTypeArea
		data.GraphTypeLine = library.GraphTypeLine

		data.StackModeNone = library.StackModeNone
		data.StackModeNormal = library.StackModeNormal
		data.StackModePercent = library.StackModePercent
	} else if data.Path == "" {
		tmplFile = "graph_list.html"
	}

	if tmplFile == "" {
		return os.ErrNotExist
	}

	return server.execTemplate(
		writer,
		data,
		path.Join(server.Config.BaseDir, "html", "layout.html"),
		path.Join(server.Config.BaseDir, "html", "common", "element.html"),
		path.Join(server.Config.BaseDir, "html", "admin", "layout.html"),
		path.Join(server.Config.BaseDir, "html", "admin", tmplFile),
	)
}

func (server *Server) serveAdminGroup(writer http.ResponseWriter, request *http.Request) error {
	var (
		groupType int
		tmplFile  string
	)

	data := struct {
		URLPrefix string
		Section   string
		Path      string
		Origins   []string
	}{
		URLPrefix: server.Config.URLPrefix,
	}

	data.Section, data.Path = splitAdminURLPath(request.URL.Path)

	if data.Section == "sourcegroups" {
		groupType = library.LibraryItemSourceGroup
	} else if data.Section == "metricgroups" {
		groupType = library.LibraryItemMetricGroup
	}

	if data.Path != "" && (data.Path == "add" || server.Library.ItemExists(data.Path, groupType)) {
		tmplFile = "group_edit.html"

		for originName := range server.Catalog.Origins {
			data.Origins = append(data.Origins, originName)
		}
	} else if data.Path == "" {
		tmplFile = "group_list.html"
	}

	if tmplFile == "" {
		return os.ErrNotExist
	}

	return server.execTemplate(
		writer,
		data,
		path.Join(server.Config.BaseDir, "html", "layout.html"),
		path.Join(server.Config.BaseDir, "html", "common", "element.html"),
		path.Join(server.Config.BaseDir, "html", "admin", "layout.html"),
		path.Join(server.Config.BaseDir, "html", "admin", tmplFile),
	)
}

func (server *Server) serveAdminScale(writer http.ResponseWriter, request *http.Request) error {
	var tmplFile string

	data := struct {
		URLPrefix string
		Section   string
		Path      string
	}{
		URLPrefix: server.Config.URLPrefix,
	}

	data.Section, data.Path = splitAdminURLPath(request.URL.Path)

	if data.Path != "" && (data.Path == "add" || server.Library.ItemExists(data.Path, library.LibraryItemScale)) {
		tmplFile = "scale_edit.html"
	} else if data.Path == "" {
		tmplFile = "scale_list.html"
	}

	if tmplFile == "" {
		return os.ErrNotExist
	}

	return server.execTemplate(
		writer,
		data,
		path.Join(server.Config.BaseDir, "html", "layout.html"),
		path.Join(server.Config.BaseDir, "html", "common", "element.html"),
		path.Join(server.Config.BaseDir, "html", "admin", "layout.html"),
		path.Join(server.Config.BaseDir, "html", "admin", tmplFile),
	)
}

func (server *Server) serveAdminIndex(writer http.ResponseWriter, request *http.Request) error {
	return server.execTemplate(
		writer,
		struct {
			URLPrefix string
			Section   string
			Stats     *statsResponse
		}{
			URLPrefix: server.Config.URLPrefix,
			Section:   "",
			Stats:     server.getStats(writer, request),
		},
		path.Join(server.Config.BaseDir, "html", "layout.html"),
		path.Join(server.Config.BaseDir, "html", "common", "element.html"),
		path.Join(server.Config.BaseDir, "html", "admin", "layout.html"),
		path.Join(server.Config.BaseDir, "html", "admin", "index.html"),
	)
}

func splitAdminURLPath(path string) (string, string) {
	chunks := strings.SplitN(strings.TrimPrefix(path, urlAdminPath), "/", 2)
	return chunks[0], chunks[1]
}

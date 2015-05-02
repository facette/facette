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
	} else if strings.HasPrefix(request.URL.Path, urlAdminPath+"units/") {
		err = server.serveAdminUnit(writer, request)
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
		http.StatusOK,
		struct {
			URLPrefix string
			ReadOnly  bool
			Section   string
		}{
			URLPrefix: server.Config.URLPrefix,
			ReadOnly:  server.Config.ReadOnly,
			Section:   strings.TrimRight(strings.TrimPrefix(request.URL.Path, urlAdminPath), "/"),
		},
		path.Join(server.Config.BaseDir, "template", "layout.html"),
		path.Join(server.Config.BaseDir, "template", "common", "element.html"),
		path.Join(server.Config.BaseDir, "template", "admin", "layout.html"),
		path.Join(server.Config.BaseDir, "template", "admin", "catalog_list.html"),
	)
}

func (server *Server) serveAdminCollection(writer http.ResponseWriter, request *http.Request) error {
	var tmplFile string

	data := struct {
		URLPrefix string
		ReadOnly  bool
		Section   string
		Path      string
	}{
		URLPrefix: server.Config.URLPrefix,
		ReadOnly:  server.Config.ReadOnly,
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
		http.StatusOK,
		data,
		path.Join(server.Config.BaseDir, "template", "layout.html"),
		path.Join(server.Config.BaseDir, "template", "common", "element.html"),
		path.Join(server.Config.BaseDir, "template", "admin", "layout.html"),
		path.Join(server.Config.BaseDir, "template", "admin", tmplFile),
	)
}

func (server *Server) serveAdminGraph(writer http.ResponseWriter, request *http.Request) error {
	var tmplFile string

	data := struct {
		URLPrefix           string
		ReadOnly            bool
		Template            bool
		Section             string
		Path                string
		GraphTypeArea       int
		GraphTypeLine       int
		StackModeNone       int
		StackModeNormal     int
		StackModePercent    int
		GraphUnitTypeFixed  int
		GraphUnitTypeMetric int
	}{
		URLPrefix: server.Config.URLPrefix,
		ReadOnly:  server.Config.ReadOnly,
	}

	if request.FormValue("templates") != "" {
		data.Template = true
	}

	data.Section, data.Path = splitAdminURLPath(request.URL.Path)

	if data.Path != "" && (data.Path == "add" || server.Library.ItemExists(data.Path, library.LibraryItemGraph)) {
		if request.FormValue("linked") != "" {
			tmplFile = "graph_edit_linked.html"
		} else {
			tmplFile = "graph_edit.html"
		}

		data.GraphTypeArea = library.GraphTypeArea
		data.GraphTypeLine = library.GraphTypeLine

		data.StackModeNone = library.StackModeNone
		data.StackModeNormal = library.StackModeNormal
		data.StackModePercent = library.StackModePercent

		data.GraphUnitTypeFixed = library.GraphUnitTypeFixed
		data.GraphUnitTypeMetric = library.GraphUnitTypeMetric
	} else if data.Path == "" {
		tmplFile = "graph_list.html"
	}

	if tmplFile == "" {
		return os.ErrNotExist
	}

	return server.execTemplate(
		writer,
		http.StatusOK,
		data,
		path.Join(server.Config.BaseDir, "template", "layout.html"),
		path.Join(server.Config.BaseDir, "template", "common", "element.html"),
		path.Join(server.Config.BaseDir, "template", "admin", "layout.html"),
		path.Join(server.Config.BaseDir, "template", "admin", tmplFile),
	)
}

func (server *Server) serveAdminGroup(writer http.ResponseWriter, request *http.Request) error {
	var (
		groupType int
		tmplFile  string
	)

	data := struct {
		URLPrefix string
		ReadOnly  bool
		Section   string
		Path      string
		Origins   []string
	}{
		URLPrefix: server.Config.URLPrefix,
		ReadOnly:  server.Config.ReadOnly,
	}

	data.Section, data.Path = splitAdminURLPath(request.URL.Path)

	if data.Section == "sourcegroups" {
		groupType = library.LibraryItemSourceGroup
	} else if data.Section == "metricgroups" {
		groupType = library.LibraryItemMetricGroup
	}

	if data.Path != "" && (data.Path == "add" || server.Library.ItemExists(data.Path, groupType)) {
		tmplFile = "group_edit.html"

		for _, origin := range server.Catalog.GetOrigins() {
			data.Origins = append(data.Origins, origin.Name)
		}
	} else if data.Path == "" {
		tmplFile = "group_list.html"
	}

	if tmplFile == "" {
		return os.ErrNotExist
	}

	return server.execTemplate(
		writer,
		http.StatusOK,
		data,
		path.Join(server.Config.BaseDir, "template", "layout.html"),
		path.Join(server.Config.BaseDir, "template", "common", "element.html"),
		path.Join(server.Config.BaseDir, "template", "admin", "layout.html"),
		path.Join(server.Config.BaseDir, "template", "admin", tmplFile),
	)
}

func (server *Server) serveAdminUnit(writer http.ResponseWriter, request *http.Request) error {
	var tmplFile string

	data := struct {
		URLPrefix        string
		ReadOnly         bool
		Section          string
		Path             string
		UnitTypeAbsolute int
		UnitTypeDuration int
	}{
		URLPrefix: server.Config.URLPrefix,
		ReadOnly:  server.Config.ReadOnly,
	}

	data.Section, data.Path = splitAdminURLPath(request.URL.Path)

	if data.Path != "" && (data.Path == "add" || server.Library.ItemExists(data.Path, library.LibraryItemUnit)) {
		tmplFile = "unit_edit.html"

		data.UnitTypeAbsolute = library.UnitTypeAbsolute
		data.UnitTypeDuration = library.UnitTypeDuration
	} else if data.Path == "" {
		tmplFile = "unit_list.html"
	}

	if tmplFile == "" {
		return os.ErrNotExist
	}

	return server.execTemplate(
		writer,
		http.StatusOK,
		data,
		path.Join(server.Config.BaseDir, "template", "layout.html"),
		path.Join(server.Config.BaseDir, "template", "common", "element.html"),
		path.Join(server.Config.BaseDir, "template", "admin", "layout.html"),
		path.Join(server.Config.BaseDir, "template", "admin", tmplFile),
	)
}

func (server *Server) serveAdminScale(writer http.ResponseWriter, request *http.Request) error {
	var tmplFile string

	data := struct {
		URLPrefix string
		ReadOnly  bool
		Section   string
		Path      string
	}{
		URLPrefix: server.Config.URLPrefix,
		ReadOnly:  server.Config.ReadOnly,
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
		http.StatusOK,
		data,
		path.Join(server.Config.BaseDir, "template", "layout.html"),
		path.Join(server.Config.BaseDir, "template", "common", "element.html"),
		path.Join(server.Config.BaseDir, "template", "admin", "layout.html"),
		path.Join(server.Config.BaseDir, "template", "admin", tmplFile),
	)
}

func (server *Server) serveAdminIndex(writer http.ResponseWriter, request *http.Request) error {
	return server.execTemplate(
		writer,
		http.StatusOK,
		struct {
			URLPrefix        string
			ReadOnly         bool
			HideBuildDetails bool
			Section          string
			Build            *buildInfo
			Stats            *statsResponse
		}{
			URLPrefix:        server.Config.URLPrefix,
			ReadOnly:         server.Config.ReadOnly,
			HideBuildDetails: server.Config.HideBuildDetails,
			Section:          "",
			Build:            server.buildInfo,
			Stats:            server.getStats(writer, request),
		},
		path.Join(server.Config.BaseDir, "template", "layout.html"),
		path.Join(server.Config.BaseDir, "template", "common", "element.html"),
		path.Join(server.Config.BaseDir, "template", "admin", "layout.html"),
		path.Join(server.Config.BaseDir, "template", "admin", "index.html"),
	)
}

func splitAdminURLPath(path string) (string, string) {
	chunks := strings.SplitN(strings.TrimPrefix(path, urlAdminPath), "/", 2)
	return chunks[0], chunks[1]
}

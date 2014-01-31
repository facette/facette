package server

import (
	"facette/library"
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
)

func (server *Server) adminHandle(writer http.ResponseWriter, request *http.Request) {
	var (
		data struct {
			URLPrefix        string
			Section          string
			Path             string
			Origins          []string
			GraphTypeArea    int
			GraphTypeLine    int
			StackModeNone    int
			StackModeNormal  int
			StackModePercent int
		}
		err        error
		groupType  int
		tmpl       *template.Template
		tmplFile   string
		tmplFolder string
	)

	if !server.handleAuth(writer, request) {
		server.handleError(writer, http.StatusUnauthorized)
		return
	}

	// Set template data
	data.URLPrefix = server.Config.URLPrefix
	data.Section = mux.Vars(request)["section"]
	data.Path = mux.Vars(request)["path"]

	tmplFile = "unknown"

	if data.Section == "graphs" || data.Section == "collections" {
		tmplFolder = data.Section

		data.GraphTypeArea = library.GraphTypeArea
		data.GraphTypeLine = library.GraphTypeLine

		data.StackModeNone = library.StackModeNone
		data.StackModeNormal = library.StackModeNormal
		data.StackModePercent = library.StackModePercent

		if data.Path != "" && (data.Path == "add" ||
			data.Section == "graphs" && server.Library.ItemExists(data.Path, library.LibraryItemGraph) ||
			data.Section == "collections" && server.Library.ItemExists(data.Path, library.LibraryItemCollection)) {
			tmplFile = "edit.html"
		} else if data.Path == "" {
			tmplFile = "list.html"
		}
	} else if data.Section == "sourcegroups" || data.Section == "metricgroups" {
		tmplFolder = "groups"

		for originName := range server.Catalog.Origins {
			data.Origins = append(data.Origins, originName)
		}

		if data.Section == "sourcegroups" {
			groupType = library.LibraryItemSourceGroup
		} else if data.Section == "metricgroups" {
			groupType = library.LibraryItemMetricGroup
		}

		if data.Path != "" && (data.Path == "add" || server.Library.ItemExists(data.Path, groupType)) {
			tmplFile = "edit.html"
		} else if data.Path == "" {
			tmplFile = "list.html"
		}
	} else if data.Section == "origins" || data.Section == "sources" || data.Section == "metrics" {
		tmplFolder = "catalog"
		tmplFile = "list.html"
	} else if data.Section == "" {
		tmplFolder = ""
		tmplFile = "index.html"
	}

	// Return template data
	if tmpl, err = template.New("layout.html").Funcs(template.FuncMap{
		"asset":  server.templateAsset,
		"eq":     templateEqual,
		"ne":     templateNotEqual,
		"substr": templateSubstr,
	}).ParseFiles(
		path.Join(server.Config.BaseDir, "html", "layout.html"),
		path.Join(server.Config.BaseDir, "html", "common", "element.html"),
		path.Join(server.Config.BaseDir, "html", "admin", "layout.html"),
		path.Join(server.Config.BaseDir, "html", "admin", tmplFolder, tmplFile),
	); err == nil {
		err = tmpl.Execute(writer, data)
	}

	if err != nil {
		log.Println("ERROR: " + err.Error())

		if os.IsNotExist(err) {
			server.handleError(writer, http.StatusNotFound)
		} else {
			server.handleError(writer, http.StatusInternalServerError)
		}
	}
}

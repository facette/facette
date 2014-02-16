package server

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/facette/facette/pkg/library"
	"github.com/facette/facette/thirdparty/github.com/gorilla/mux"
)

func (server *Server) adminHandle(writer http.ResponseWriter, request *http.Request) {
	var data struct {
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

	if !server.handleAuth(writer, request) {
		server.handleError(writer, http.StatusUnauthorized)
		return
	}

	// Set template data
	data.URLPrefix = server.Config.URLPrefix
	data.Section = mux.Vars(request)["section"]
	data.Path = mux.Vars(request)["path"]

	tmplFile := "unknown"
	tmplPrefix := ""

	if data.Section == "graphs" || data.Section == "collections" {
		tmplPrefix = data.Section[:len(data.Section)-1] + "_"

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
		var groupType int

		tmplPrefix = "group_"

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
		tmplPrefix = "catalog_"
		tmplFile = "list.html"
	} else if data.Section == "" {
		tmplPrefix = ""
		tmplFile = "index.html"
	}

	// Return template data
	tmpl, err := template.New("layout.html").Funcs(template.FuncMap{
		"asset":  server.templateAsset,
		"eq":     templateEqual,
		"ne":     templateNotEqual,
		"substr": templateSubstr,
	}).ParseFiles(
		path.Join(server.Config.BaseDir, "html", "layout.html"),
		path.Join(server.Config.BaseDir, "html", "common", "element.html"),
		path.Join(server.Config.BaseDir, "html", "admin", "layout.html"),
		path.Join(server.Config.BaseDir, "html", "admin", tmplPrefix+tmplFile),
	)
	if err == nil {
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

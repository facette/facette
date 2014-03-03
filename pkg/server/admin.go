package server

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/facette/facette/pkg/library"
)

func (server *Server) handleAdmin(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "GET" && request.Method != "HEAD" {
		server.handleResponse(writer, nil, http.StatusMethodNotAllowed)
		return
	} else if !server.handleAuth(writer, request) {
		server.handleError(writer, http.StatusUnauthorized)
		return
	}

	tmpl, err := template.New("layout.html").Funcs(template.FuncMap{
		"asset":  server.templateAsset,
		"eq":     templateEqual,
		"ne":     templateNotEqual,
		"substr": templateSubstr,
	}).ParseFiles(
		path.Join(server.Config.BaseDir, "html", "layout.html"),
		path.Join(server.Config.BaseDir, "html", "common", "element.html"),
		path.Join(server.Config.BaseDir, "html", "admin", "layout.html"),
	)

	if strings.HasPrefix(request.URL.Path, urlAdminPath+"sourcegroups/") ||
		strings.HasPrefix(request.URL.Path, urlAdminPath+"metricgroups/") {
		err = server.handleAdminGroup(writer, request, tmpl)
	} else if strings.HasPrefix(request.URL.Path, urlAdminPath+"graphs/") {
		err = server.handleAdminGraph(writer, request, tmpl)
	} else if strings.HasPrefix(request.URL.Path, urlAdminPath+"collections/") {
		err = server.handleAdminCollection(writer, request, tmpl)
	} else if request.URL.Path == urlAdminPath+"origins/" || request.URL.Path == urlAdminPath+"sources/" ||
		request.URL.Path == urlAdminPath+"metrics/" {
		err = server.handleAdminCatalog(writer, request, tmpl)
	} else {
		err = server.handleAdminIndex(writer, request, tmpl)
	}

	if os.IsNotExist(err) {
		server.handleError(writer, http.StatusNotFound)
	} else if err != nil {
		log.Println("ERROR: " + err.Error())
		server.handleError(writer, http.StatusInternalServerError)
	}
}

func (server *Server) handleAdminCatalog(writer http.ResponseWriter, request *http.Request,
	tmpl *template.Template) error {

	var data struct {
		URLPrefix string
		Section   string
	}

	// Set template data
	data.URLPrefix = server.Config.URLPrefix
	data.Section = strings.TrimRight(strings.TrimPrefix(request.URL.Path, urlAdminPath), "/")

	// Execute template
	tmpl, err := tmpl.ParseFiles(path.Join(server.Config.BaseDir, "html", "admin", "catalog_list.html"))
	if err != nil {
		return err
	}

	return tmpl.Execute(writer, data)
}

func (server *Server) handleAdminCollection(writer http.ResponseWriter, request *http.Request,
	tmpl *template.Template) error {

	var (
		data struct {
			URLPrefix string
			Section   string
			Path      string
		}
		tmplFile string
	)

	// Set template data
	data.URLPrefix = server.Config.URLPrefix
	data.Section, data.Path = splitAdminURLPath(request.URL.Path)

	if data.Path != "" && (data.Path == "add" || server.Library.ItemExists(data.Path, library.LibraryItemCollection)) {
		tmplFile = "collection_edit.html"
	} else if data.Path == "" {
		tmplFile = "collection_list.html"
	}

	if tmplFile == "" {
		return os.ErrNotExist
	}

	// Execute template
	tmpl, err := tmpl.ParseFiles(path.Join(server.Config.BaseDir, "html", "admin", tmplFile))
	if err != nil {
		return err
	}

	return tmpl.Execute(writer, data)
}

func (server *Server) handleAdminGraph(writer http.ResponseWriter, request *http.Request,
	tmpl *template.Template) error {

	var (
		data struct {
			URLPrefix        string
			Section          string
			Path             string
			GraphTypeArea    int
			GraphTypeLine    int
			StackModeNone    int
			StackModeNormal  int
			StackModePercent int
		}
		tmplFile string
	)

	// Set template data
	data.URLPrefix = server.Config.URLPrefix
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

	// Execute template
	tmpl, err := tmpl.ParseFiles(path.Join(server.Config.BaseDir, "html", "admin", tmplFile))
	if err != nil {
		return err
	}

	return tmpl.Execute(writer, data)
}

func (server *Server) handleAdminGroup(writer http.ResponseWriter, request *http.Request,
	tmpl *template.Template) error {

	var (
		data struct {
			URLPrefix string
			Section   string
			Path      string
			Origins   []string
		}
		groupType int
		tmplFile  string
	)

	// Set template data
	data.URLPrefix = server.Config.URLPrefix
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

	// Execute template
	tmpl, err := tmpl.ParseFiles(path.Join(server.Config.BaseDir, "html", "admin", tmplFile))
	if err != nil {
		return err
	}

	return tmpl.Execute(writer, data)
}

func (server *Server) handleAdminIndex(writer http.ResponseWriter, request *http.Request,
	tmpl *template.Template) error {

	var data struct {
		URLPrefix string
		Section   string
		Stats     *statsResponse
	}

	// Set template data
	data.URLPrefix = server.Config.URLPrefix
	data.Section = ""
	data.Stats = server.getStats(writer, request)

	// Execute template
	tmpl, err := tmpl.ParseFiles(path.Join(server.Config.BaseDir, "html", "admin", "index.html"))
	if err != nil {
		return err
	}

	return tmpl.Execute(writer, data)
}

func splitAdminURLPath(path string) (string, string) {
	chunks := strings.SplitN(strings.TrimPrefix(path, urlAdminPath), "/", 2)
	return chunks[0], chunks[1]
}

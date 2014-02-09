package server

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/facette/facette/pkg/backend"
	"github.com/facette/facette/pkg/library"
	"github.com/facette/facette/thirdparty/github.com/gorilla/mux"
)

func (server *Server) browseHandleCollection(writer http.ResponseWriter, request *http.Request,
	tmpl *template.Template) error {
	type collectionData struct {
		*library.Collection
		Parent string
	}

	var (
		data struct {
			URLPrefix  string
			Collection *collectionData
			Request    *http.Request
		}
		err  error
		item interface{}
	)

	// Set template data
	data.URLPrefix = server.Config.URLPrefix
	data.Collection = &collectionData{Collection: &library.Collection{}}

	data.Collection.ID = mux.Vars(request)["collection"]

	if item, err = server.Library.GetItem(data.Collection.ID, library.LibraryItemCollection); err != nil {
		return err
	}

	data.Collection.Collection = item.(*library.Collection)

	if request.FormValue("q") != "" {
		data.Collection.Collection = server.Library.FilterCollection(data.Collection.Collection, request.FormValue("q"))
	}

	if data.Collection.Collection.Parent != nil {
		data.Collection.Parent = data.Collection.Collection.Parent.ID
	} else {
		data.Collection.Parent = "null"
	}

	// Execute template
	if tmpl, err = tmpl.ParseFiles(
		path.Join(server.Config.BaseDir, "html", "layout.html"),
		path.Join(server.Config.BaseDir, "html", "common", "element.html"),
		path.Join(server.Config.BaseDir, "html", "common", "graph.html"),
		path.Join(server.Config.BaseDir, "html", "browse", "layout.html"),
		path.Join(server.Config.BaseDir, "html", "browse", "collection.html"),
	); err != nil {
		return err
	}

	data.Request = request

	return tmpl.Execute(writer, data)
}

func (server *Server) browseHandleIndex(writer http.ResponseWriter, request *http.Request,
	tmpl *template.Template) error {
	var (
		data struct {
			URLPrefix string
		}
		err error
	)

	// Set template data
	data.URLPrefix = server.Config.URLPrefix

	// Execute template
	if tmpl, err = tmpl.ParseFiles(
		path.Join(server.Config.BaseDir, "html", "layout.html"),
		path.Join(server.Config.BaseDir, "html", "common", "element.html"),
		path.Join(server.Config.BaseDir, "html", "browse", "layout.html"),
		path.Join(server.Config.BaseDir, "html", "browse", "index.html"),
	); err != nil {
		return err
	}

	return tmpl.Execute(writer, data)
}

func (server *Server) browseHandleSearch(writer http.ResponseWriter, request *http.Request,
	tmpl *template.Template) error {
	var (
		chunks []string
		data   struct {
			URLPrefix   string
			Count       int
			Request     *http.Request
			Sources     []*backend.Source
			Collections []*library.Collection
		}
		err error
	)

	// Set template data
	data.URLPrefix = server.Config.URLPrefix
	data.Request = request

	// Perform search filtering
	if request.FormValue("q") != "" {
		for _, chunk := range strings.Split(strings.ToLower(request.FormValue("q")), " ") {
			chunks = append(chunks, strings.Trim(chunk, " \t"))
		}

		for _, origin := range server.Catalog.Origins {
			for _, source := range origin.Sources {
				for _, chunk := range chunks {
					if strings.Index(strings.ToLower(source.Name), chunk) == -1 {
						goto nextOrigin
					}
				}

				data.Sources = append(data.Sources, source)
			nextOrigin:
			}
		}

		for _, collection := range server.Library.Collections {
			for _, chunk := range chunks {
				if strings.Index(strings.ToLower(collection.Name), chunk) == -1 {
					goto nextCollection
				}
			}

			data.Collections = append(data.Collections, collection)
		nextCollection:
		}
	}

	data.Count = len(data.Sources) + len(data.Collections)

	// Execute template
	if tmpl, err = tmpl.ParseFiles(
		path.Join(server.Config.BaseDir, "html", "layout.html"),
		path.Join(server.Config.BaseDir, "html", "common", "element.html"),
		path.Join(server.Config.BaseDir, "html", "browse", "layout.html"),
		path.Join(server.Config.BaseDir, "html", "browse", "search.html"),
	); err != nil {
		return err
	}

	return tmpl.Execute(writer, data)
}

func (server *Server) browseHandleSource(writer http.ResponseWriter, request *http.Request,
	tmpl *template.Template) error {
	var (
		data struct {
			URLPrefix  string
			Collection *library.Collection
			Request    *http.Request
		}
		err        error
		sourceName string
	)

	sourceName = mux.Vars(request)["source"]

	// Set template data
	data.URLPrefix = server.Config.URLPrefix

	if data.Collection, err = server.Library.GetCollectionTemplate(sourceName); err != nil {
		return err
	}

	if request.FormValue("q") != "" {
		data.Collection = server.Library.FilterCollection(data.Collection, request.FormValue("q"))
	}

	// Execute template
	if tmpl, err = tmpl.ParseFiles(
		path.Join(server.Config.BaseDir, "html", "layout.html"),
		path.Join(server.Config.BaseDir, "html", "common", "element.html"),
		path.Join(server.Config.BaseDir, "html", "common", "graph.html"),
		path.Join(server.Config.BaseDir, "html", "browse", "layout.html"),
		path.Join(server.Config.BaseDir, "html", "browse", "collection.html"),
	); err != nil {
		return err
	}

	data.Request = request

	return tmpl.Execute(writer, data)
}

func (server *Server) browseHandle(writer http.ResponseWriter, request *http.Request) {
	var (
		err  error
		tmpl *template.Template
	)

	if request.Method != "GET" && request.Method != "HEAD" {
		server.handleResponse(writer, http.StatusMethodNotAllowed)
		return
	}

	// Redirect to browse
	if request.URL.Path == "/" {
		http.Redirect(writer, request, server.Config.URLPrefix+URLBrowsePath+"/", 301)
		return
	}

	// Return template data
	tmpl = template.New("layout.html").Funcs(template.FuncMap{
		"asset": server.templateAsset,
		"eq":    templateEqual,
		"ne":    templateNotEqual,
		"dump":  templateDumpMap,
		"hl":    templateHighlight,
	})

	// Execute template
	if mux.Vars(request)["source"] != "" {
		err = server.browseHandleSource(writer, request, tmpl)
	} else if mux.Vars(request)["collection"] != "" {
		err = server.browseHandleCollection(writer, request, tmpl)
	} else if strings.HasSuffix(request.URL.Path, "/search") {
		err = server.browseHandleSearch(writer, request, tmpl)
	} else {
		err = server.browseHandleIndex(writer, request, tmpl)
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

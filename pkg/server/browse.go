package server

import (
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/facette/facette/pkg/catalog"
	"github.com/facette/facette/pkg/library"
	"github.com/facette/facette/pkg/logger"
)

func (server *Server) serveBrowse(writer http.ResponseWriter, request *http.Request) {
	var err error

	if request.Method != "GET" && request.Method != "HEAD" {
		server.serveResponse(writer, nil, http.StatusMethodNotAllowed)
		return
	}

	// Redirect to default location
	if request.URL.Path == "/" {
		http.Redirect(writer, request, server.Config.URLPrefix+urlBrowsePath, 301)
		return
	}

	setHTTPCacheHeaders(writer)

	if strings.HasPrefix(request.URL.Path, urlBrowsePath+"collections/") {
		err = server.serveBrowseCollection(writer, request)
	} else if strings.HasPrefix(request.URL.Path, urlBrowsePath+"graphs/") {
		err = server.serveBrowseGraph(writer, request)
	} else if request.URL.Path == urlBrowsePath+"search" {
		err = server.serveBrowseSearch(writer, request)
	} else if request.URL.Path == urlBrowsePath {
		err = server.serveBrowseIndex(writer, request)
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

func (server *Server) serveBrowseIndex(writer http.ResponseWriter, request *http.Request) error {
	return server.execTemplate(
		writer,
		http.StatusOK,
		struct {
			URLPrefix string
		}{
			URLPrefix: server.Config.URLPrefix,
		},
		path.Join(server.Config.BaseDir, "template", "layout.html"),
		path.Join(server.Config.BaseDir, "template", "common", "element.html"),
		path.Join(server.Config.BaseDir, "template", "browse", "layout.html"),
		path.Join(server.Config.BaseDir, "template", "browse", "index.html"),
	)
}

func (server *Server) serveBrowseCollection(writer http.ResponseWriter, request *http.Request) error {
	type collectionData struct {
		*library.Collection
		Parent string
	}

	data := struct {
		URLPrefix  string
		Collection *collectionData
		Request    *http.Request
	}{
		URLPrefix:  server.Config.URLPrefix,
		Collection: &collectionData{Collection: &library.Collection{}},
		Request:    request,
	}

	data.Collection.ID = strings.TrimPrefix(request.URL.Path, urlBrowsePath+"collections/")

	item, err := server.Library.GetItem(data.Collection.ID, library.LibraryItemCollection)
	if err != nil {
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

	return server.execTemplate(
		writer,
		http.StatusOK,
		data,
		path.Join(server.Config.BaseDir, "template", "layout.html"),
		path.Join(server.Config.BaseDir, "template", "common", "element.html"),
		path.Join(server.Config.BaseDir, "template", "common", "graph.html"),
		path.Join(server.Config.BaseDir, "template", "browse", "layout.html"),
		path.Join(server.Config.BaseDir, "template", "browse", "collection.html"),
	)
}

func (server *Server) serveBrowseGraph(writer http.ResponseWriter, request *http.Request) error {
	data := struct {
		URLPrefix string
		Graph     *library.Graph
		Request   *http.Request
	}{
		URLPrefix: server.Config.URLPrefix,
		Request:   request,
	}

	item, err := server.Library.GetItem(
		strings.TrimPrefix(request.URL.Path, urlBrowsePath+"graphs/"),
		library.LibraryItemGraph,
	)
	if err != nil {
		return err
	}

	data.Graph = item.(*library.Graph)

	return server.execTemplate(
		writer,
		http.StatusOK,
		data,
		path.Join(server.Config.BaseDir, "template", "layout.html"),
		path.Join(server.Config.BaseDir, "template", "common", "element.html"),
		path.Join(server.Config.BaseDir, "template", "common", "graph.html"),
		path.Join(server.Config.BaseDir, "template", "browse", "layout.html"),
		path.Join(server.Config.BaseDir, "template", "browse", "graph.html"),
	)
}

func (server *Server) serveBrowseSearch(writer http.ResponseWriter, request *http.Request) error {
	data := struct {
		URLPrefix   string
		Count       int
		Request     *http.Request
		Sources     []*catalog.Source
		Collections []*library.Collection
	}{
		URLPrefix: server.Config.URLPrefix,
		Request:   request,
	}

	// Perform search filtering
	if request.FormValue("q") != "" {
		chunks := make([]string, 0)

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

	return server.execTemplate(
		writer,
		http.StatusOK,
		data,
		path.Join(server.Config.BaseDir, "template", "layout.html"),
		path.Join(server.Config.BaseDir, "template", "common", "element.html"),
		path.Join(server.Config.BaseDir, "template", "browse", "layout.html"),
		path.Join(server.Config.BaseDir, "template", "browse", "search.html"),
	)
}

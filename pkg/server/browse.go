package server

import (
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/facette/facette/pkg/library"
	"github.com/facette/facette/pkg/logger"
	"github.com/facette/facette/pkg/utils"
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
	} else if request.URL.Path == urlBrowsePath+"search" ||
		request.URL.Path == urlBrowsePath+"opensearch.xml" {
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
			Request   *http.Request
		}{
			URLPrefix: server.Config.URLPrefix,
			Request:   request,
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

	data.Collection.Collection = server.Library.FilterCollection(item.(*library.Collection), request.FormValue("q"))

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
		URLBase     string
		URLPrefix   string
		Count       int
		Request     *http.Request
		Collections []*library.Collection
		Graphs      []*library.Graph
	}{
		URLPrefix: server.Config.URLPrefix,
		Request:   request,
	}

	// Handle OpenSearch
	if request.URL.Path == urlBrowsePath+"opensearch.xml" {
		data.URLBase = utils.HTTPGetURLBase(request)

		writer.Header().Set("Content-Type", "text/xml; charset=utf-8")

		return server.execTemplate(
			writer,
			http.StatusOK,
			data,
			path.Join(server.Config.BaseDir, "template", "opensearch.xml"),
		)
	}

	// Perform search filtering
	if request.FormValue("q") != "" {
		chunks := make([]string, 0)

		for _, chunk := range strings.Split(strings.ToLower(request.FormValue("q")), " ") {
			chunks = append(chunks, "*"+strings.Trim(chunk, " \t")+"*")
		}

		for _, collection := range server.Library.Collections {
			for _, chunk := range chunks {
				if ok, _ := path.Match(chunk, strings.ToLower(collection.Name)); !ok {
					goto nextCollection
				}
			}

			data.Collections = append(data.Collections, collection)
		nextCollection:
		}

		for _, graph := range server.Library.Graphs {
			for _, chunk := range chunks {
				if ok, _ := path.Match(chunk, strings.ToLower(graph.Name)); !ok {
					goto nextGraph
				}
			}

			data.Graphs = append(data.Graphs, graph)
		nextGraph:
		}
	}

	data.Count = len(data.Collections) + len(data.Graphs)

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

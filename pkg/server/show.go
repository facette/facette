package server

import (
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/facette/facette/pkg/config"
	"github.com/facette/facette/pkg/library"
	"github.com/facette/facette/pkg/logger"
)

func (server *Server) serveShow(writer http.ResponseWriter, request *http.Request) {
	var err error

	if request.Method != "GET" && request.Method != "HEAD" {
		server.serveResponse(writer, nil, http.StatusMethodNotAllowed)
		return
	}

	setHTTPCacheHeaders(writer)

	if strings.HasPrefix(request.URL.Path, urlShowPath+"graphs/") {
		err = server.serveShowGraph(writer, request)
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

func (server *Server) serveShowGraph(writer http.ResponseWriter, request *http.Request) error {
	data := struct {
		URLPrefix string
		API       config.APIConfig
		Graph     *library.Graph
		Request   *http.Request
		Range     string
	}{
		URLPrefix: server.Config.URLPrefix,
		API:       server.Config.API,
		Range:     request.FormValue("range"),
		Request:   request,
	}

	item, err := server.Library.GetItem(
		strings.TrimPrefix(request.URL.Path, urlShowPath+"graphs/"),
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
		path.Join(server.Config.BaseDir, "template", "show", "layout.html"),
		path.Join(server.Config.BaseDir, "template", "show", "graph.html"),
	)
}

package server

import (
	"github.com/fatih/goset"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
	"time"
)

func (server *Server) reloadHandle(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "GET" && request.Method != "HEAD" {
		server.handleResponse(writer, http.StatusMethodNotAllowed)
		return
	} else if !server.handleAuth(writer, request) {
		return
	}

	server.Reload()

	server.handleJSON(writer, statusResponse{"OK"})
}

func (server *Server) statHandle(writer http.ResponseWriter, request *http.Request) {
	var (
		metrics *goset.Set
		sources *goset.Set
		result  *statResponse
	)

	if request.Method != "GET" && request.Method != "HEAD" {
		server.handleResponse(writer, http.StatusMethodNotAllowed)
		return
	}

	sources = goset.New()
	metrics = goset.New()

	for _, origin := range server.Catalog.Origins {
		for key, source := range origin.Sources {
			sources.Add(key)

			for key := range source.Metrics {
				metrics.Add(key)
			}
		}
	}

	result = &statResponse{
		Origins:        len(server.Catalog.Origins),
		Sources:        sources.Size(),
		Metrics:        metrics.Size(),
		CatalogUpdated: server.Catalog.Updated.Format(time.RFC3339),

		Graphs:      len(server.Library.Graphs),
		Collections: len(server.Library.Collections),
		Groups:      len(server.Library.Groups),
	}

	server.handleJSON(writer, result)
}

func (server *Server) waitHandle(writer http.ResponseWriter, request *http.Request) {
	var (
		err  error
		tmpl *template.Template
	)

	// Execute template
	if tmpl, err = template.New("layout.html").ParseFiles(
		path.Join(server.Config.BaseDir, "html", "layout.html"),
		path.Join(server.Config.BaseDir, "html", "wait.html"),
	); err == nil {
		err = tmpl.Execute(writer, nil)
	}

	if err != nil {
		if os.IsNotExist(err) {
			log.Println("ERROR: " + err.Error())
			server.handleResponse(writer, http.StatusNotFound)
		} else {
			log.Println("ERROR: " + err.Error())
			server.handleResponse(writer, http.StatusInternalServerError)
		}
	}
}

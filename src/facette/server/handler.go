package server

import (
	"github.com/fatih/set"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
	"time"
)

type resourceResponse struct {
	Scales [][2]interface{} `json:"scales"`
}

type statResponse struct {
	Origins        int    `json:"origins"`
	Sources        int    `json:"sources"`
	Metrics        int    `json:"metrics"`
	CatalogUpdated string `json:"catalog_updated"`

	Graphs      int `json:"graphs"`
	Collections int `json:"collections"`
	Groups      int `json:"groups"`
}

func (server *Server) reloadHandle(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "GET" && request.Method != "HEAD" {
		server.handleResponse(writer, http.StatusMethodNotAllowed)
		return
	} else if !server.handleAuth(writer, request) {
		server.handleResponse(writer, http.StatusUnauthorized)
		return
	}

	server.Reload()

	server.handleJSON(writer, statusResponse{"OK"})
}

func (server *Server) resourceHandle(writer http.ResponseWriter, request *http.Request) {
	var (
		result *resourceResponse
	)

	result = &resourceResponse{
		Scales: server.Config.Scales,
	}

	server.handleJSON(writer, result)
}

func (server *Server) statHandle(writer http.ResponseWriter, request *http.Request) {
	var (
		metrics *set.Set
		sources *set.Set
		result  *statResponse
	)

	if request.Method != "GET" && request.Method != "HEAD" {
		server.handleResponse(writer, http.StatusMethodNotAllowed)
		return
	}

	sources = set.New()
	metrics = set.New()

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
		data struct {
			URLPrefix string
		}
		err  error
		tmpl *template.Template
	)

	// Set template data
	data.URLPrefix = server.Config.URLPrefix

	// Execute template
	if tmpl, err = template.New("layout.html").ParseFiles(
		path.Join(server.Config.BaseDir, "html", "layout.html"),
		path.Join(server.Config.BaseDir, "html", "wait.html"),
	); err == nil {
		err = tmpl.Execute(writer, data)
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

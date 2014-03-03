package server

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"html/template"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/facette/facette/thirdparty/github.com/fatih/set"
)

type serverResponse struct {
	Message string `json:"message"`
}

type statsResponse struct {
	Origins        int    `json:"origins"`
	Sources        int    `json:"sources"`
	Metrics        int    `json:"metrics"`
	CatalogUpdated string `json:"catalog_updated"`

	Graphs      int `json:"graphs"`
	Collections int `json:"collections"`
	Groups      int `json:"groups"`
}

type resourceResponse struct {
	Scales [][2]interface{} `json:"scales"`
}

func (server *Server) handleAuth(writer http.ResponseWriter, request *http.Request) bool {
	authorization := request.Header.Get("Authorization")

	if strings.HasPrefix(authorization, "Basic ") {
		data, err := base64.StdEncoding.DecodeString(authorization[6:])
		if err != nil {
			return false
		}

		chunks := strings.Split(string(data), ":")
		if len(chunks) != 2 {
			return false
		}

		if server.AuthHandler.Authenticate(chunks[0], chunks[1]) {
			return true
		}
	}

	writer.Header().Add("WWW-Authenticate", "Basic realm=\"Authorization Required\"")

	return false
}

func (server *Server) handleError(writer http.ResponseWriter, status int) {
	var data struct {
		URLPrefix string
		Status    int
	}

	// Set template data
	data.URLPrefix = server.Config.URLPrefix
	data.Status = status

	// Execute template
	tmplData := bytes.NewBuffer(nil)

	tmpl, err := template.New("layout.html").Funcs(template.FuncMap{
		"asset": server.templateAsset,
		"eq":    templateEqual,
		"ne":    templateNotEqual,
	}).ParseFiles(
		path.Join(server.Config.BaseDir, "html", "layout.html"),
		path.Join(server.Config.BaseDir, "html", "error.html"),
	)
	if err == nil {
		err = tmpl.Execute(tmplData, data)
	}

	if err != nil {
		log.Println("ERROR: " + err.Error())
		server.handleResponse(writer, nil, status)
	}

	// Handle HTTP response with status code
	writer.WriteHeader(status)
	writer.Write(tmplData.Bytes())
}

func (server *Server) handleReload(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "GET" && request.Method != "HEAD" {
		server.handleResponse(writer, serverResponse{mesgMethodNotAllowed}, http.StatusMethodNotAllowed)
		return
	} else if !server.handleAuth(writer, request) {
		server.handleResponse(writer, serverResponse{mesgAuthenticationRequired}, http.StatusUnauthorized)
		return
	}

	server.Reload()

	server.handleResponse(writer, nil, http.StatusOK)
}

func (server *Server) handleResource(writer http.ResponseWriter, request *http.Request) {
	server.handleResponse(writer, &resourceResponse{
		Scales: server.Config.Scales,
	}, http.StatusOK)
}

func (server *Server) handleResponse(writer http.ResponseWriter, data interface{}, status int) {
	var err error

	output := make([]byte, 0)

	if data != nil {
		output, err = json.Marshal(data)
		if err != nil {
			server.handleResponse(writer, nil, http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	}

	writer.WriteHeader(status)

	if len(output) > 0 {
		writer.Write(output)
		writer.Write([]byte("\n"))
	}
}

func (server *Server) handleStatic(writer http.ResponseWriter, request *http.Request) {
	mimeType := mime.TypeByExtension(filepath.Ext(request.URL.Path))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	writer.Header().Set("Content-Type", mimeType)

	// Handle static files
	http.ServeFile(writer, request, path.Join(server.Config.BaseDir, request.URL.Path))
}

func (server *Server) handleStats(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "GET" && request.Method != "HEAD" {
		server.handleResponse(writer, serverResponse{mesgMethodNotAllowed}, http.StatusMethodNotAllowed)
		return
	}

	server.handleResponse(writer, server.getStats(writer, request), http.StatusOK)
}

func (server *Server) handleWait(writer http.ResponseWriter, request *http.Request) {
	var data struct {
		URLPrefix string
	}

	// Set template data
	data.URLPrefix = server.Config.URLPrefix

	// Execute template
	tmpl, err := template.New("layout.html").Funcs(template.FuncMap{
		"asset": server.templateAsset,
	}).ParseFiles(
		path.Join(server.Config.BaseDir, "html", "layout.html"),
		path.Join(server.Config.BaseDir, "html", "wait.html"),
	)
	if err == nil {
		err = tmpl.Execute(writer, data)
	}

	if err != nil {
		if os.IsNotExist(err) {
			server.handleError(writer, http.StatusNotFound)
		} else {
			log.Println("ERROR: " + err.Error())
			server.handleError(writer, http.StatusInternalServerError)
		}
	}
}

func (server *Server) getStats(writer http.ResponseWriter, request *http.Request) *statsResponse {
	sourceSet := set.New()
	metricSet := set.New()

	for _, origin := range server.Catalog.Origins {
		for key, source := range origin.Sources {
			sourceSet.Add(key)

			for key := range source.Metrics {
				metricSet.Add(key)
			}
		}
	}

	return &statsResponse{
		Origins:        len(server.Catalog.Origins),
		Sources:        sourceSet.Size(),
		Metrics:        metricSet.Size(),
		CatalogUpdated: server.Catalog.Updated.Format(time.RFC3339),

		Graphs:      len(server.Library.Graphs),
		Collections: len(server.Library.Collections),
		Groups:      len(server.Library.Groups),
	}
}

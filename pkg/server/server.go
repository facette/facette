package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"mime"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/facette/facette/pkg/auth"
	"github.com/facette/facette/pkg/catalog"
	"github.com/facette/facette/pkg/config"
	"github.com/facette/facette/pkg/library"
	"github.com/facette/facette/thirdparty/github.com/etix/stoppableListener"
	"github.com/facette/facette/thirdparty/github.com/gorilla/handlers"
	"github.com/facette/facette/thirdparty/github.com/gorilla/mux"
)

const (
	// URLAdminPath represents administration panel's base URL path
	URLAdminPath string = "/admin"
	// URLBrowsePath represents browse base URL path
	URLBrowsePath string = "/browse"
	// URLCatalogPath represents catalog's base URL path
	URLCatalogPath string = "/catalog"
	// URLLibraryPath represents library's base URL path
	URLLibraryPath string = "/library"
	// ServerStopWait represents the time to wait before force-closing connections when stopping
	ServerStopWait int = 5
)

type statusResponse struct {
	Message string `json:"message"`
}

// Server is the main service handler of Facette.
type Server struct {
	Config      *config.Config
	Listener    *stoppableListener.StoppableListener
	AuthHandler auth.Handler
	Catalog     *catalog.Catalog
	Library     *library.Library
	Loading     bool
	debugLevel  int
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
		server.handleResponse(writer, status)
	}

	// Handle HTTP response with status code
	writer.WriteHeader(status)
	writer.Write(tmplData.Bytes())
}

func (server *Server) handleJSON(writer http.ResponseWriter, data interface{}) {
	// Handle HTTP JSON response
	output, err := json.Marshal(data)
	if err != nil {
		log.Println("ERROR: " + err.Error())
		server.handleResponse(writer, http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.Write(output)
	writer.Write([]byte("\n"))
}

func (server *Server) handleResponse(writer http.ResponseWriter, status int) {
	// Handle HTTP response with status code
	http.Error(writer, "", status)
}

func (server *Server) handleStatic(writer http.ResponseWriter, request *http.Request) {
	mimeType := mime.TypeByExtension(filepath.Ext(request.URL.Path))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	writer.Header().Set("Content-Type", mimeType)

	// Handle static files
	http.ServeFile(writer, request, path.Join(server.Config.BaseDir, "static", request.URL.Path[7:]))
}

// LoadConfig loads the server configuration using confPath as configuration file path.
func (server *Server) LoadConfig(confPath string) error {
	return server.Config.Load(confPath)
}

// Reload reloads the configuration and refreshes both catalog and library.
func (server *Server) Reload() error {
	log.Printf("INFO: reloading configuration from `%s' file", server.Config.Path)

	server.Loading = true

	if err := server.Config.Reload(); err != nil {
		return err
	}

	if err := server.AuthHandler.Refresh(); err != nil {
		return err
	}

	server.Catalog.Refresh()
	server.Library.Update()

	server.Loading = false

	return nil
}

// Run starts the server, loading configuration and serving HTTP responses.
func (server *Server) Run() error {
	var accessOutput *os.File

	if server.Config == nil {
		return fmt.Errorf("configuration should be loaded first")
	}

	// Set server logging ouput
	if server.Config.ServerLog != "" {
		dirPath, _ := path.Split(server.Config.ServerLog)
		os.MkdirAll(dirPath, 0755)

		serverOutput, _ := os.OpenFile(server.Config.ServerLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		log.SetOutput(serverOutput)
	}

	// Prepare authentication backend
	authHandler, err := auth.NewAuth(server.Config.Auth, server.debugLevel)
	if err != nil {
		return err
	}

	server.AuthHandler = authHandler
	go server.AuthHandler.Refresh()

	log.Printf("INFO: server about to listen on %s", server.Config.BindAddr)

	// Initialize instance
	go server.Catalog.Refresh()
	go server.Library.Update()

	// Register routes
	router := mux.NewRouter()

	router.NotFoundHandler = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		server.handleError(writer, http.StatusNotFound)
	})

	router.PathPrefix("/static/").HandlerFunc(server.handleStatic)

	router.MatcherFunc(func(request *http.Request, match *mux.RouteMatch) bool {
		return server.Loading
	}).HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if strings.HasPrefix(request.URL.Path, URLAdminPath) || strings.HasPrefix(request.URL.Path, URLBrowsePath) {
			server.waitHandle(writer, request)
		} else if request.URL.Path == "/reload" {
			for {
				if !server.Loading {
					break
				}

				time.Sleep(time.Second)
			}

			server.handleResponse(writer, http.StatusOK)
		} else {
			server.handleResponse(writer, http.StatusServiceUnavailable)
		}
	})

	router.HandleFunc(URLCatalogPath+"/origins", server.originList)
	router.HandleFunc(URLCatalogPath+"/origins/{name}", server.originShow)
	router.HandleFunc(URLCatalogPath+"/sources", server.sourceList)
	router.HandleFunc(URLCatalogPath+"/sources/{name}", server.sourceShow)
	router.HandleFunc(URLCatalogPath+"/metrics", server.metricList)
	router.HandleFunc(URLCatalogPath+"/metrics/{path:.*}", server.metricShow)

	router.HandleFunc(URLLibraryPath+"/sourcegroups", server.groupHandle)
	router.HandleFunc(URLLibraryPath+"/sourcegroups/{id}", server.groupHandle)
	router.HandleFunc(URLLibraryPath+"/metricgroups", server.groupHandle)
	router.HandleFunc(URLLibraryPath+"/metricgroups/{id}", server.groupHandle)
	router.HandleFunc(URLLibraryPath+"/expand", server.groupExpand)

	router.HandleFunc(URLLibraryPath+"/graphs", server.graphHandle)
	router.HandleFunc(URLLibraryPath+"/graphs/plots", server.plotHandle)
	router.HandleFunc(URLLibraryPath+"/graphs/{id}", server.graphHandle)

	router.HandleFunc(URLLibraryPath+"/collections", server.collectionHandle)
	router.HandleFunc(URLLibraryPath+"/collections/{id}", server.collectionHandle)

	router.HandleFunc(URLAdminPath+"/", server.adminHandle)
	router.HandleFunc(URLAdminPath+"/{section}/{path:.*}", server.adminHandle)

	router.HandleFunc(URLBrowsePath+"/", server.browseHandle)
	router.HandleFunc(URLBrowsePath+"/sources/{source}", server.browseHandle)
	router.HandleFunc(URLBrowsePath+"/collections/{collection}", server.browseHandle)
	router.HandleFunc(URLBrowsePath+"/search", server.browseHandle)

	router.HandleFunc("/reload", server.reloadHandle)
	router.HandleFunc("/resources", server.resourceHandle)
	router.HandleFunc("/stats", server.statHandle)

	router.HandleFunc("/", server.browseHandle)

	// Set access logging output
	if server.Config.AccessLog == "" {
		accessOutput = os.Stderr
	} else {
		dirPath, _ := path.Split(server.Config.AccessLog)
		os.MkdirAll(dirPath, 0755)

		accessOutput, _ = os.OpenFile(server.Config.AccessLog, os.O_CREATE|os.O_WRONLY, 0644)
	}

	// Set HTTP handler
	http.Handle("/", handlers.CombinedLoggingHandler(accessOutput, router))

	// Start listener
	listener, err := net.Listen("tcp", server.Config.BindAddr)
	if err != nil {
		return err
	}

	server.Listener = stoppableListener.Handle(listener)
	err = http.Serve(server.Listener, nil)

	if server.Listener.Stopped {
		/* Wait for the clients to disconnect */
		for i := 0; i < ServerStopWait; i++ {
			if clientCount := server.Listener.ConnCount.Get(); clientCount == 0 {
				break
			}

			time.Sleep(time.Second)
		}

		clientCount := server.Listener.ConnCount.Get()

		if clientCount > 0 {
			log.Fatalf("INFO: server stopped after %d seconds with %d client(s) still connected", ServerStopWait,
				clientCount)
		} else {
			log.Println("INFO: server stopped gracefully")
		}
	} else if err != nil {
		return err
	}

	return nil
}

// Stop stops the server.
func (server *Server) Stop() {
	server.Listener.Stop <- true
}

// NewServer creates a new instance of Server.
func NewServer(debugLevel int) (*Server, error) {
	// Create new server instance
	server := &Server{
		Config:     &config.Config{},
		debugLevel: debugLevel,
	}

	server.Catalog = catalog.NewCatalog(server.Config, debugLevel)
	server.Library = library.NewLibrary(server.Config, server.Catalog, debugLevel)

	return server, nil
}

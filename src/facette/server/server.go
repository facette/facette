package server

import (
	"encoding/json"
	"facette/auth"
	"facette/backend"
	"facette/config"
	"facette/library"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
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
)

// Server is the main service handler of Facette.
type Server struct {
	Config     *config.Config
	Auth       *auth.Auth
	Catalog    *backend.Catalog
	Library    *library.Library
	Loading    bool
	debugLevel int
}

func (server *Server) handleJSON(writer http.ResponseWriter, data interface{}) {
	var (
		err    error
		output []byte
	)

	// Handle HTTP JSON response
	if output, err = json.Marshal(data); err != nil {
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
	// Handle static files
	http.ServeFile(writer, request, path.Join(server.Config.BaseDir, "share", "static", request.URL.Path[7:]))
}

// LoadConfig loads the server configuration using confPath as configuration file path.
func (server *Server) LoadConfig(confPath string) error {
	return server.Config.Load(confPath)
}

// Reload reloads the configuration and refreshes both catalog and library.
func (server *Server) Reload() error {
	var (
		err error
	)

	log.Printf("INFO: reloading configuration from `%s' file", server.Config.Path)

	server.Loading = true

	if err = server.Config.Reload(); err != nil {
		return err
	}

	server.Auth.Update()
	server.Catalog.Update()
	server.Library.Update()

	server.Loading = false

	return nil
}

// Run starts the server, loading configuration and serving HTTP responses.
func (server *Server) Run() error {
	var (
		accessOutput *os.File
		dirPath      string
		err          error
		router       *mux.Router
		serverOutput *os.File
	)

	if server.Config == nil {
		return fmt.Errorf("configuration should be loaded first")
	}

	// Set server logging ouput
	if server.Config.ServerLog != "" {
		dirPath, _ = path.Split(server.Config.ServerLog)
		os.MkdirAll(dirPath, 0755)

		serverOutput, _ = os.OpenFile(server.Config.ServerLog, os.O_CREATE|os.O_WRONLY, 0644)
		log.SetOutput(serverOutput)
	}

	log.Printf("INFO: server about to listen on %s", server.Config.BindAddr)

	// Get origins from configuration
	for originName, origin := range server.Config.Origins {
		if _, err = server.Catalog.AddOrigin(originName, origin.Backend); err != nil {
			log.Printf("ERROR: %s\n", err.Error())
		}
	}

	// Initialize instance
	go server.Auth.Update()
	go server.Catalog.Update()
	go server.Library.Update()

	// Register routes
	router = mux.NewRouter()

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

	router.HandleFunc(URLCatalogPath+"/expand", server.expandList)

	router.HandleFunc(URLLibraryPath+"/sourcegroups", server.groupHandle)
	router.HandleFunc(URLLibraryPath+"/sourcegroups/{id}", server.groupHandle)
	router.HandleFunc(URLLibraryPath+"/metricgroups", server.groupHandle)
	router.HandleFunc(URLLibraryPath+"/metricgroups/{id}", server.groupHandle)

	router.HandleFunc(URLLibraryPath+"/graphs", server.graphHandle)
	router.HandleFunc(URLLibraryPath+"/graphs/plots", server.plotHandle)
	router.HandleFunc(URLLibraryPath+"/graphs/values", server.plotValues)
	router.HandleFunc(URLLibraryPath+"/graphs/{id}", server.graphHandle)

	router.HandleFunc(URLLibraryPath+"/collections", server.collectionHandle)
	router.HandleFunc(URLLibraryPath+"/collections/{id}", server.collectionHandle)

	router.HandleFunc(URLAdminPath+"/", server.adminHandle)
	router.HandleFunc(URLAdminPath+"/{section}/{path:.*}", server.adminHandle)

	router.HandleFunc(URLBrowsePath+"/", server.browseHandle)
	router.HandleFunc(URLBrowsePath+"/sources/{source}", server.browseHandle)
	router.HandleFunc(URLBrowsePath+"/collections/{collection}", server.browseHandle)
	router.HandleFunc(URLBrowsePath+"/search", server.browseHandleSearch)

	router.HandleFunc("/reload", server.reloadHandle)

	router.HandleFunc("/stats", server.statHandle)

	router.HandleFunc("/", server.browseHandle)

	// Set access logging output
	if server.Config.AccessLog == "" {
		accessOutput = os.Stderr
	} else {
		dirPath, _ = path.Split(server.Config.AccessLog)
		os.MkdirAll(dirPath, 0755)

		accessOutput, _ = os.OpenFile(server.Config.AccessLog, os.O_CREATE|os.O_WRONLY, 0644)
	}

	// Run server instance
	http.Handle("/", handlers.CombinedLoggingHandler(accessOutput, router))
	return http.ListenAndServe(server.Config.BindAddr, nil)
}

// NewServer creates a new instance of Server.
func NewServer(debugLevel int) (*Server, error) {
	var (
		server *Server
	)

	// Create new server instance
	server = &Server{Config: &config.Config{}, debugLevel: debugLevel}
	server.Auth = auth.NewAuth(server.Config, debugLevel)
	server.Catalog = backend.NewCatalog(server.Config, debugLevel)
	server.Library = library.NewLibrary(server.Config, server.Catalog, debugLevel)

	return server, nil
}

// Package server implements the serving of the backend and the web UI.
package server

import (
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/facette/facette/pkg/catalog"
	"github.com/facette/facette/pkg/config"
	"github.com/facette/facette/pkg/library"
	"github.com/facette/facette/thirdparty/github.com/etix/stoppableListener"
)

const (
	serverStopWait  int    = 10
	urlAdminPath    string = "/admin/"
	urlBrowsePath   string = "/browse/"
	urlCatalogPath  string = "/catalog/"
	urlLibraryPath  string = "/library/"
	urlReloadPath   string = "/reload"
	urlResourcePath string = "/resources"
	urlStaticPath   string = "/static/"
	urlStatsPath    string = "/stats"
)

// Server is the main structure of the server handler.
type Server struct {
	Config     *config.Config
	Listener   *stoppableListener.StoppableListener
	Catalog    *catalog.Catalog
	Library    *library.Library
	Loading    bool
	StartTime  time.Time
	debugLevel int
}

// NewServer creates a new instance of server.
func NewServer(configPath string, debugLevel int) *Server {
	return &Server{
		Config:     &config.Config{Path: configPath},
		debugLevel: debugLevel,
	}
}

// Reload reloads the configuration and refreshes both catalog and library.
func (server *Server) Reload() error {
	log.Printf("NOTICE: reloading server")

	server.Loading = true

	if err := server.Config.Reload(); err != nil {
		log.Printf("ERROR: an error occured while reloading configuration: %s", err.Error())
		return err
	}

	server.Catalog.Refresh()
	server.Library.Refresh()

	server.Loading = false

	return nil
}

// Run starts the server serving the HTTP responses.
func (server *Server) Run() error {
	server.StartTime = time.Now()

	// Load server configuration
	if err := server.Config.Reload(); err != nil {
		return err
	}

	// Set server logging ouput
	if server.Config.ServerLog != "" && server.Config.ServerLog != "-" {
		dirPath, _ := path.Split(server.Config.ServerLog)
		os.MkdirAll(dirPath, 0755)

		serverOutput, err := os.OpenFile(server.Config.ServerLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Printf("ERROR: unable to open `%s' log file", server.Config.ServerLog)
			return err
		}

		defer serverOutput.Close()

		log.SetOutput(serverOutput)
	}

	// Handle pid file creation if set
	if server.Config.PidFile != "" {
		fd, err := os.OpenFile(server.Config.PidFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			log.Println("ERROR: unable to create pid file")
			return err
		}

		defer fd.Close()

		fd.Write([]byte(strconv.Itoa(os.Getpid()) + "\n"))
	}

	// Create catalog and library instances
	server.Catalog = catalog.NewCatalog(server.Config, server.debugLevel)
	go server.Catalog.Refresh()

	server.Library = library.NewLibrary(server.Config, server.Catalog, server.debugLevel)
	go server.Library.Refresh()

	// Prepare router
	router := NewRouter(server)

	router.HandleFunc(urlStaticPath, server.serveStatic)
	router.HandleFunc(urlCatalogPath, server.serveCatalog)
	router.HandleFunc(urlLibraryPath, server.serveLibrary)
	router.HandleFunc(urlAdminPath, server.serveAdmin)
	router.HandleFunc(urlBrowsePath, server.serveBrowse)
	router.HandleFunc(urlReloadPath, server.serveReload)
	router.HandleFunc(urlResourcePath, server.serveResource)
	router.HandleFunc(urlStatsPath, server.serveStats)

	router.HandleFunc("/", server.serveBrowse)

	http.Handle("/", router)

	// Start serving HTTP requests
	listener, err := net.Listen("tcp", server.Config.BindAddr)
	if err != nil {
		return err
	}

	log.Printf("INFO: server listening on %s", server.Config.BindAddr)

	server.Listener = stoppableListener.Handle(listener)
	err = http.Serve(server.Listener, nil)

	// Server shutdown triggered
	if server.Listener.Stopped {
		// Close catalog
		server.Catalog.Close()

		/* Wait for the clients to disconnect */
		for i := 0; i < serverStopWait; i++ {
			if clientCount := server.Listener.ConnCount.Get(); clientCount == 0 {
				break
			}

			time.Sleep(time.Second)
		}

		clientCount := server.Listener.ConnCount.Get()

		if clientCount > 0 {
			log.Printf("INFO: server stopped after %d seconds with %d client(s) still connected",
				serverStopWait,
				clientCount)
		} else {
			log.Println("INFO: server stopped gracefully")
		}

		// Remove pid file
		if server.Config.PidFile != "" {
			os.Remove(server.Config.PidFile)
		}
	} else if err != nil {
		return err
	}

	return nil
}

// Stop stops the server.
func (server *Server) Stop() {
	log.Printf("NOTICE: shutting down server")

	server.Listener.Stop <- true
}

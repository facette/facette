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

	"github.com/facette/facette/pkg/auth"
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
	Config      *config.Config
	Listener    *stoppableListener.StoppableListener
	AuthHandler auth.Handler
	Catalog     *catalog.Catalog
	Library     *library.Library
	Loading     bool
	debugLevel  int
}

// Reload reloads the configuration and refreshes authentication handler, catalog and library.
func (server *Server) Reload() error {
	server.Loading = true

	if err := server.Config.Reload(); err != nil {
		log.Printf("ERROR: an error occued while reloading configuration: %s", err.Error())
		return err
	}

	if server.AuthHandler != nil {
		server.AuthHandler.Refresh()
	}

	server.Catalog.Refresh()
	server.Library.Refresh()

	server.Loading = false

	return nil
}

// Run starts the server serving the HTTP responses.
func (server *Server) Run() error {
	// Load server configuration
	if err := server.Config.Reload(); err != nil {
		return err
	}

	// Set server logging ouput
	if server.Config.ServerLog != "" {
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

	// Create authentication handler
	authHandler, err := auth.NewAuth(server.Config.Auth, server.debugLevel)
	if err != nil {
		return err
	}

	if authHandler != nil {
		server.AuthHandler = authHandler
		go server.AuthHandler.Refresh()
	}

	// Prepare router
	router := NewRouter(server)

	router.HandleFunc(urlStaticPath, server.handleStatic)
	router.HandleFunc(urlCatalogPath, server.handleCatalog)
	router.HandleFunc(urlLibraryPath, server.handleLibrary)
	router.HandleFunc(urlAdminPath, server.handleAdmin)
	router.HandleFunc(urlBrowsePath, server.handleBrowse)
	router.HandleFunc(urlReloadPath, server.handleReload)
	router.HandleFunc(urlResourcePath, server.handleResource)
	router.HandleFunc(urlStatsPath, server.handleStats)

	router.HandleFunc("/", server.handleBrowse)

	http.Handle("/", router)

	// Start serving HTTP requests
	listener, err := net.Listen("tcp", server.Config.BindAddr)
	if err != nil {
		return err
	}

	log.Printf("INFO: server listening on %s", server.Config.BindAddr)

	server.Listener = stoppableListener.Handle(listener)
	err = http.Serve(server.Listener, nil)

	if server.Listener.Stopped {
		/* Wait for the clients to disconnect */
		for i := 0; i < serverStopWait; i++ {
			if clientCount := server.Listener.ConnCount.Get(); clientCount == 0 {
				break
			}

			time.Sleep(time.Second)
		}

		clientCount := server.Listener.ConnCount.Get()

		if clientCount > 0 {
			log.Printf("INFO: server stopped after %d seconds with %d client(s) still connected", serverStopWait,
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
	server.Listener.Stop <- true
}

// NewServer creates a new instance of server.
func NewServer(configPath string, debugLevel int) *Server {
	return &Server{
		Config:     &config.Config{Path: configPath},
		debugLevel: debugLevel,
	}
}

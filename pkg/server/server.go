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
	// URLCatalogPath represents the catalog base URL path
	URLCatalogPath string = "/catalog"
	// URLLibraryPath represents the library base URL path
	URLLibraryPath string = "/library"
	// ServerStopWait represents the time to wait before force-closing connections when stopping server.
	ServerStopWait int = 10
)

// Server is the main structure of the server handler.
type Server struct {
	Config      *config.Config
	Listener    *stoppableListener.StoppableListener
	AuthHandler auth.Handler
	Catalog     *catalog.Catalog
	Library     *library.Library
	debugLevel  int
}

// Reload reloads the configuration and refreshes authentication handler, catalog and library.
func (server *Server) Reload() error {
	if err := server.Config.Reload(); err != nil {
		log.Printf("ERROR: an error occued while reloading configuration: %s", err.Error())
		return err
	}

	server.AuthHandler.Refresh()
	server.Catalog.Refresh()
	server.Library.Refresh()

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

	server.AuthHandler = authHandler
	go server.AuthHandler.Refresh()

	// Prepare router
	router := NewRouter(server.debugLevel)

	router.HandleFunc(URLCatalogPath+"/origins/", server.handleOrigin)
	router.HandleFunc(URLCatalogPath+"/sources/", server.handleSource)
	router.HandleFunc(URLCatalogPath+"/metrics/", server.handleMetric)

	router.HandleFunc(URLLibraryPath+"/sourcegroups/", server.handleGroup)
	router.HandleFunc(URLLibraryPath+"/metricgroups/", server.handleGroup)
	router.HandleFunc(URLLibraryPath+"/expand", server.handleGroupExpand)
	router.HandleFunc(URLLibraryPath+"/graphs/plots", server.handleGraphPlots)
	router.HandleFunc(URLLibraryPath+"/graphs/", server.handleGraph)
	router.HandleFunc(URLLibraryPath+"/collections/", server.handleCollection)

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
		for i := 0; i < ServerStopWait; i++ {
			if clientCount := server.Listener.ConnCount.Get(); clientCount == 0 {
				break
			}

			time.Sleep(time.Second)
		}

		clientCount := server.Listener.ConnCount.Get()

		if clientCount > 0 {
			log.Printf("INFO: server stopped after %d seconds with %d client(s) still connected", ServerStopWait,
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

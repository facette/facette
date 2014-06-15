// Package server implements the serving of the backend and the web UI.
package server

import (
	"fmt"
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
	"github.com/facette/facette/pkg/provider"
	"github.com/facette/facette/pkg/worker"
	"github.com/facette/facette/thirdparty/github.com/etix/stoppableListener"
)

const (
	serverStopWait int    = 10
	urlStaticPath  string = "/static/"
	urlAdminPath   string = "/admin/"
	urlBrowsePath  string = "/browse/"
	urlReloadPath  string = "/reload"
	urlCatalogPath string = "/api/v1/catalog/"
	urlLibraryPath string = "/api/v1/library/"
	urlStatsPath   string = "/api/v1/stats"
)

// Server is the main structure of the server handler.
type Server struct {
	Config          *config.Config
	Listener        *stoppableListener.StoppableListener
	Catalog         *catalog.Catalog
	Library         *library.Library
	Loading         bool
	StartTime       time.Time
	providers       map[string]*provider.Provider
	providerWorkers worker.WorkerPool
	catalogWorker   *worker.Worker
	debugLevel      int
}

// NewServer creates a new instance of server.
func NewServer(configPath, logPath string, debugLevel int) *Server {
	return &Server{
		Config:     &config.Config{Path: configPath, LogFile: logPath},
		providers:  make(map[string]*provider.Provider),
		debugLevel: debugLevel,
	}
}

// Reload reloads the configuration and refreshes both catalog and library.
func (server *Server) Reload(config bool) error {
	log.Printf("NOTICE: reloading server")

	server.Loading = true

	if config {
		if err := server.Config.Reload(); err != nil {
			log.Printf("ERROR: an error occurred while reloading configuration: %s", err.Error())
			return err
		}
	}

	server.providerWorkers.Broadcast(eventCatalogRefresh, nil)
	server.Library.Refresh()

	server.Loading = false

	return nil
}

// Run starts the server serving the HTTP responses.
func (server *Server) Run() error {
	server.StartTime = time.Now()

	// Set server logging ouput
	if server.Config.LogFile != "" && server.Config.LogFile != "-" {
		dirPath, _ := path.Split(server.Config.LogFile)
		os.MkdirAll(dirPath, 0755)

		serverOutput, err := os.OpenFile(server.Config.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Printf("ERROR: unable to open log file: %s", err)
			return err
		}

		defer serverOutput.Close()

		log.SetOutput(serverOutput)
	}

	// Load server configuration
	if err := server.Config.Reload(); err != nil {
		log.Printf("ERROR: unable to load configuration: %s", err)
		return err
	}

	// Handle pid file creation if set
	if server.Config.PidFile != "" {
		fd, err := os.OpenFile(server.Config.PidFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("unable to create pid file `%s'", server.Config.PidFile)
		}

		defer fd.Close()

		fd.Write([]byte(strconv.Itoa(os.Getpid()) + "\n"))
	}

	// Create new catalog instance
	server.Catalog = catalog.NewCatalog(server.debugLevel)

	// Instanciate catalog worker
	server.catalogWorker = worker.NewWorker()
	server.catalogWorker.RegisterEvent(eventInit, workerCatalogInit)
	server.catalogWorker.RegisterEvent(eventShutdown, workerCatalogShutdown)
	server.catalogWorker.RegisterEvent(eventRun, workerCatalogRun)

	if err := server.catalogWorker.SendEvent(eventInit, false, server.Catalog); err != nil {
		return err
	}

	server.catalogWorker.SendEvent(eventRun, true, nil)

	// Instanciate providers
	for providerName, providerConfig := range server.Config.Providers {
		server.providers[providerName] = provider.NewProvider(providerName, providerConfig, server.Catalog)
	}

	if err := server.startOriginWorkers(); err != nil {
		return err
	}

	// Send initial catalog refresh event to origin workers
	server.providerWorkers.Broadcast(eventCatalogRefresh, nil)

	// Create library instance
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
		// Shutdown running origin workers
		server.StopOriginWorkers()

		// Shutdown catalog worker
		if err := server.catalogWorker.SendEvent(eventShutdown, false, nil); err != nil {
			log.Printf("WARNING: catalog worker did not shut down successfully: %s", err)
		}

		// Close catalog
		server.Catalog.Close()

		/* Wait for the clients to disconnect */
		for i := 0; i < serverStopWait; i++ {
			if clientCount := server.Listener.ConnCount.Get(); clientCount == 0 {
				break
			}

			time.Sleep(time.Second) // TODO: WTF? Use a waitgroup or proper timeout system
		}

		clientCount := server.Listener.ConnCount.Get()

		if clientCount > 0 {
			log.Printf("INFO: server stopped after %d second(s) with %d client(s) still connected",
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

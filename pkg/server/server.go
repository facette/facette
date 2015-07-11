// Package server implements the serving of the backend and the web UI.
package server

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/facette/facette/pkg/catalog"
	"github.com/facette/facette/pkg/config"
	"github.com/facette/facette/pkg/connector"
	"github.com/facette/facette/pkg/library"
	"github.com/facette/facette/pkg/logger"
	"github.com/facette/facette/pkg/provider"
	"github.com/facette/facette/pkg/worker"
	uuid "github.com/nu7hatch/gouuid"
)

var (
	version   string
	buildDate string
)

// Server is the main structure of the server handler.
type Server struct {
	ID              string
	Config          *config.Config
	Catalog         *catalog.Catalog
	Library         *library.Library
	providers       map[string]*provider.Provider
	providerWorkers worker.Pool
	catalogWorker   *worker.Worker
	serveWorker     *worker.Worker
	configPath      string
	logPath         string
	logLevel        int
	startTime       time.Time
	stopping        bool
	wg              *sync.WaitGroup
	buildInfo       *buildInfo
}

// NewServer creates a new instance of server.
func NewServer(configPath, logPath string, logLevel int) *Server {
	return &Server{
		Config: &config.Config{
			BindAddr:     config.DefaultBindAddr,
			BaseDir:      config.DefaultBaseDir,
			DataDir:      config.DefaultDataDir,
			ProvidersDir: config.DefaultProvidersDir,
			PidFile:      config.DefaultPidFile,
			SocketUser:   config.DefaultSocketUser,
			SocketGroup:  config.DefaultSocketGroup,
		},
		configPath: configPath,
		logPath:    logPath,
		logLevel:   logLevel,
		providers:  make(map[string]*provider.Provider),
		wg:         &sync.WaitGroup{},
	}
}

// Refresh refreshes both catalog and library.
func (server *Server) Refresh() {
	server.providerWorkers.Broadcast(eventCatalogRefresh, nil)
	server.Library.Refresh()
}

// Run starts the server serving the HTTP responses.
func (server *Server) Run() error {
	server.startTime = time.Now()

	// Set server build information
	server.buildInfo = &buildInfo{
		Version:    version,
		BuildDate:  buildDate,
		Compiler:   fmt.Sprintf("%s (%s)", runtime.Compiler, runtime.Version()),
		Connectors: make([]string, 0),
	}

	for connector := range connector.Connectors {
		server.buildInfo.Connectors = append(server.buildInfo.Connectors, connector)
	}

	sort.Strings(server.buildInfo.Connectors)

	// Set up server logging
	if server.logPath != "" && server.logPath != "-" {
		dirPath, _ := path.Split(server.logPath)
		os.MkdirAll(dirPath, 0755)

		serverOutput, err := os.OpenFile(server.logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			logger.Log(logger.LevelError, "server", "unable to open log file: %s", err)
			return err
		}

		defer serverOutput.Close()

		logger.SetOutput(serverOutput)
	}

	logger.SetLevel(server.logLevel)

	// Load server configuration
	if err := server.Config.Load(server.configPath); err != nil {
		logger.Log(logger.LevelError, "server", "unable to load configuration: %s", err)
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

	server.wg.Add(1)

	// Generate unique server instance identifier
	uuidTemp, err := uuid.NewV4()
	if err != nil {
		return err
	}

	server.ID = uuidTemp.String()

	// Create new catalog instance
	server.Catalog = catalog.NewCatalog()

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

	if err := server.startProviderWorkers(); err != nil {
		return err
	}

	// Send initial catalog refresh event to provider workers
	server.providerWorkers.Broadcast(eventCatalogRefresh, nil)

	// Create library instance
	server.Library = library.NewLibrary(server.Config, server.Catalog)
	go server.Library.Refresh()

	// Instanciate serve worker
	server.serveWorker = worker.NewWorker()
	server.serveWorker.RegisterEvent(eventInit, workerServeInit)
	server.serveWorker.RegisterEvent(eventShutdown, workerServeShutdown)
	server.serveWorker.RegisterEvent(eventRun, workerServeRun)

	if err := server.serveWorker.SendEvent(eventInit, false, server); err != nil {
		return err
	} else if err := server.serveWorker.SendEvent(eventRun, false, nil); err != nil {
		return err
	}

	server.wg.Wait()

	return nil
}

// Stop stops the server.
func (server *Server) Stop() {
	if server.stopping {
		return
	}

	logger.Log(logger.LevelNotice, "server", "shutting down server")

	server.stopping = true

	// Shutdown serve worker
	if err := server.serveWorker.SendEvent(eventShutdown, false, nil); err != nil {
		logger.Log(logger.LevelWarning, "server", "serve worker did not shut down successfully: %s", err)
	}

	// Shutdown running provider workers
	server.stopProviderWorkers()

	// Shutdown catalog worker
	if err := server.catalogWorker.SendEvent(eventShutdown, false, nil); err != nil {
		logger.Log(logger.LevelWarning, "server", "catalog worker did not shut down successfully: %s", err)
	}

	server.Catalog.Close()

	// Remove pid file
	if server.Config.PidFile != "" {
		logger.Log(logger.LevelDebug, "server", "removing `%s' pid file", server.Config.PidFile)

		os.Remove(server.Config.PidFile)
	}

	server.wg.Done()
}

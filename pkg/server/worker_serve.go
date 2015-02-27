package server

import (
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/facette/facette/pkg/logger"
	"github.com/facette/facette/pkg/worker"
)

const (
	urlStaticPath  string = "/static/"
	urlAdminPath   string = "/admin/"
	urlBrowsePath  string = "/browse/"
	urlShowPath    string = "/show/"
	urlCatalogPath string = "/api/v1/catalog/"
	urlLibraryPath string = "/api/v1/library/"
	urlPlotsPath   string = "/api/v1/plots"
	urlStatsPath   string = "/api/v1/stats"
)

func workerServeInit(w *worker.Worker, args ...interface{}) {
	var server = args[0].(*Server)

	logger.Log(logger.LevelDebug, "serveWorker", "init")

	// Worker properties:
	// 0: server instance (*Server)
	w.Props = append(w.Props, server)

	w.ReturnErr(nil)
}

func workerServeShutdown(w *worker.Worker, args ...interface{}) {
	logger.Log(logger.LevelDebug, "serveWorker", "shutdown")

	w.SendJobSignal(jobSignalShutdown)

	w.ReturnErr(nil)
}

func workerServeRun(w *worker.Worker, args ...interface{}) {
	var server = w.Props[0].(*Server)

	defer w.Shutdown()

	logger.Log(logger.LevelDebug, "serveWorker", "starting")

	// Prepare router
	router := NewRouter(server)

	router.HandleFunc(urlStaticPath, server.serveStatic)
	router.HandleFunc(urlCatalogPath, server.serveCatalog)
	router.HandleFunc(urlLibraryPath, server.serveLibrary)
	router.HandleFunc(urlPlotsPath, server.servePlots)
	router.HandleFunc(urlAdminPath, server.serveAdmin)
	router.HandleFunc(urlBrowsePath, server.serveBrowse)
	router.HandleFunc(urlShowPath, server.serveShow)
	router.HandleFunc(urlStatsPath, server.serveStats)

	router.HandleFunc("/", server.serveBrowse)

	http.Handle("/", router)

	// Start serving HTTP requests
	netType := "tcp"
	address := server.Config.BindAddr
	for _, scheme := range [...]string{"tcp", "tcp4", "tcp6", "unix"} {
		prefix := scheme + "://"

		if strings.HasPrefix(address, prefix) {
			netType = scheme
			address = strings.TrimPrefix(address, prefix)
			break
		}
	}

	listener, err := net.Listen(netType, address)
	if err != nil {
		w.ReturnErr(err)
		return
	}

	logger.Log(logger.LevelInfo, "serveWorker", "listening on %s", server.Config.BindAddr)

	if netType == "unix" {
		// Change owning user and group
		if server.Config.SocketUser >= 0 || server.Config.SocketGroup >= 0 {
			logger.Log(logger.LevelDebug, "serveWorker", "changing ownership of unix socket to UID %v and GID %v",
				server.Config.SocketUser, server.Config.SocketGroup)
			err = os.Chown(address, server.Config.SocketUser, server.Config.SocketGroup)
			if err != nil {
				listener.Close()
				w.ReturnErr(err)
				return
			}
		}

		// Change mode
		if server.Config.SocketMode != nil {
			mode, err := strconv.ParseUint(*server.Config.SocketMode, 8, 32)
			if err != nil {
				logger.Log(logger.LevelError, "serveWorker", "socket_mode is invalid")
				listener.Close()
				w.ReturnErr(err)
				return
			}

			logger.Log(logger.LevelDebug, "serveWorker", "changing file permissions mode of unix socket to %04o", mode)
			err = os.Chmod(address, os.FileMode(mode))
			if err != nil {
				listener.Close()
				w.ReturnErr(err)
				return
			}
		}
	}

	go http.Serve(listener, nil)

	for {
		select {
		case cmd := <-w.ReceiveJobSignals():
			switch cmd {
			case jobSignalShutdown:
				logger.Log(logger.LevelInfo, "serveWorker", "received shutdown command, stopping job")

				listener.Close()

				logger.Log(logger.LevelInfo, "serveWorker", "server listener closed")

				w.State = worker.JobStopped

				return

			default:
				logger.Log(logger.LevelInfo, "serveWorker", "received unknown command, ignoring")
			}
		}
	}

	w.ReturnErr(nil)
}

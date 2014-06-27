package server

import (
	"net"
	"net/http"

	"github.com/facette/facette/pkg/logger"
	"github.com/facette/facette/pkg/worker"
)

const (
	urlStaticPath  string = "/static/"
	urlAdminPath   string = "/admin/"
	urlBrowsePath  string = "/browse/"
	urlShowPath    string = "/show/"
	urlReloadPath  string = "/reload"
	urlCatalogPath string = "/api/v1/catalog/"
	urlLibraryPath string = "/api/v1/library/"
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
	router.HandleFunc(urlAdminPath, server.serveAdmin)
	router.HandleFunc(urlBrowsePath, server.serveBrowse)
	router.HandleFunc(urlShowPath, server.serveShow)
	router.HandleFunc(urlReloadPath, server.serveReload)
	router.HandleFunc(urlStatsPath, server.serveStats)

	router.HandleFunc("/", server.serveBrowse)

	http.Handle("/", router)

	// Start serving HTTP requests
	listener, err := net.Listen("tcp", server.Config.BindAddr)
	if err != nil {
		w.ReturnErr(err)
		return
	}

	logger.Log(logger.LevelInfo, "serveWorker", "listening on %s", server.Config.BindAddr)

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

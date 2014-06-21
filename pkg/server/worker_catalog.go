package server

import (
	"github.com/facette/facette/pkg/catalog"
	"github.com/facette/facette/pkg/logger"
	"github.com/facette/facette/pkg/worker"
)

func workerCatalogInit(w *worker.Worker, args ...interface{}) {
	var catalog = args[0].(*catalog.Catalog)

	logger.Log(logger.LevelDebug, "catalogWorker", "init")

	// Worker properties:
	// 0: catalog instance (*catalog.Catalog)
	w.Props = append(w.Props, catalog)

	w.ReturnErr(nil)
}

func workerCatalogShutdown(w *worker.Worker, args ...interface{}) {
	logger.Log(logger.LevelDebug, "catalogWorker", "shutdown")

	w.SendJobSignal(jobSignalShutdown)

	w.ReturnErr(nil)
}

func workerCatalogRun(w *worker.Worker, args ...interface{}) {
	var serverCatalog = w.Props[0].(*catalog.Catalog)

	defer w.Shutdown()

	logger.Log(logger.LevelDebug, "catalogWorker", "starting")

	w.State = worker.JobStarted

	for {
		select {
		case cmd := <-w.ReceiveJobSignals():
			switch cmd {
			case jobSignalShutdown:
				logger.Log(logger.LevelInfo, "catalogWorker", "received shutdown command, stopping job")

				w.State = worker.JobStopped

				w.ReturnErr(nil)

				return

			default:
				logger.Log(logger.LevelNotice, "catalogWorker", "received unknown command, ignoring")
			}

		case record := <-serverCatalog.RecordChan:
			serverCatalog.Insert(record)
		}
	}
}

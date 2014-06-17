package server

import (
	"log"

	"github.com/facette/facette/pkg/catalog"
	"github.com/facette/facette/pkg/worker"
)

func workerCatalogInit(w *worker.Worker, args ...interface{}) {
	var catalog = args[0].(*catalog.Catalog)

	log.Println("DEBUG: catalogWorker: init")

	// Worker properties:
	// 0: catalog instance (*catalog.Catalog)
	w.Props = append(w.Props, catalog)

	w.ReturnErr(nil)
}

func workerCatalogShutdown(w *worker.Worker, args ...interface{}) {
	log.Println("DEBUG: catalogWorker: shutdown")

	w.SendJobSignal(jobSignalShutdown)

	w.ReturnErr(nil)
}

func workerCatalogRun(w *worker.Worker, args ...interface{}) {
	var serverCatalog = w.Props[0].(*catalog.Catalog)

	defer w.Shutdown()

	log.Println("DEBUG: catalogWorker: starting")

	w.State = worker.JobStarted

	for {
		select {
		case cmd := <-w.ReceiveJobSignals():
			switch cmd {
			case jobSignalShutdown:
				log.Println("INFO: catalogWorker: received shutdown command, stopping job")

				w.State = worker.JobStopped

				w.ReturnErr(nil)

				return

			default:
				log.Println("NOTICE: catalogWorker: received unknown command, ignoring")
			}

		case record := <-serverCatalog.RecordChan:
			serverCatalog.Insert(record)
		}
	}
}

// Package server implements the serving of the backend and the web UI.
package server

import (
	"fmt"
	"log"
	"time"

	"github.com/facette/facette/pkg/catalog"
	"github.com/facette/facette/pkg/config"
	"github.com/facette/facette/pkg/connector"
	"github.com/facette/facette/pkg/worker"
)

const (
	_ = iota
	eventInit
	eventRun
	eventCatalogRefresh
	eventShutdown

	_ = iota
	jobSignalRefresh
	jobSignalShutdown
)

func (server *Server) startOriginWorkers() error {
	server.originWorkers = worker.NewWorkerPool()

	log.Println("DEBUG: declaring origin workers")

	for _, origin := range server.Catalog.Origins {
		connectorType, err := config.GetString(origin.Config.Connector, "type", true)
		if err != nil {
			return fmt.Errorf("origin `%s' connector: %s", origin.Name, err)
		} else if _, ok := connector.Connectors[connectorType]; !ok {
			return fmt.Errorf("origin `%s' uses unknown connector type `%s'", origin.Name, connectorType)
		}

		originWorker := worker.NewWorker()
		originWorker.RegisterEvent(eventInit, workerOriginInit)
		originWorker.RegisterEvent(eventShutdown, workerOriginShutdown)
		originWorker.RegisterEvent(eventRun, workerOriginRun)
		originWorker.RegisterEvent(eventCatalogRefresh, workerOriginRefresh)

		server.originWorkers.Add(originWorker)

		if err := originWorker.SendEvent(eventInit, false, origin, connectorType); err != nil {
			log.Printf("ERROR: in origin `%s', %s", origin.Name, err.Error())
			log.Printf("WARNING: discarding origin `%s'", origin.Name)
			continue
		}

		originWorker.SendEvent(eventRun, true, nil)

		log.Printf("DEBUG: declared origin worker `%s'", origin.Name)
	}

	return nil
}

// StopOriginWorkers stop all running origin workers.
func (server *Server) StopOriginWorkers() {
	server.originWorkers.Broadcast(eventShutdown, nil)

	// Wait for all workers to shut down
	server.originWorkers.Wg.Wait()
}

func workerOriginInit(w *worker.Worker, args ...interface{}) {
	var (
		origin        = args[0].(*catalog.Origin)
		connectorType = args[1].(string)
	)

	log.Printf("DEBUG: originWorker[%s]: init", origin.Name)

	// Instanciate the connector according to its type
	connector, err := connector.Connectors[connectorType](origin)
	if err != nil {
		w.ErrorChan <- err
	}

	// Worker properties:
	// 0: origin instance (catalog.Origin)
	// 1: connector instance (*connector.Connector)
	w.Props = append(w.Props, origin, connector)

	w.ErrorChan <- nil
}

func workerOriginShutdown(w *worker.Worker, args ...interface{}) {
	var origin = w.Props[0].(*catalog.Origin)

	log.Printf("DEBUG: originWorker[%s]: shutdown", origin.Name)

	w.SendJobSignal(jobSignalShutdown)
}

func workerOriginRun(w *worker.Worker, args ...interface{}) {
	var (
		origin     = w.Props[0].(*catalog.Origin)
		connector  = w.Props[1].(connector.Connector)
		timeTicker *time.Ticker
		timeChan   <-chan time.Time
	)

	defer func() { w.State = worker.JobStopped }()
	defer w.Shutdown()

	log.Printf("DEBUG: originWorker[%s]: starting", origin.Name)

	// If origin `refresh_interval` has been configured, set up a time ticker
	if origin.Config.RefreshInterval > 0 {
		timeTicker = time.NewTicker(time.Duration(origin.Config.RefreshInterval) * time.Second)
		timeChan = timeTicker.C
	}

	for {
		select {
		case _ = <-timeChan:
			if err := connector.Refresh(origin); err != nil {
				w.ErrorChan <- fmt.Errorf("unable to refresh origin `%s': %s", origin.Name, err)
			}

			origin.LastRefresh = time.Now()
			w.ErrorChan <- nil

		case cmd := <-w.ReceiveJobSignals():
			switch cmd {
			case jobSignalRefresh:
				log.Printf("INFO: originWorker[%s]: received refresh command", origin.Name)
				if err := connector.Refresh(origin); err != nil {
					w.ErrorChan <- fmt.Errorf("unable to refresh origin `%s': %s", origin.Name, err)
				}

				origin.LastRefresh = time.Now()
				w.ErrorChan <- nil

			case jobSignalShutdown:
				log.Printf("INFO: originWorker[%s]: received shutdown command, stopping job", origin.Name)

				w.State = worker.JobStopped

				if timeTicker != nil {
					// Stop refresh time ticker
					timeTicker.Stop()
				}

				return

			default:
				log.Println("NOTICE: originWorker[%s]: received unknown command, ignoring", origin.Name)
			}
		}
	}
}

func workerOriginRefresh(w *worker.Worker, args ...interface{}) {
	w.SendJobSignal(jobSignalRefresh)

	w.ErrorChan <- nil
}

func workerCatalogInit(w *worker.Worker, args ...interface{}) {
	var catalog = args[0].(*catalog.Catalog)

	log.Println("DEBUG: catalogWorker: init")

	// Worker properties:
	// 0: catalog instance (*catalog.Catalog)
	w.Props = append(w.Props, catalog)
}

func workerCatalogShutdown(w *worker.Worker, args ...interface{}) {
	log.Println("DEBUG: catalogWorker: shutdown")

	w.SendJobSignal(jobSignalShutdown)

	w.ErrorChan <- nil
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

				w.ErrorChan <- nil

				return

			default:
				log.Println("NOTICE: catalogWorker: received unknown command, ignoring")
			}

		case record := <-serverCatalog.RecordChan:
			// TODO: filter
			serverCatalog.Insert(record.Origin, record.Source, record.Metric)
		}
	}
}

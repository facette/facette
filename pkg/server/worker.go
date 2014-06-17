// Package server implements the serving of the backend and the web UI.
package server

import (
	"fmt"
	"log"
	"time"

	"github.com/facette/facette/pkg/catalog"
	"github.com/facette/facette/pkg/config"
	"github.com/facette/facette/pkg/connector"
	"github.com/facette/facette/pkg/provider"
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

func (server *Server) startProviderWorkers() error {
	server.providerWorkers = worker.NewWorkerPool()

	log.Println("DEBUG: declaring providers")

	for _, prov := range server.providers {
		connectorType, err := config.GetString(prov.Config.Connector, "type", true)
		if err != nil {
			return fmt.Errorf("provider `%s' connector: %s", prov.Name, err)
		} else if _, ok := connector.Connectors[connectorType]; !ok {
			return fmt.Errorf("provider `%s' uses unknown connector type `%s'", prov.Name, connectorType)
		}

		providerWorker := worker.NewWorker()
		providerWorker.RegisterEvent(eventInit, workerProviderInit)
		providerWorker.RegisterEvent(eventShutdown, workerProviderShutdown)
		providerWorker.RegisterEvent(eventRun, workerProviderRun)
		providerWorker.RegisterEvent(eventCatalogRefresh, workerProviderRefresh)

		if err := providerWorker.SendEvent(eventInit, false, prov, connectorType); err != nil {
			log.Printf("ERROR: in provider `%s', %s", prov.Name, err.Error())
			log.Printf("WARNING: discarding provider `%s'", prov.Name)
			continue
		}

		// Add worker into pool if initialization went fine
		server.providerWorkers.Add(providerWorker)

		providerWorker.SendEvent(eventRun, true, nil)

		log.Printf("DEBUG: declared provider `%s'", prov.Name)
	}

	return nil
}

func (server *Server) stopProviderWorkers() {
	server.providerWorkers.Broadcast(eventShutdown, nil)

	// Wait for all workers to shut down
	server.providerWorkers.Wg.Wait()
}

func workerProviderInit(w *worker.Worker, args ...interface{}) {
	var (
		prov          = args[0].(*provider.Provider)
		connectorType = args[1].(string)
	)

	log.Printf("DEBUG: providerWorker[%s]: init", prov.Name)

	// Instanciate the connector according to its type
	conn, err := connector.Connectors[connectorType](prov.Config.Connector)
	if err != nil {
		w.ReturnErr(err)
		return
	}

	prov.Connector = conn.(connector.Connector)

	// Worker properties:
	// 0: provider instance (*provider.Provider)
	w.Props = append(w.Props, prov)

	w.ReturnErr(nil)
}

func workerProviderShutdown(w *worker.Worker, args ...interface{}) {
	var prov = w.Props[0].(*provider.Provider)

	log.Printf("DEBUG: providerWorker[%s]: shutdown", prov.Name)

	w.SendJobSignal(jobSignalShutdown)
}

func workerProviderRun(w *worker.Worker, args ...interface{}) {
	var (
		prov       = w.Props[0].(*provider.Provider)
		timeTicker *time.Ticker
		timeChan   <-chan time.Time
	)

	defer func() { w.State = worker.JobStopped }()
	defer w.Shutdown()

	log.Printf("DEBUG: providerWorker[%s]: starting", prov.Name)

	// If provider `refresh_interval` has been configured, set up a time ticker
	if prov.Config.RefreshInterval > 0 {
		timeTicker = time.NewTicker(time.Duration(prov.Config.RefreshInterval) * time.Second)
		timeChan = timeTicker.C
	}

	for {
		select {
		case _ = <-timeChan:
			if err := prov.Connector.Refresh(prov.Name, prov.Filters.Input); err != nil {
				log.Printf("ERROR: unable to refresh provider `%s': %s", prov.Name, err)
				continue
			}

			prov.LastRefresh = time.Now()

		case cmd := <-w.ReceiveJobSignals():
			switch cmd {
			case jobSignalRefresh:
				log.Printf("INFO: providerWorker[%s]: received refresh command", prov.Name)

				if err := prov.Connector.Refresh(prov.Name, prov.Filters.Input); err != nil {
					log.Printf("ERROR: unable to refresh provider `%s': %s", prov.Name, err)
					continue
				}

				prov.LastRefresh = time.Now()

			case jobSignalShutdown:
				log.Printf("INFO: providerWorker[%s]: received shutdown command, stopping job", prov.Name)

				w.State = worker.JobStopped

				if timeTicker != nil {
					// Stop refresh time ticker
					timeTicker.Stop()
				}

				return

			default:
				log.Println("NOTICE: providerWorker[%s]: received unknown command, ignoring", prov.Name)
			}
		}
	}
}

func workerProviderRefresh(w *worker.Worker, args ...interface{}) {
	var prov = w.Props[0].(*provider.Provider)

	log.Printf("DEBUG: providerWorker[%s]: refresh", prov.Name)

	w.SendJobSignal(jobSignalRefresh)
}

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

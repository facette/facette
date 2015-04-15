package server

import (
	"fmt"
	"time"

	"github.com/facette/facette/pkg/config"
	"github.com/facette/facette/pkg/connector"
	"github.com/facette/facette/pkg/logger"
	"github.com/facette/facette/pkg/provider"
	"github.com/facette/facette/pkg/worker"
)

func (server *Server) startProviderWorkers() error {
	server.providerWorkers = worker.NewPool()

	logger.Log(logger.LevelDebug, "server", "declaring providers")

	for _, prov := range server.providers {
		connectorType, err := config.GetString(prov.Config.Connector, "type", true)
		if err != nil {
			return fmt.Errorf("provider `%s' connector: %s", prov.Name, err)
		} else if _, ok := connector.Connectors[connectorType]; !ok {
			return fmt.Errorf("provider `%s' uses unknown connector type `%s'", prov.Name, connectorType)
		}

		// Append server identifier to provider configuration
		prov.Config.Connector["_id"] = server.ID

		providerWorker := worker.NewWorker()
		providerWorker.RegisterEvent(eventInit, workerProviderInit)
		providerWorker.RegisterEvent(eventShutdown, workerProviderShutdown)
		providerWorker.RegisterEvent(eventRun, workerProviderRun)
		providerWorker.RegisterEvent(eventCatalogRefresh, workerProviderRefresh)

		if err := providerWorker.SendEvent(eventInit, false, prov, connectorType); err != nil {
			logger.Log(logger.LevelWarning, "server", "in provider `%s', %s", prov.Name, err)
			logger.Log(logger.LevelWarning, "server", "discarding provider `%s'", prov.Name)
			continue
		}

		// Add worker into pool if initialization went fine
		server.providerWorkers.Add(providerWorker)

		providerWorker.SendEvent(eventRun, true, nil)

		logger.Log(logger.LevelDebug, "server", "declared provider `%s'", prov.Name)
	}

	return nil
}

func (server *Server) stopProviderWorkers() {
	server.providerWorkers.Broadcast(eventShutdown, nil)

	// Wait for all workers to shut down
	server.providerWorkers.Wg.Wait()

	// Shut down providers filtering goroutine
	for _, prov := range server.providers {
		close(prov.Filters.Input)
	}
}

func workerProviderInit(w *worker.Worker, args ...interface{}) {
	var (
		prov          = args[0].(*provider.Provider)
		connectorType = args[1].(string)
	)

	logger.Log(logger.LevelDebug, "provider", "%s: init", prov.Name)

	// Instanciate the connector according to its type
	conn, err := connector.Connectors[connectorType](prov.Name, prov.Config.Connector)
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

	logger.Log(logger.LevelDebug, "provider", "%s: shutdown", prov.Name)

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

	logger.Log(logger.LevelDebug, "provider", "%s: starting", prov.Name)

	// If provider `refresh_interval` has been configured, set up a time ticker
	if prov.Config.RefreshInterval > 0 {
		timeTicker = time.NewTicker(time.Duration(prov.Config.RefreshInterval) * time.Second)
		timeChan = timeTicker.C
	}

	for {
		select {
		case _ = <-timeChan:
			logger.Log(logger.LevelDebug, "provider", "%s: performing refresh from connector", prov.Name)

			if err := prov.Connector.Refresh(prov.Name, prov.Filters.Input); err != nil {
				logger.Log(logger.LevelError, "provider", "%s: unable to refresh: %s", prov.Name, err)
				continue
			}

			prov.LastRefresh = time.Now()

		case cmd := <-w.ReceiveJobSignals():
			switch cmd {
			case jobSignalRefresh:
				logger.Log(logger.LevelInfo, "provider", "%s: received refresh command", prov.Name)

				if err := prov.Connector.Refresh(prov.Name, prov.Filters.Input); err != nil {
					logger.Log(logger.LevelError, "provider", "%s: unable to refresh: %s", prov.Name, err)
					continue
				}

				prov.LastRefresh = time.Now()

			case jobSignalShutdown:
				logger.Log(logger.LevelInfo, "provider", "%s: received shutdown command, stopping job", prov.Name)

				w.State = worker.JobStopped

				if timeTicker != nil {
					// Stop refresh time ticker
					timeTicker.Stop()
				}

				return

			default:
				logger.Log(logger.LevelNotice, "provider", "%s: received unknown command, ignoring", prov.Name)
			}
		}
	}
}

func workerProviderRefresh(w *worker.Worker, args ...interface{}) {
	var prov = w.Props[0].(*provider.Provider)

	logger.Log(logger.LevelDebug, "provider", "%s: refresh", prov.Name)

	w.SendJobSignal(jobSignalRefresh)
}

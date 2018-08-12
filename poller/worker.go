package poller

import (
	"time"

	"facette.io/facette/catalog"
	"facette.io/facette/connector"
	"facette.io/facette/storage"
	"facette.io/logger"
)

const (
	_ = iota
	workerCmdRefresh
	workerCmdShutdown
)

type worker struct {
	poller     *Poller
	provider   *storage.Provider
	logger     *logger.Logger
	connector  connector.Connector
	catalog    *catalog.Catalog
	filters    *catalog.FilterChain
	refreshing bool
	cmdChan    chan int
}

func newWorker(poller *Poller, provider *storage.Provider, logger *logger.Logger) (*worker, error) {
	// Initialize provider connector handler
	c, err := connector.NewConnector(provider.Connector, provider.Name, provider.Settings, logger)
	if err != nil {
		return nil, err
	}

	return &worker{
		poller:    poller,
		logger:    logger,
		provider:  provider,
		connector: c,
		catalog:   catalog.NewCatalog(provider.Name),
		filters:   catalog.NewFilterChain(&provider.Filters),
		cmdChan:   make(chan int),
	}, nil
}

func (w *worker) Run() {
	var ticker *time.Ticker

	w.poller.wg.Add(1)
	defer func() { w.poller.wg.Done() }()

	w.logger.Debug("started")

	// Register catalog into main searcher instance
	w.poller.searcher.Register(w.catalog)

	// Set catalog priority if defined
	if w.provider.Priority > 0 {
		w.logger.Debug("setting %q catalog priority to %d", w.provider.Name, w.provider.Priority)
		w.catalog.SetPriority(w.provider.Priority)
	}

	// Create new time ticker for automatic refresh
	f := func() {
		w.cmdChan <- workerCmdRefresh
	}
	go f()

	if w.provider.RefreshInterval > 0 {
		ticker = time.NewTicker(time.Duration(w.provider.RefreshInterval) * time.Second)

		go func() {
			for range ticker.C {
				f()
			}
		}()
	}

	for {
		select {
		case cmd := <-w.cmdChan:
			switch cmd {
			case workerCmdRefresh:
				w.logger.Debug("refreshing %q provider", w.provider.Name)

				go func() {
					w.refreshing = true
					err := w.connector.Refresh(w.filters.Input)
					if err != nil {
						w.logger.Error("provider %q encountered an error: %s", w.provider.Name, err)
					}
					w.refreshing = false
				}()

			case workerCmdShutdown:
				// Stop automatic refresh time ticker if any
				if ticker != nil {
					ticker.Stop()
				}

				goto stop
			}

		case record := <-w.filters.Output:
			// Append new metric into provider catalog
			w.logger.Debug("appending record %s in %q catalog", record, w.provider.Name)
			w.catalog.Insert(record)

		case msg := <-w.filters.Messages:
			w.logger.Debug("%s", msg)
		}
	}

stop:
	w.logger.Debug("stopped")
}

func (w *worker) Shutdown() {
	// Unregister catalog from main searcher instance
	w.poller.searcher.Unregister(w.catalog)

	w.cmdChan <- workerCmdShutdown
	close(w.cmdChan)
}

func (w *worker) Refresh() {
	if w == nil || w.refreshing {
		return
	}

	w.cmdChan <- workerCmdRefresh
}

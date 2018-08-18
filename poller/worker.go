package poller

import (
	"os"
	"path/filepath"
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
	c, err := connector.New(provider.Connector, provider.Name, provider.Settings, logger)
	if err != nil {
		return nil, err
	}

	return &worker{
		poller:    poller,
		logger:    logger,
		provider:  provider,
		connector: c,
		filters:   catalog.NewFilterChain(&provider.Filters),
		cmdChan:   make(chan int),
	}, nil
}

func (w *worker) Run() {
	var ticker *time.Ticker

	w.poller.wg.Add(1)
	defer func() { w.poller.wg.Done() }()

	w.logger.Debug("started")

	// Restore previous catalog state for a warm startup
	statePath := w.catalogDumpPath()

	_, err := os.Stat(statePath)
	if err == nil {
		start := time.Now()

		catalog := catalog.New(w.provider.Name, w.connector)
		if w.provider.Priority > 0 {
			catalog.Priority = w.provider.Priority
		}

		err = catalog.Restore(statePath)
		if err != nil {
			w.logger.Warning("failed to restore catalog state: %s", err)
		}

		w.catalog = catalog
		w.poller.searcher.Register(w.catalog)

		w.logger.Debug("restored previous catalog state in %s", time.Since(start))
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
					catalog := catalog.New(w.provider.Name, w.connector)
					if w.provider.Priority > 0 {
						catalog.Priority = w.provider.Priority
					}

					for record := range w.filters.Output {
						if record == nil {
							break
						}

						err := catalog.Insert(record)
						if err != nil {
							w.logger.Warning("failed to insert record %s to catalog: %s", record, err)
							continue
						}

						w.logger.Debug("inserted record %s in %q catalog", record, w.provider.Name)
					}

					// Register or replace catalog into searcher
					if w.catalog != nil {
						w.poller.searcher.Unregister(w.catalog)
					}
					w.catalog = catalog
					w.poller.searcher.Register(w.catalog)
				}()

				go func() {
					w.refreshing = true

					err := w.connector.Refresh(w.filters.Input)
					if err != nil {
						w.logger.Error("provider %q encountered an error: %s", w.provider.Name, err)
					}

					// Send nil record to stop processing
					w.filters.Input <- nil

					w.refreshing = false
				}()

			case workerCmdShutdown:
				// Stop automatic refresh time ticker if any
				if ticker != nil {
					ticker.Stop()
				}

				goto stop
			}

		case msg := <-w.filters.Messages:
			w.logger.Debug("%s", msg)
		}
	}

stop:
	w.logger.Debug("stopped")
}

func (w *worker) Shutdown() {
	if w.catalog != nil {
		// Unregister catalog from searcher instance
		w.poller.searcher.Unregister(w.catalog)

		// Dump current catalog state for future warm startup
		statePath := w.catalogDumpPath()

		stateDirPath := filepath.Dir(statePath)
		_, err := os.Stat(stateDirPath)
		if os.IsNotExist(err) {
			err = os.MkdirAll(stateDirPath, 0750)
			if err != nil {
				w.logger.Error("failed to create state directory: %s", err)
			}
		}

		if err == nil {
			err = w.catalog.Dump(statePath)
			if err != nil {
				w.logger.Warning("failed to dump catalog state: %s", err)
			}
		}
	}

	w.cmdChan <- workerCmdShutdown
	close(w.cmdChan)
}

func (w *worker) Refresh() {
	if w == nil || w.refreshing {
		return
	}

	w.cmdChan <- workerCmdRefresh
}

func (w *worker) catalogDumpPath() string {
	return filepath.Join(w.poller.config.Cache.Path, "state", w.provider.Name+".catalog")
}

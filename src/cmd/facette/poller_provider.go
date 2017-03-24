package main

import (
	"fmt"
	"sync"
	"time"

	"facette/backend"
	"facette/catalog"
	"facette/connector"
	"facette/worker"
)

const (
	_ = iota
	providerCmdRefresh
	providerCmdShutdown
)

type providerWorker struct {
	worker.CommonWorker

	poller     *pollerWorker
	provider   *backend.Provider
	connector  connector.Connector
	catalog    *catalog.Catalog
	filters    *catalog.FilterChain
	refreshing bool
	cmdChan    chan int
	wg         *sync.WaitGroup
}

func newProviderWorker(poller *pollerWorker, prov *backend.Provider) (*providerWorker, error) {
	// Initialize provider connector handler
	c, err := connector.NewConnector(prov.Connector, prov.Name, prov.Settings,
		poller.log.Context(fmt.Sprintf("poller[%s]", prov.Name)))
	if err != nil {
		return nil, err
	}

	return &providerWorker{
		poller:    poller,
		provider:  prov,
		connector: c,
		catalog:   catalog.NewCatalog(prov.Name),
		filters:   catalog.NewFilterChain(prov.Filters),
		cmdChan:   make(chan int),
		wg:        &sync.WaitGroup{},
	}, nil
}

func (w *providerWorker) Run(wg *sync.WaitGroup) {
	var (
		ticker   *time.Ticker
		timeChan <-chan time.Time
	)

	defer wg.Done()

	w.wg.Add(1)
	w.poller.log.Debug("provider %q started", w.provider.Name)

	// Register catalog into main searcher instance
	w.poller.service.searcher.Register(w.catalog)

	// Set catalog priority if defined
	if w.provider.Priority > 0 {
		w.poller.log.Debug("setting %q catalog priority to %d", w.provider.Name, w.provider.Priority)
		w.catalog.SetPriority(w.provider.Priority)
	}

	// Create new time ticker for automatic refresh
	refresh, _ := w.provider.Settings.GetInt("refresh_interval", 0)
	if refresh > 0 {
		ticker = time.NewTicker(time.Duration(refresh) * time.Second)
		timeChan = ticker.C
	}

	for {
		select {
		case _ = <-timeChan:
			// Trigger automatic provider refresh
			w.cmdChan <- providerCmdRefresh

		case cmd := <-w.cmdChan:
			switch cmd {
			case providerCmdRefresh:
				w.poller.log.Debug("refreshing %q provider", w.provider.Name)

				go func() {
					w.refreshing = true

					if err := w.connector.Refresh(w.filters.Input); err != nil {
						w.poller.log.Error("provider %q encountered an error: %s", w.provider.Name, err)
					}

					w.refreshing = false
				}()

			case providerCmdShutdown:
				// Stop automatic refresh time ticker if any
				if ticker != nil {
					ticker.Stop()
				}

				goto stop
			}

		case record := <-w.filters.Output:
			// Append new metric into provider catalog
			w.poller.log.Debug("appending record %s in %q catalog", record, w.provider.Name)
			w.catalog.Insert(record)

		case msg := <-w.filters.Messages:
			w.poller.log.Debug("%s", msg)
		}
	}

stop:
	w.poller.log.Debug("provider %q stopped", w.provider.Name)
	w.wg.Done()
}

func (w *providerWorker) Shutdown() {
	// Unregister catalog from main searcher instance
	w.poller.service.searcher.Unregister(w.catalog)

	// Trigger provider shutdown
	w.cmdChan <- providerCmdShutdown
	w.wg.Wait()
	close(w.cmdChan)

	w.CommonWorker.Shutdown()
}

func (w *providerWorker) Refresh() {
	if w.refreshing {
		fmt.Println("nope!")
		return
	}

	// Trigger provider refresh
	w.cmdChan <- providerCmdRefresh
}

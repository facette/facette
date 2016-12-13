package main

import (
	"sync"

	"facette/backend"
	"facette/worker"

	"github.com/facette/logger"
)

type pollerWorker struct {
	sync.Mutex
	worker.CommonWorker

	service   *Service
	log       *logger.Logger
	pool      *worker.Pool
	providers map[string]*providerWorker
	workers   map[string]*worker.Worker
}

func newPollerWorker(s *Service) *pollerWorker {
	return &pollerWorker{
		service:   s,
		log:       s.log.Context("poller"),
		pool:      worker.NewPool(),
		providers: make(map[string]*providerWorker),
		workers:   make(map[string]*worker.Worker),
	}
}

func (w *pollerWorker) Run(wg *sync.WaitGroup) {
	defer wg.Done()

	w.log.Debug("worker started")

	// Get providers list from backend
	providers := []backend.Provider{}

	if _, err := w.service.backend.List(&providers, map[string]interface{}{"enabled": true}, nil, 0, 0); err != nil {
		w.log.Error("failed to list providers: %s", err)
		return
	}

	// Start providers and apply catalog searcher priorities
	for _, p := range providers {
		w.StartProvider(p)
	}

	w.service.searcher.ApplyPriorities()
}

func (w *pollerWorker) Shutdown() {
	// Trigger providers shutdown
	w.pool.Shutdown()
	w.pool.Wait()

	// Shutdown worker
	w.log.Debug("worker stopped")
	w.CommonWorker.Shutdown()
}

func (w *pollerWorker) StartProvider(prov backend.Provider) {
	var err error

	w.Lock()
	defer w.Unlock()

	if !prov.Enabled {
		w.providers[prov.ID] = nil
		return
	}

	if _, ok := w.providers[prov.ID]; ok {
		w.log.Warning("provider %q is already registered", prov.ID)
		return
	}

	// Initialize new provider worker and perform initial refresh
	w.providers[prov.ID], err = newProviderWorker(w, &prov)
	if err != nil {
		w.log.Error("failed to start %q provider: %s", prov.Name, err)
		return
	}

	w.workers[prov.ID] = worker.NewWorker(w.providers[prov.ID])

	w.pool.AddAndRun(w.workers[prov.ID])
	w.providers[prov.ID].Refresh()
}

func (w *pollerWorker) StopProvider(prov backend.Provider, update bool) {
	w.Lock()

	if pw, ok := w.providers[prov.ID]; ok {
		// Stop running provider
		if pw != nil {
			(*pw).Shutdown()
		}
		delete(w.providers, prov.ID)
	}

	if pw, ok := w.workers[prov.ID]; ok {
		// Remove worker from pool
		w.pool.Remove(pw)
		delete(w.workers, prov.ID)
	}

	w.Unlock()

	// Try to restart provider instance if in update mode
	if update {
		w.StartProvider(prov)
	}
}

func (w *pollerWorker) Refresh() {
	for _, prov := range w.providers {
		go (*prov).Refresh()
	}
}

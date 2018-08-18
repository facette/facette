package poller

import (
	"context"
	"fmt"
	"sync"

	"facette.io/facette/catalog"
	"facette.io/facette/config"
	"facette.io/facette/storage"
	"facette.io/logger"
	"github.com/pkg/errors"
)

// Poller represents a poller instance.
type Poller struct {
	sync.Mutex

	ctx      context.Context
	storage  *storage.Storage
	searcher *catalog.Searcher
	config   *config.Config
	logger   *logger.Logger
	workers  map[string]*worker
	errors   map[string]error
	wg       *sync.WaitGroup
}

// New creates a new poller instance.
func New(
	ctx context.Context,
	storage *storage.Storage,
	searcher *catalog.Searcher,
	config *config.Config,
	logger *logger.Logger,
) *Poller {
	return &Poller{
		ctx:      ctx,
		storage:  storage,
		searcher: searcher,
		config:   config,
		logger:   logger,
		workers:  make(map[string]*worker),
		errors:   make(map[string]error),
		wg:       &sync.WaitGroup{},
	}
}

// Run starts polling the providers.
func (p *Poller) Run() error {
	var providers []*storage.Provider

	p.logger.Info("started")

	// Get providers list from storage
	_, err := p.storage.SQL().List(&providers, map[string]interface{}{"enabled": true}, nil, 0, 0, false)
	if err != nil {
		return errors.Wrap(err, "cannot list providers")
	}

	// Start providers and apply catalog searcher priorities
	for _, prov := range providers {
		p.StartWorker(prov)
	}
	p.searcher.ApplyPriorities()

	// Wait for main context cancellation
	<-p.ctx.Done()
	p.Shutdown()
	p.wg.Wait()

	p.logger.Info("stopped")

	return nil
}

// Shutdown stops the providers polling.
func (p *Poller) Shutdown() {
	for _, w := range p.workers {
		if w != nil {
			go p.StopWorker(w.provider, false)
		}
	}
}

// StartWorker starts a new poller worker given a storage provider.
func (p *Poller) StartWorker(prov *storage.Provider) {
	var err error

	p.Lock()
	defer p.Unlock()

	if !prov.Enabled {
		return
	}

	if _, ok := p.workers[prov.ID]; ok {
		p.logger.Warning("worker %q is already registered", prov.Name)
		return
	}

	// Initialize new poller worker and perform initial refresh
	p.workers[prov.ID], err = newWorker(p, prov, p.logger.Context(fmt.Sprintf("poller[%s]", prov.Name)))
	if err != nil {
		p.errors[prov.ID] = err
		p.logger.Error("failed to start %q worker: %s", prov.Name, err)
		return
	}

	p.errors[prov.ID] = nil
	go p.workers[prov.ID].Run()
}

// StopWorker stops an existing poller worker.
func (p *Poller) StopWorker(prov *storage.Provider, update bool) {
	p.Lock()
	if w, ok := p.workers[prov.ID]; ok {
		w.Shutdown()
		delete(p.workers, prov.ID)
	}
	p.Unlock()

	// Try to restart provider instance if in update mode
	if update {
		p.StartWorker(prov)
	}
}

// WorkerError returns the error returned on poller worker initialization.
func (p *Poller) WorkerError(id string) error {
	err, _ := p.errors[id]
	return err
}

// RefreshAll triggers a refresh on all the registered poller workers.
func (p *Poller) RefreshAll() {
	p.Lock()
	defer p.Unlock()

	for _, w := range p.workers {
		go w.Refresh()
	}
}

// Refresh triggers a refresh on an existing poller worker.
func (p *Poller) Refresh(prov storage.Provider) {
	p.Lock()
	defer p.Unlock()

	if w, ok := p.workers[prov.ID]; ok {
		go w.Refresh()
	}
}

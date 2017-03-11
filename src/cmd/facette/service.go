package main

import (
	"facette/backend"
	"facette/catalog"
	"facette/worker"
	"fmt"

	"github.com/facette/logger"
)

// Service represents a service struct.
type Service struct {
	config   *config
	log      *logger.Logger
	backend  *backend.Backend
	poller   *pollerWorker
	searcher *catalog.Searcher
	workers  *worker.Pool
	stopping bool
}

// NewService returns a new service instance.
func NewService(config *config) *Service {
	return &Service{
		config:   config,
		searcher: catalog.NewSearcher(),
		workers:  worker.NewPool(),
	}
}

// Run starts the service processing.
func (s *Service) Run() error {
	var err error

	// Initialize logger
	s.log, err = logger.NewLogger(logger.FileConfig{Level: s.config.LogLevel, Path: s.config.LogPath})
	if err != nil {
		return err
	}

	s.log.Info("service started")

	s.backend, err = backend.NewBackend(s.config.Backend, s.log.Context("backend"))
	if err != nil {
		s.log.Error("failed to initialize backend: %s", err)
		return nil
	}

	// Register and initialize workers
	s.poller = newPollerWorker(s)

	s.workers.Add(
		worker.NewWorker(newHTTPWorker(s)),
		worker.NewWorker(s.poller),
	)

	if err = s.workers.Init(); err != nil {
		return fmt.Errorf("failed to initialize workers: %s", err)
	}

	// Start workers and wait until they stop
	s.workers.Run()
	s.workers.Wait()

	s.log.Info("service stopped")

	return nil
}

// Shutdown stops the service.
func (s *Service) Shutdown() {
	if s.stopping {
		return
	}
	s.stopping = true

	s.log.Notice("received shutdown signal, stopping")

	// Close backend
	if s.backend != nil {
		s.backend.Close()
	}

	// Broadcast shutdown job to workers
	s.workers.Shutdown()
}

// Refresh broadcasts the refresh job to the service workers.
func (s *Service) Refresh() {
	s.log.Info("received refresh signal, broadcasting")

	// Trigger poller providers refresh
	s.poller.Refresh()
}

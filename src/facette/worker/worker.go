package worker

import "sync"

// Worker represents a worker instance.
type Worker interface {
	Init() error
	Run(*sync.WaitGroup)
	Shutdown()
	Stopping() bool
}

// NewWorker creates a new worker instance.
func NewWorker(worker interface{}) *Worker {
	w := worker.(Worker)
	return &w
}

// CommonWorker implements the common methods of a worker.
type CommonWorker struct {
	stopping bool
}

// Init initializes a worker.
func (w *CommonWorker) Init() error {
	return nil
}

// Shutdown stops a worker.
func (w *CommonWorker) Shutdown() {
	if w.stopping {
		return
	}
	w.stopping = true
}

// Stopping returns whether or not the worker is stopping.
func (w *CommonWorker) Stopping() bool {
	return w.stopping
}

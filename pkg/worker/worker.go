// Package worker provides background worker service.
package worker

import (
	"fmt"
	"sync"
)

const (
	_ = iota
	// JobStarted represents a started job state.
	JobStarted
	// JobStopped represents a stopped job state.
	JobStopped
)

type workerJob func(*Worker, ...interface{})

// Worker represents a worker instance structure.
type Worker struct {
	Props     []interface{}
	State     int
	events    map[int]workerJob
	eventChan chan workerEvent
	jobChan   chan int
	errorChan chan error
	wg        *sync.WaitGroup
}

// Pool represents a pool of worker instances.
type Pool struct {
	Workers []*Worker
	Wg      *sync.WaitGroup
}

type workerEvent struct {
	Type  int
	Async bool
	Args  []interface{}
}

// NewWorker instantiates a new worker.
func NewWorker() *Worker {
	worker := &Worker{
		State:     JobStopped,
		Props:     make([]interface{}, 0),
		events:    make(map[int]workerJob),
		eventChan: make(chan workerEvent),
		jobChan:   make(chan int),
		errorChan: make(chan error),
	}

	go func(worker *Worker) {
		for event := range worker.eventChan {
			if _, ok := worker.events[event.Type]; !ok {
				continue
			}

			go worker.events[event.Type](worker, event.Args...)
		}
	}(worker)

	return worker
}

// RegisterEvent registers a new event callback for the worker.
func (worker *Worker) RegisterEvent(event int, callback workerJob) error {
	if _, ok := worker.events[event]; ok {
		return fmt.Errorf("callback event already registered")
	}

	worker.events[event] = callback

	return nil
}

// SendEvent sends a event directly to the worker.
func (worker *Worker) SendEvent(event int, async bool, args ...interface{}) error {
	worker.eventChan <- workerEvent{
		Type:  event,
		Async: async,
		Args:  args,
	}

	if async {
		return nil
	}

	return <-worker.errorChan
}

// SendJobSignal sends a signal to a worker job.
func (worker *Worker) SendJobSignal(signal int) {
	worker.jobChan <- signal
}

// ReceiveJobSignals returns a channel receiving worker job signals.
func (worker *Worker) ReceiveJobSignals() chan int {
	return worker.jobChan
}

// ReturnErr returns an error to the sender of a synchronous event.
func (worker *Worker) ReturnErr(err error) {
	worker.errorChan <- err
}

// Shutdown shuts down the worker.
func (worker *Worker) Shutdown() {
	close(worker.eventChan)
	close(worker.jobChan)
	close(worker.errorChan)

	if worker.wg != nil {
		worker.wg.Done()
	}
}

// NewPool creates a new worker pool.
func NewPool() Pool {
	return Pool{
		Workers: make([]*Worker, 0),
		Wg:      &sync.WaitGroup{},
	}
}

// Add adds worker to the worker pool.
func (workerPool *Pool) Add(worker *Worker) {
	workerPool.Wg.Add(1)
	worker.wg = workerPool.Wg

	workerPool.Workers = append(workerPool.Workers, worker)
}

// Broadcast sends an event to all workers of the worker pool.
func (workerPool Pool) Broadcast(event int, args ...interface{}) {
	for _, worker := range workerPool.Workers {
		worker.SendEvent(event, true, args...)
	}
}

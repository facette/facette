package worker

import "sync"

// Pool represents a worker pool instance.
type Pool struct {
	workers []*Worker
	wg      *sync.WaitGroup
}

// NewPool creates a new worker pool instance.
func NewPool() *Pool {
	return &Pool{
		workers: []*Worker{},
		wg:      &sync.WaitGroup{},
	}
}

// Add appends a worker to the pool.
func (p *Pool) Add(workers ...*Worker) {
	for _, w := range workers {
		p.workers = append(p.workers, w)
		p.wg.Add(1)
	}
}

// AddAndRun appends a worker to the pool then runs it.
func (p *Pool) AddAndRun(workers ...*Worker) {
	for _, w := range workers {
		p.Add(w)
		go (*w).Run(p.wg)
	}
}

// Remove removes a worker from the pool.
func (p *Pool) Remove(w *Worker) {
	for i, worker := range p.workers {
		if worker == w {
			p.workers = append(p.workers[:i], p.workers[i+1:]...)
			break
		}
	}
}

// Init initializes the pool workers. It stops at the first encountered error.
func (p *Pool) Init() error {
	for _, w := range p.workers {
		if err := (*w).Init(); err != nil {
			return err
		}
	}

	return nil
}

// Run starts all the pool workers.
func (p *Pool) Run() {
	for _, w := range p.workers {
		go (*w).Run(p.wg)
	}
}

// Shutdown stops all the pool workers.
func (p *Pool) Shutdown() {
	for _, w := range p.workers {
		go (*w).Shutdown()
	}
}

// Wait waits for all all pool workers to terminate.
func (p *Pool) Wait() {
	p.wg.Wait()
}

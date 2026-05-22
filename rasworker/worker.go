// Package rasworker provides a generic worker pool for concurrent job execution
// with graceful shutdown and configurable error handling.
package rasworker

import (
	"context"
	"log/slog"
	"sync"
)

// Job is a function that performs work. It receives a context that is cancelled on shutdown.
type Job func(ctx context.Context) error

// ErrorHandler is a callback invoked when a job returns an error.
type ErrorHandler func(err error)

// Pool manages a pool of worker goroutines that process jobs from a queue.
type Pool struct {
	jobs          chan Job
	wg            sync.WaitGroup
	ctx           context.Context
	cancel        context.CancelFunc
	workers       int
	errorHandlers []ErrorHandler
	handlerMu     sync.RWMutex
}

// NewPool creates a new worker pool with the specified number of workers and job queue size.
// Call [Pool.Start] to begin processing jobs.
func NewPool(workers, queueSize int) *Pool {
	ctx, cancel := context.WithCancel(context.Background())
	return &Pool{
		jobs:          make(chan Job, queueSize),
		ctx:           ctx,
		cancel:        cancel,
		workers:       workers,
		errorHandlers: make([]ErrorHandler, 0),
	}
}

// NewPoolWithErrorHandler creates a new worker pool with an initial error handler.
// Additional handlers can be added with [Pool.AddErrorHandler].
func NewPoolWithErrorHandler(workers, queueSize int, onError ErrorHandler) *Pool {
	ctx, cancel := context.WithCancel(context.Background())

	return &Pool{
		jobs:          make(chan Job, queueSize),
		ctx:           ctx,
		cancel:        cancel,
		workers:       workers,
		errorHandlers: []ErrorHandler{onError},
	}
}

// AddErrorHandler adds an error handler that is called when a job returns an error.
// It is safe to call after [Pool.Start].
func (p *Pool) AddErrorHandler(onError ErrorHandler) {
	p.handlerMu.Lock()
	defer p.handlerMu.Unlock()
	p.errorHandlers = append(p.errorHandlers, onError)
}

// Start launches the worker goroutines. Jobs can be submitted via [Pool.Submit].
func (p *Pool) Start() {
	for i := range p.workers {
		p.wg.Add(1)
		go p.worker(i)
	}
	slog.Info("worker pool started", "workers", p.workers)
}

// worker runs in a goroutine and processes jobs from the queue.
func (p *Pool) worker(id int) {
	defer p.wg.Done()
	for {
		select {
		case <-p.ctx.Done():
			slog.Debug("worker shutting down", "workerId", id)
			return
		case job, ok := <-p.jobs:
			if !ok {
				return
			}
			if err := job(p.ctx); err != nil {
				p.handlerMu.RLock()
				for _, handler := range p.errorHandlers {
					if handler != nil {
						handler(err)
					}
				}
				p.handlerMu.RUnlock()
			}
		}
	}
}

// Submit adds a job to the queue for processing.
// It returns true if the job was queued, or false if the queue is full.
func (p *Pool) Submit(job Job) bool {
	select {
	case p.jobs <- job:
		return true
	default:
		slog.Warn("worker pool full, job dropped")
		return false
	}
}

// Shutdown gracefully stops the worker pool, waiting for queued jobs to complete.
// If the context is cancelled before all jobs finish, in-flight jobs are cancelled
// and the function returns the context error.
func (p *Pool) Shutdown(ctx context.Context) error {
	close(p.jobs)

	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		slog.Info("worker pool shutdown complete")
		return nil
	case <-ctx.Done():
		p.cancel()
		slog.Warn("worker pool shutdown timeout, canceling in-flight jobs")
		<-done
		return ctx.Err()
	}
}

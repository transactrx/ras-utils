package rasworker

import (
	"context"
	"log/slog"
	"sync"
)

type Job func(ctx context.Context) error

type Pool struct {
	jobs    chan Job
	wg      sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc
	workers int
}

func NewPool(workers, queueSize int) *Pool {
	ctx, cancel := context.WithCancel(context.Background())
	return &Pool{
		jobs:    make(chan Job, queueSize),
		ctx:     ctx,
		cancel:  cancel,
		workers: workers,
	}
}

func (p *Pool) Start() {
	for i := range p.workers {
		p.wg.Add(1)
		go p.worker(i)
	}
	slog.Info("worker pool started", "workers", p.workers)
}

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
				slog.Error("job failed", "workerId", id, "error", err)
			}
		}
	}
}

func (p *Pool) Submit(job Job) bool {
	select {
	case p.jobs <- job:
		return true
	default:
		slog.Warn("worker pool full, job dropped")
		return false
	}
}

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

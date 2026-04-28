package rasworker

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewPool(t *testing.T) {
	p := NewPool(5, 100)
	if p == nil {
		t.Fatal("NewPool returned nil")
	}
	if p.workers != 5 {
		t.Errorf("expected 5 workers, got %d", p.workers)
	}
	if cap(p.jobs) != 100 {
		t.Errorf("expected queue size 100, got %d", cap(p.jobs))
	}
}

func TestPool_StartAndShutdown(t *testing.T) {
	p := NewPool(3, 10)
	p.Start()
	if err := p.Shutdown(context.Background()); err != nil {
		t.Errorf("Shutdown returned error: %v", err)
	}
}

func TestPool_Submit_ExecutesJob(t *testing.T) {
	p := NewPool(2, 10)
	p.Start()

	var executed atomic.Bool
	done := make(chan struct{})

	ok := p.Submit(func(ctx context.Context) error {
		executed.Store(true)
		close(done)
		return nil
	})

	if !ok {
		t.Fatal("Submit returned false")
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("job not executed within timeout")
	}

	if !executed.Load() {
		t.Error("job was not executed")
	}

	p.Shutdown(context.Background())
}

func TestPool_Submit_MultipleJobs(t *testing.T) {
	p := NewPool(4, 100)
	p.Start()
	defer p.Shutdown(context.Background())

	var counter atomic.Int32
	var wg sync.WaitGroup
	jobCount := 50

	for i := 0; i < jobCount; i++ {
		wg.Add(1)
		ok := p.Submit(func(ctx context.Context) error {
			counter.Add(1)
			wg.Done()
			return nil
		})
		if !ok {
			t.Fatalf("Submit failed for job %d", i)
		}
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("not all jobs completed within timeout")
	}

	if counter.Load() != int32(jobCount) {
		t.Errorf("expected %d jobs executed, got %d", jobCount, counter.Load())
	}
}

func TestPool_Submit_QueueFull(t *testing.T) {
	p := NewPool(1, 1)

	blockingJob := make(chan struct{})
	p.jobs <- func(ctx context.Context) error {
		<-blockingJob
		return nil
	}

	ok := p.Submit(func(ctx context.Context) error {
		return nil
	})

	if ok {
		t.Error("Submit should return false when queue is full")
	}

	close(blockingJob)
}

func TestPool_JobError(t *testing.T) {
	p := NewPool(1, 10)
	p.Start()
	defer p.Shutdown(context.Background())

	done := make(chan struct{})
	expectedErr := errors.New("job error")

	p.Submit(func(ctx context.Context) error {
		defer close(done)
		return expectedErr
	})

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("job not executed within timeout")
	}
}

func TestPool_ContextCancellation(t *testing.T) {
	p := NewPool(2, 10)
	p.Start()

	var jobCtxCanceled atomic.Bool
	started := make(chan struct{})
	done := make(chan struct{})

	p.Submit(func(ctx context.Context) error {
		close(started)
		<-ctx.Done()
		jobCtxCanceled.Store(true)
		close(done)
		return ctx.Err()
	})

	<-started

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	err := p.Shutdown(ctx)

	if err != context.DeadlineExceeded {
		t.Errorf("expected DeadlineExceeded, got %v", err)
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("job context not canceled within timeout")
	}

	if !jobCtxCanceled.Load() {
		t.Error("job context should be canceled on shutdown timeout")
	}
}

func TestPool_ShutdownDrainsQueue(t *testing.T) {
	p := NewPool(2, 100)
	p.Start()

	var counter atomic.Int32
	jobCount := 20

	for range jobCount {
		p.Submit(func(ctx context.Context) error {
			time.Sleep(10 * time.Millisecond)
			counter.Add(1)
			return nil
		})
	}

	err := p.Shutdown(context.Background())
	if err != nil {
		t.Errorf("Shutdown returned error: %v", err)
	}

	if counter.Load() != int32(jobCount) {
		t.Errorf("expected %d jobs processed before shutdown, got %d", jobCount, counter.Load())
	}
}

func TestPool_ConcurrentSubmit(t *testing.T) {
	p := NewPool(4, 1000)
	p.Start()
	defer p.Shutdown(context.Background())

	var counter atomic.Int32
	var wg sync.WaitGroup
	submitters := 10
	jobsPerSubmitter := 100

	for s := 0; s < submitters; s++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < jobsPerSubmitter; j++ {
				p.Submit(func(ctx context.Context) error {
					counter.Add(1)
					return nil
				})
			}
		}()
	}

	wg.Wait()
	time.Sleep(500 * time.Millisecond)

	expected := int32(submitters * jobsPerSubmitter)
	if counter.Load() != expected {
		t.Errorf("expected %d jobs executed, got %d", expected, counter.Load())
	}
}

func TestNewPoolWithErrorHandler(t *testing.T) {
	var called atomic.Bool
	expectedErr := errors.New("test error")

	p := NewPoolWithErrorHandler(1, 10, func(err error) {
		if err == expectedErr {
			called.Store(true)
		}
	})
	p.Start()
	defer p.Shutdown(context.Background())

	done := make(chan struct{})
	p.Submit(func(ctx context.Context) error {
		defer close(done)
		return expectedErr
	})

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("job not executed within timeout")
	}

	time.Sleep(10 * time.Millisecond)
	if !called.Load() {
		t.Error("error handler was not called")
	}
}

func TestPool_AddErrorHandler(t *testing.T) {
	p := NewPool(1, 10)

	var handler1Called, handler2Called atomic.Bool
	expectedErr := errors.New("test error")

	p.AddErrorHandler(func(err error) {
		if err == expectedErr {
			handler1Called.Store(true)
		}
	})
	p.AddErrorHandler(func(err error) {
		if err == expectedErr {
			handler2Called.Store(true)
		}
	})

	p.Start()
	defer p.Shutdown(context.Background())

	done := make(chan struct{})
	p.Submit(func(ctx context.Context) error {
		defer close(done)
		return expectedErr
	})

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("job not executed within timeout")
	}

	time.Sleep(10 * time.Millisecond)
	if !handler1Called.Load() {
		t.Error("first error handler was not called")
	}
	if !handler2Called.Load() {
		t.Error("second error handler was not called")
	}
}

func TestPool_AddErrorHandler_Concurrent(t *testing.T) {
	p := NewPool(4, 100)
	p.Start()
	defer p.Shutdown(context.Background())

	var handlersAdded atomic.Int32
	var errorsHandled atomic.Int32

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			p.AddErrorHandler(func(err error) {
				errorsHandled.Add(1)
			})
			handlersAdded.Add(1)
		}()
	}
	wg.Wait()

	done := make(chan struct{})
	p.Submit(func(ctx context.Context) error {
		defer close(done)
		return errors.New("trigger handlers")
	})

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("job not executed within timeout")
	}

	time.Sleep(10 * time.Millisecond)
	if errorsHandled.Load() != handlersAdded.Load() {
		t.Errorf("expected %d handlers called, got %d", handlersAdded.Load(), errorsHandled.Load())
	}
}

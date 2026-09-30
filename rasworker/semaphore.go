package rasworker

import "context"

// Semaphore limits concurrent access to a resource.
// It is implemented as a buffered channel where the buffer size is the concurrency limit.
//
// Basic usage:
//
//	sem := rasworker.NewSemaphore(10)
//	sem.Acquire()
//	defer sem.Release()
//	// ... do work ...
type Semaphore struct {
	sem chan struct{}
}

// NewSemaphore creates a new [Semaphore] with the specified concurrency limit.
// The limit must be positive; a limit of zero creates a semaphore that blocks all acquires.
func NewSemaphore(limit int) *Semaphore {
	return &Semaphore{
		sem: make(chan struct{}, limit),
	}
}

// Acquire blocks until a slot is available, then acquires it.
// Use [Semaphore.Release] or [Semaphore.TryRelease] to free the slot when done.
func (s *Semaphore) Acquire() {
	s.sem <- struct{}{}
}

// AcquireContext blocks until a slot is available or the context is cancelled.
// It returns nil if the slot was acquired, or the context error if cancelled.
// If the context is cancelled, no slot is acquired and Release must not be called.
func (s *Semaphore) AcquireContext(ctx context.Context) error {
	select {
	case s.sem <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// TryAcquire attempts to acquire a slot without blocking.
// It returns true if the slot was acquired, false if no slots are available.
// If false is returned, no slot is acquired and Release must not be called.
func (s *Semaphore) TryAcquire() bool {
	select {
	case s.sem <- struct{}{}:
		return true
	default:
		return false
	}
}

// Release frees a slot previously acquired by [Semaphore.Acquire], [Semaphore.AcquireContext],
// or [Semaphore.TryAcquire]. It blocks indefinitely if called without a matching acquire.
// For non-blocking release with error detection, use [Semaphore.TryRelease].
func (s *Semaphore) Release() {
	<-s.sem
}

// TryRelease attempts to free a slot without blocking.
// It returns true if a slot was released, false if no slots were held.
// Use this for defensive programming or when misuse detection is needed.
func (s *Semaphore) TryRelease() bool {
	select {
	case <-s.sem:
		return true
	default:
		return false
	}
}

// Limit returns the maximum number of concurrent acquisitions allowed.
func (s *Semaphore) Limit() int {
	return cap(s.sem)
}

// Available returns the number of slots currently available.
// This value is informational and may change immediately after being read;
// do not use it for synchronization decisions.
func (s *Semaphore) Available() int {
	return cap(s.sem) - len(s.sem)
}

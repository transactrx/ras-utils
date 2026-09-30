package rasworker

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestSemaphore_Acquire(t *testing.T) {
	sem := NewSemaphore(2)

	// Should acquire immediately
	sem.Acquire()
	sem.Acquire()

	if sem.Available() != 0 {
		t.Errorf("expected 0 available, got %d", sem.Available())
	}

	// Release one
	sem.Release()
	if sem.Available() != 1 {
		t.Errorf("expected 1 available, got %d", sem.Available())
	}
}

func TestSemaphore_TryAcquire(t *testing.T) {
	sem := NewSemaphore(1)

	if !sem.TryAcquire() {
		t.Error("first TryAcquire should succeed")
	}

	if sem.TryAcquire() {
		t.Error("second TryAcquire should fail when full")
	}

	sem.Release()

	if !sem.TryAcquire() {
		t.Error("TryAcquire should succeed after release")
	}
}

func TestSemaphore_AcquireContext(t *testing.T) {
	sem := NewSemaphore(1)
	sem.Acquire()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := sem.AcquireContext(ctx)
	if err != context.DeadlineExceeded {
		t.Errorf("expected DeadlineExceeded, got %v", err)
	}

	// Release and try again
	sem.Release()
	ctx2, cancel2 := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel2()

	err = sem.AcquireContext(ctx2)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestSemaphore_ConcurrentAccess(t *testing.T) {
	sem := NewSemaphore(3)
	var maxConcurrent atomic.Int32
	var current atomic.Int32
	var wg sync.WaitGroup

	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem.Acquire()
			defer sem.Release()

			c := current.Add(1)
			if c > maxConcurrent.Load() {
				maxConcurrent.Store(c)
			}
			time.Sleep(10 * time.Millisecond)
			current.Add(-1)
		}()
	}

	wg.Wait()

	if maxConcurrent.Load() > 3 {
		t.Errorf("max concurrent exceeded limit: got %d, want <= 3", maxConcurrent.Load())
	}
}

func TestSemaphore_Limit(t *testing.T) {
	sem := NewSemaphore(5)
	if sem.Limit() != 5 {
		t.Errorf("expected limit 5, got %d", sem.Limit())
	}
}

func TestSemaphore_Available(t *testing.T) {
	sem := NewSemaphore(3)

	if sem.Available() != 3 {
		t.Errorf("expected 3 available, got %d", sem.Available())
	}

	sem.Acquire()
	if sem.Available() != 2 {
		t.Errorf("expected 2 available, got %d", sem.Available())
	}

	sem.Acquire()
	sem.Acquire()
	if sem.Available() != 0 {
		t.Errorf("expected 0 available, got %d", sem.Available())
	}
}

func TestSemaphore_TryRelease(t *testing.T) {
	sem := NewSemaphore(2)

	// TryRelease on empty semaphore should return false
	if sem.TryRelease() {
		t.Error("TryRelease should return false when no slots are held")
	}

	// Acquire and then TryRelease should succeed
	sem.Acquire()
	if !sem.TryRelease() {
		t.Error("TryRelease should return true after Acquire")
	}

	// TryRelease again should fail
	if sem.TryRelease() {
		t.Error("TryRelease should return false after slot already released")
	}

	// Multiple acquires and releases
	sem.Acquire()
	sem.Acquire()
	if !sem.TryRelease() {
		t.Error("first TryRelease should succeed")
	}
	if !sem.TryRelease() {
		t.Error("second TryRelease should succeed")
	}
	if sem.TryRelease() {
		t.Error("third TryRelease should fail")
	}
}

func TestNewSemaphore_PanicsOnZeroLimit(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for zero limit")
		}
	}()
	NewSemaphore(0)
}

func TestNewSemaphore_PanicsOnNegativeLimit(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for negative limit")
		}
	}()
	NewSemaphore(-1)
}

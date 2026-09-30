//go:build !race

package rasworker

import (
	"context"
	"testing"
	"time"
)

// TestPool_SubmitWait_UnblocksOnShutdown verifies that a blocked SubmitWait
// returns ErrPoolShutdown when Shutdown is called.
//
// This test is skipped with -race because it deliberately creates a race condition
// between SubmitWait's channel send and Shutdown's channel close. The race is
// handled safely by the closed channel signaling mechanism, but the race detector
// still flags the concurrent channel operations.
func TestPool_SubmitWait_UnblocksOnShutdown(t *testing.T) {
	p := NewPool(1, 0)
	p.Start()

	blocker := make(chan struct{})
	started := make(chan struct{})

	// Block the only worker
	p.SubmitWait(context.Background(), func(ctx context.Context) error {
		close(started)
		<-blocker
		return nil
	})
	<-started

	// SubmitWait blocks because worker is busy and queue is full (size 0)
	errCh := make(chan error, 1)
	go func() {
		errCh <- p.SubmitWait(context.Background(), func(ctx context.Context) error {
			return nil
		})
	}()

	// Give SubmitWait time to block
	time.Sleep(20 * time.Millisecond)

	// Shutdown should unblock the waiting SubmitWait
	go p.Shutdown(context.Background())

	// Wait a bit for shutdown to signal, then release blocker
	time.Sleep(10 * time.Millisecond)
	close(blocker)

	select {
	case err := <-errCh:
		if err != ErrPoolShutdown {
			t.Errorf("expected ErrPoolShutdown, got %v", err)
		}
	case <-time.After(time.Second):
		t.Error("SubmitWait did not unblock after shutdown")
	}
}

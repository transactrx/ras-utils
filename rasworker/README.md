# rasworker

Generic worker pool with configurable concurrency and graceful shutdown.

## Features

- Fixed-size worker pool
- Channel-based job queue
- Graceful shutdown with queue draining
- Context cancellation support

## Installation

```go
import "github.com/transactrx/ras-utils/rasworker"
```

## Usage

### Basic Worker Pool

```go
// Create pool with 10 workers and queue size of 100
pool := rasworker.NewPool(10, 100)

// Start the workers
pool.Start()

// Submit work
pool.Submit(func() {
    // do work
})

// Graceful shutdown (waits for queued work to complete)
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
pool.Shutdown(ctx)
```

### With Context Cancellation

```go
ctx, cancel := context.WithCancel(context.Background())
pool := rasworker.NewPoolWithContext(ctx, 5, 50)
pool.Start()

// Submit work that respects context
pool.Submit(func() {
    select {
    case <-ctx.Done():
        return  // cancelled
    default:
        // do work
    }
})

// Cancel all work
cancel()
```

### Non-blocking Submit

```go
// Returns false if queue is full
if !pool.TrySubmit(func() { /* work */ }) {
    log.Println("queue full, work dropped")
}
```

### Monitoring

```go
// Current queue depth
pending := pool.QueueSize()

// Check if pool is running
if pool.IsRunning() {
    // ...
}
```

## API Reference

### Types

- `Pool` - Worker pool manager

### Constructors

- `NewPool(workers, queueSize int) *Pool` - Create pool
- `NewPoolWithContext(ctx context.Context, workers, queueSize int) *Pool` - Create pool with context

### Methods

- `Start()` - Start worker goroutines
- `Submit(fn func())` - Queue work (blocks if queue full)
- `TrySubmit(fn func()) bool` - Queue work (returns false if queue full)
- `Shutdown(ctx context.Context) error` - Graceful shutdown
- `QueueSize() int` - Current pending work count
- `IsRunning() bool` - Check if pool is active

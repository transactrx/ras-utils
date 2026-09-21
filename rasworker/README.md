# rasworker

Generic worker pool with configurable concurrency and graceful shutdown.

## Features

- Fixed-size worker pool
- Channel-based job queue
- Graceful shutdown with queue draining
- Context cancellation support
- Configurable error handling

## Installation

```go
import "github.com/transactrx/ras-utils/rasworker"
```

## Usage

### Basic Worker Pool

```go
// Create pool with 10 workers and queue size of 100
pool := rasworker.NewPool(10, 100)
pool.Start()

// Submit work (jobs receive a context for cancellation)
pool.Submit(func(ctx context.Context) error {
    // do work
    return nil
})

// Graceful shutdown (waits for queued work to complete)
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
pool.Shutdown(ctx)
```

### With Parent Context

Jobs automatically cancel when the parent context is cancelled:

```go
appCtx, appCancel := context.WithCancel(context.Background())

pool := rasworker.NewPool(5, 50, rasworker.WithContext(appCtx))
pool.Start()

pool.Submit(func(ctx context.Context) error {
    select {
    case <-ctx.Done():
        return ctx.Err() // cancelled
    default:
        // do work
        return nil
    }
})

// Cancelling appCtx stops all jobs
appCancel()
```

### With Error Handler

```go
pool := rasworker.NewPool(10, 100,
    rasworker.WithErrorHandler(func(err error) {
        slog.Error("job failed", "error", err)
    }),
)
pool.Start()

pool.Submit(func(ctx context.Context) error {
    return errors.New("something went wrong") // triggers error handler
})
```

### Combined Options

```go
pool := rasworker.NewPool(10, 100,
    rasworker.WithContext(appCtx),
    rasworker.WithErrorHandler(logError),
    rasworker.WithErrorHandler(metrics.RecordError), // multiple handlers OK
)
```

### Adding Error Handlers After Start

```go
pool := rasworker.NewPool(10, 100)
pool.Start()

// Safe to add handlers while running
pool.AddErrorHandler(func(err error) {
    slog.Warn("job error", "error", err)
})
```

### Non-blocking Submit

```go
// Returns false if queue is full (job is dropped)
if !pool.Submit(func(ctx context.Context) error { return nil }) {
    log.Println("queue full, work dropped")
}
```

## API Reference

### Types

- `Pool` - Worker pool manager
- `Job` - `func(ctx context.Context) error`
- `ErrorHandler` - `func(err error)`
- `PoolOption` - Functional option for configuration

### Constructors

- `NewPool(workers, queueSize int, opts ...PoolOption) *Pool` - Create pool with options

### Options

- `WithContext(ctx context.Context)` - Set parent context for cancellation propagation
- `WithErrorHandler(fn ErrorHandler)` - Add error handler (can specify multiple)

### Methods

- `Start()` - Start worker goroutines
- `Submit(job Job) bool` - Queue work, returns false if queue full
- `AddErrorHandler(fn ErrorHandler)` - Add error handler (safe to call after Start)
- `Shutdown(ctx context.Context) error` - Graceful shutdown; returns context error if timeout exceeded

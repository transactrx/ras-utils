# rasworker

Generic worker pool and concurrency primitives with graceful shutdown.

## Features

- Fixed-size worker pool with channel-based job queue
- Semaphore for bounding concurrent operations
- Graceful shutdown with queue draining
- Context cancellation support
- Configurable error handling

## Installation

```go
import "github.com/transactrx/ras-utils/rasworker"
```

## Worker Pool

### Basic Usage

```go
// Create pool with 10 workers and queue size of 100
pool := rasworker.NewPool(10, 100)
pool.Start()

// Submit work (non-blocking, drops if queue full)
pool.Submit(func(ctx context.Context) error {
    // do work
    return nil
})

// Graceful shutdown (waits for queued work to complete)
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
pool.Shutdown(ctx)
```

### Blocking Submit

Use `SubmitWait` to block until the job is queued instead of dropping:

```go
// Blocks until job is queued or context cancelled
err := pool.SubmitWait(ctx, func(ctx context.Context) error {
    return doWork()
})
if err != nil {
    // context was cancelled before job could be queued
}
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

## Semaphore

Semaphore limits concurrent access to a resource. Use it to bound goroutines or rate-limit external API calls.

### Basic Usage

```go
// Limit to 10 concurrent operations
sem := rasworker.NewSemaphore(10)

sem.Acquire()        // blocks until slot available
defer sem.Release()  // always release with defer
doWork()
```

### With Context Timeout

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

err := sem.AcquireContext(ctx)
if err != nil {
    // timed out waiting for slot
    return err
}
defer sem.Release()
doWork()
```

### Non-blocking Acquire

```go
if sem.TryAcquire() {
    defer sem.Release()
    doWork()
} else {
    // no slots available, handle gracefully
}
```

### Defensive Release

Use `TryRelease` to detect misuse (release without acquire):

```go
sem.Acquire()
defer func() {
    if !sem.TryRelease() {
        slog.Error("semaphore released without matching acquire")
    }
}()
doWork()
```

### Bounding External API Calls

```go
var apiSem = rasworker.NewSemaphore(10)

func CallExternalAPI(ctx context.Context, req Request) (*Response, error) {
    apiSem.Acquire()
    defer apiSem.Release()
    
    return httpClient.Do(ctx, req)
}
```

## API Reference

### Pool

| Function | Description |
|----------|-------------|
| `NewPool(workers, queueSize int, opts ...PoolOption) *Pool` | Create pool with options |
| `WithContext(ctx context.Context) PoolOption` | Set parent context for cancellation |
| `WithErrorHandler(fn ErrorHandler) PoolOption` | Add error handler (can specify multiple) |

| Method | Description |
|--------|-------------|
| `Start()` | Start worker goroutines |
| `Submit(job Job) bool` | Queue work; returns false if queue full (job dropped) |
| `SubmitWait(ctx context.Context, job Job) error` | Queue work; blocks until queued or context cancelled |
| `AddErrorHandler(fn ErrorHandler)` | Add error handler (safe to call after Start) |
| `Shutdown(ctx context.Context) error` | Graceful shutdown; returns context error if timeout exceeded |

### Semaphore

| Function | Description |
|----------|-------------|
| `NewSemaphore(limit int) *Semaphore` | Create semaphore with concurrency limit |

| Method | Description |
|--------|-------------|
| `Acquire()` | Block until slot available |
| `AcquireContext(ctx context.Context) error` | Block until slot available or context cancelled |
| `TryAcquire() bool` | Non-blocking acquire; returns false if no slots |
| `Release()` | Free a slot; blocks indefinitely if called without acquire |
| `TryRelease() bool` | Non-blocking release; returns false if no slots held |
| `Limit() int` | Return max concurrent acquisitions |
| `Available() int` | Return currently available slots (informational only) |

### Types

```go
type Job func(ctx context.Context) error
type ErrorHandler func(err error)
```

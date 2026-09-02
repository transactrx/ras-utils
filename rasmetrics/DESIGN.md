# rasmetrics Design Spec

## Purpose

Reusable Prometheus metrics module for TransactRx Go services. Provides consistent observability across clinicalPlus, notificationEngine, rasUrlShortener, and other services with minimal boilerplate.

## Requirements

- Go 1.27+ (uses goroutine leak profiling and traceback labels)

## Goals

1. **Zero-config defaults** — runtime metrics (goroutines, memory, GC) out of the box
2. **Worker pool tracking** — Inc/Dec pattern for monitoring active goroutines per pool
3. **HTTP integration** — `/metrics` handler + optional request metrics middleware
4. **Consistent naming** — namespace prefix ensures metrics don't collide across services
5. **Debuggability** — leverage Go 1.27 pprof labels and goroutine leak detection

## Non-Goals

- Distributed tracing (separate concern)
- Log aggregation (use raslogging)
- Custom metric types beyond gauges/counters/histograms

## Dependencies

```
github.com/prometheus/client_golang v1.19+
```

## API Design

### Initialization

```go
package rasmetrics

// Init registers default collectors and applies options.
// Call once at startup before creating worker pools.
func Init(opts ...Option) error

// Options
func WithNamespace(ns string) Option       // e.g., "clinicalplus"
func WithRuntimeMetrics() Option           // goroutines, memory, GC (default: enabled)
func WithHTTPMetrics() Option              // request count/duration/size
func WithPprof() Option                    // expose /debug/pprof/* including goroutineleak (Go 1.27+)
func WithCustomRegistry(r *prometheus.Registry) Option  // for testing
```

### Worker Pool Tracking

```go
// WorkerPool tracks active workers for a named pool.
// Thread-safe for concurrent Inc/Dec calls.
// Automatically applies pprof labels for traceback visibility (Go 1.27+).
type WorkerPool struct {
    name  string
    gauge prometheus.Gauge
}

// NewWorkerPool creates a gauge: {namespace}_workers_active{pool="name"}
func NewWorkerPool(name string) *WorkerPool

func (p *WorkerPool) Inc()           // worker started
func (p *WorkerPool) Dec()           // worker finished
func (p *WorkerPool) Set(n int)      // set absolute count
func (p *WorkerPool) Value() float64 // current count (for testing)

// Do runs fn with pprof labels set for this pool.
// Leaked goroutines will show pool name in /debug/pprof/goroutineleak.
// Usage: pool.Do(ctx, func(ctx context.Context) { ... })
func (p *WorkerPool) Do(ctx context.Context, fn func(context.Context))
```

### HTTP

```go
// Handler returns the Prometheus metrics handler for /metrics endpoint.
func Handler() http.Handler

// PprofHandler returns a mux with /debug/pprof/* endpoints registered.
// Includes Go 1.27's /debug/pprof/goroutineleak for detecting leaked goroutines.
func PprofHandler() http.Handler

// Middleware wraps an http.Handler to record request metrics.
// Records: request_count, request_duration_seconds, request_size_bytes
func Middleware(next http.Handler) http.Handler
```

## Metrics Exposed

### Runtime (when WithRuntimeMetrics enabled)

| Metric | Type | Description |
|--------|------|-------------|
| `{ns}_runtime_goroutines` | Gauge | `runtime.NumGoroutine()` |
| `{ns}_runtime_alloc_bytes` | Gauge | Heap bytes allocated |
| `{ns}_runtime_gc_pause_seconds` | Summary | GC pause durations |

### Workers

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `{ns}_workers_active` | Gauge | `pool` | Active workers per pool |

### HTTP (when WithHTTPMetrics + Middleware used)

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `{ns}_http_requests_total` | Counter | `method`, `path`, `status` | Request count |
| `{ns}_http_request_duration_seconds` | Histogram | `method`, `path` | Latency distribution |

### Pprof Endpoints (when WithPprof enabled)

| Endpoint | Description |
|----------|-------------|
| `/debug/pprof/` | Index of available profiles |
| `/debug/pprof/goroutine` | Stack traces of all goroutines |
| `/debug/pprof/goroutineleak` | **Go 1.27+** Goroutines blocked on unreachable primitives |
| `/debug/pprof/heap` | Heap memory profile |
| `/debug/pprof/profile` | CPU profile (30s default) |
| `/debug/pprof/trace` | Execution trace |

The `goroutineleak` profile detects goroutines blocked on channels, mutexes, or conds that can never be unblocked (primitive is unreachable from any runnable goroutine). Useful for finding worker pool leaks in long-running services.

## Usage Example

```go
package main

import (
    "context"
    "net/http"
    "github.com/transactrx/ras-utils/rasmetrics"
)

func main() {
    // Initialize with namespace
    rasmetrics.Init(
        rasmetrics.WithNamespace("clinicalplus"),
        rasmetrics.WithRuntimeMetrics(),
        rasmetrics.WithHTTPMetrics(),
        rasmetrics.WithPprof(),  // enables /debug/pprof/* including goroutineleak
    )

    // Create worker pool trackers
    kafkaPool := rasmetrics.NewWorkerPool("kafka_consumers")
    natsPool := rasmetrics.NewWorkerPool("nats_publishers")

    // Option 1: Manual Inc/Dec
    go func() {
        kafkaPool.Inc()
        defer kafkaPool.Dec()
        // ... do work
    }()

    // Option 2: Use Do() for automatic pprof labels (recommended)
    // Leaked goroutines will show "pool=kafka_consumers" in traces
    go func() {
        kafkaPool.Do(context.Background(), func(ctx context.Context) {
            // ... do work
            // if this goroutine leaks, /debug/pprof/goroutineleak shows pool name
        })
    }()

    // HTTP setup
    mux := http.NewServeMux()
    mux.Handle("/metrics", rasmetrics.Handler())
    mux.Handle("/debug/pprof/", rasmetrics.PprofHandler())
    mux.Handle("/api/", rasmetrics.Middleware(apiHandler))

    http.ListenAndServe(":8080", mux)
}
```

## File Structure

```
rasmetrics/
├── DESIGN.md       # this file
├── README.md       # usage docs (post-implementation)
├── metrics.go      # Init, options, runtime collectors
├── metrics_test.go
├── workers.go      # WorkerPool type with pprof.Do integration
├── workers_test.go
├── http.go         # Handler, Middleware
├── http_test.go
├── pprof.go        # PprofHandler wrapping net/http/pprof
└── pprof_test.go
```

## Grafana Integration

Example queries:

```promql
# Active Kafka workers
clinicalplus_workers_active{pool="kafka_consumers"}

# Goroutine growth rate
rate(clinicalplus_runtime_goroutines[5m])

# Request latency p99
histogram_quantile(0.99, rate(clinicalplus_http_request_duration_seconds_bucket[5m]))
```

## Debugging Goroutine Leaks

When Grafana shows goroutine count climbing:

```bash
# Fetch leak profile from running service
curl http://localhost:8080/debug/pprof/goroutineleak > leak.prof

# View in browser
go tool pprof -http=:9090 leak.prof
```

Leaked goroutines will show their `pool` label in the traceback header (Go 1.27+), making it easy to identify which worker pool is leaking.

## Open Questions

1. **Should we support push gateway?** — Some ECS tasks are short-lived batch jobs that might not get scraped. Defer until needed.
2. **Default labels?** — Should we auto-add `service`, `version`, `env` labels from rasconfig? Adds coupling but useful for filtering.

## Implementation Plan

1. `metrics.go` — Init, options, runtime metrics
2. `workers.go` — WorkerPool gauge wrapper with pprof.Do integration
3. `http.go` — Handler and Middleware
4. `pprof.go` — PprofHandler exposing Go 1.27 endpoints
5. Tests for each file
6. README.md with examples
7. Integration test in one service (rasUrlShortener recommended — smallest footprint)

# rasevents

Event publishing via NATS with sync/async support, worker pools, graceful shutdown, and observability hooks.

## Features

- Synchronous and asynchronous event publishing
- Worker pool for async event processing
- Graceful shutdown with queue draining
- Observability hooks for metrics/monitoring
- Global and instance-based APIs

## Installation

```go
import "github.com/transactrx/ras-utils/rasevents"
```

## Environment Variables

### Required

| Variable | Description |
|----------|-------------|
| `NATS_URL` | NATS server URL (e.g., `nats://localhost:4222`) |
| `NATS_QUEUE_NAME` | Queue group name for load balancing (service only) |

### Optional

| Variable | Description |
|----------|-------------|
| `NATS_JWT` | JWT token for authenticated connections |
| `NATS_KEY` | Private key for authenticated connections |
| `NATS_DEBUG` | Enable debug logging (`true`/`false`) |
| `APPID` | Application identifier for connection naming |
| `MAX_SIZE_BEFORE_COMPRESS` | Client compression threshold (default: 2KB) |
| `MAX_SIZE_BEFORE_CHUNK` | Client chunking threshold (default: 8KB) |

### Config-based

| Variable | Description |
|----------|-------------|
| `EVENTS_DEFAULT_NAMESPACE` | Default namespace (required) |
| `EVENTS_SUBJECT` | Base NATS subject (required) |
| `EVENTS_TIMEOUT_SECONDS` | Request timeout in seconds (default: 60) |
| `EVENTS_WORKER_POOL_SIZE` | Async worker count (default: 50) |
| `EVENTS_QUEUE_SIZE` | Async queue size (default: 1000) |

## Usage

### Global Functions

```go
rasevents.Init(&rasevents.Config{
    DefaultNamespace: "MyService",
    Subject:          "custom.events.subject",
    Timeout:          30 * time.Second,
    WorkerPoolSize:   20,
    EventQueueSize:   500,
})

err := rasevents.SendEvent("PatientNotification", "Email", payload)
queued := rasevents.SendEventAsync("PatientNotification", "SMS", payload)

// Graceful shutdown (drains queue before stopping)
defer rasevents.Shutdown(context.Background())
```

### Instance-based

```go
handler := rasevents.NewEventsHandler(rasevents.Config{
    DefaultNamespace: "MyService",
    Subject:          "custom.events.subject",
    Timeout:          10 * time.Second,
    WorkerPoolSize:   5,
    EventQueueSize:   100,
}, nil) // nil client = create lazily

err := handler.SendEvent("Namespace", "EventType", payload)
queued := handler.SendEventAsync("Namespace", "EventType", payload)

// Shutdown with timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
if err := handler.Shutdown(ctx); err != nil {
    log.Printf("Shutdown interrupted: %v", err)
}
```

### Observability Hooks

```go
handler := rasevents.NewEventsHandler(rasevents.Config{
    // ... config ...
    Hooks: &rasevents.Hooks{
        // Called after each synchronous send
        OnEventSent: func(namespace, eventType string, duration time.Duration, err error) {
            metrics.RecordLatency("event_send", duration)
            if err != nil {
                metrics.IncrCounter("event_send_errors")
            }
        },
        // Called when async event is queued (dropped=true if queue full)
        OnEventQueued: func(namespace, eventType string, dropped bool) {
            if dropped {
                metrics.IncrCounter("event_dropped")
            }
        },
        // Called after async worker processes an event
        OnEventProcessed: func(namespace, eventType string, duration time.Duration, err error) {
            metrics.RecordLatency("event_process", duration)
        },
    },
}, nil)
```

### Testing

```go
// Inject mock client for testing
rasevents.SetNatsClient(mockClient)
defer rasevents.ResetNatsClient()

// Or with handler instances
handler := rasevents.NewEventsHandler(cfg, mockClient)
```

## API Reference

### Types

- `Config` - Handler configuration
- `Hooks` - Observability callbacks
- `EventsHandler` - Event publishing handler

### Global Functions

- `Init(cfg *Config)` - Initialize global handler
- `SendEvent(namespace, eventType string, payload any) error` - Synchronous send
- `SendEventAsync(namespace, eventType string, payload any) bool` - Async send
- `Shutdown(ctx context.Context) error` - Graceful shutdown
- `SetNatsClient(client NatsClient)` - Inject mock client
- `ResetNatsClient()` - Reset to default client

### Handler Methods

- `SendEvent(namespace, eventType string, payload any) error`
- `SendEventAsync(namespace, eventType string, payload any) bool`
- `Shutdown(ctx context.Context) error`

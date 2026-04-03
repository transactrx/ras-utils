# ras-utils

Go utility library providing shared helper functions for the Clinical+ ecosystem.

## Installation

```bash
go get github.com/transactrx/ras-utils
```

## Packages

### rascache

Generic in-memory key-value cache with TTL expiration and thread-safe operations.

```go
import "github.com/transactrx/ras-utils/rascache"

// Create a cache (expired items removed on access)
c := rascache.NewCache[string, User]()

// Create a cache with background cleanup (removes expired items periodically)
c := rascache.NewCacheWithCleanup[string, User](5 * time.Minute)
defer c.Stop() // stop the cleanup goroutine when done

// Set with TTL
c.Set("user:123", user, 5*time.Minute)

// Get (returns zero value and false if expired/missing)
user, ok := c.Get("user:123")

// Cache-through pattern: fetch from source if not cached
user, ok := c.TryGet("user:123", func() (rascache.CacheItem[User], bool) {
    user, err := db.GetUser(123)
    if err != nil {
        return rascache.CacheItem[User]{}, false
    }
    return rascache.NewCacheItem(user, time.Now().Add(5*time.Minute)), true
})

// Delete and Clear
c.Delete("user:123")
c.Clear()
```

### rasconfig

Database configuration and environment variable helpers.

```go
import "github.com/transactrx/ras-utils/rasconfig"

// Environment variables with defaults
host := rasconfig.GetEnvironmentVariableOrDefault("DB_HOST", "localhost")
port := rasconfig.GetEnvironmentVariableOrDefaultInt("DB_PORT", 5432)
timeout := rasconfig.GetEnvironmentVariableOrDefaultDuration("DB_TIMEOUT", "30s")

// Required environment variables (panics if missing)
apiKey := rasconfig.GetEnvironmentVariableOrPanic("API_KEY", "API_KEY is required")

// Database connection pool
cfg := &rasconfig.DBConfig{
    Host:                  "localhost",
    ReadOnlyHost:          "readonly.localhost",
    Port:                  "5432",
    DatabaseName:          "mydb",
    User:                  "user",
    Password:              "pass",
    MaxConnections:        10,
    MinConnections:        2,
    MaxConnectionLifetime: time.Hour,
    MaxConnectionIdleTime: 30 * time.Minute,
    ConnectionTimeout:     5 * time.Second,
}

pool, err := rasconfig.InitDbPool(ctx, cfg)
readOnlyPool, err := rasconfig.InitReadOnlyDbPool(ctx, cfg)
```

### rasconversion

Type conversion helpers for PostgreSQL (pgx/pgtype). Converts nullable Go types to pgtype equivalents with proper null handling.

```go
import "github.com/transactrx/ras-utils/rasconversion"

// Convert nullable Go types to pgtype (logs errors, returns invalid on failure)
pgText := rasconversion.ConvertToPgtypeString(stringPtr)
pgInt8 := rasconversion.ConvertToPgtypeInt8(int64Ptr)
pgInt2 := rasconversion.ConvertToPgtypeInt2(int32Ptr)
pgBool := rasconversion.ConvertToPgtypeBool(boolPtr)
pgTime := rasconversion.ConvertToPgtypeTimestamp(timePtr)

// Error-returning variants for explicit error handling
pgText, err := rasconversion.TryConvertToPgtypeString(stringPtr)
pgInt8, err := rasconversion.TryConvertToPgtypeInt8(int64Ptr)
pgInt2, err := rasconversion.TryConvertToPgtypeInt2(int32Ptr)
pgBool, err := rasconversion.TryConvertToPgtypeBool(boolPtr)
pgTime, err := rasconversion.TryConvertToPgtypeTimestamp(timePtr)
```

### raslogging

HTTP request logging middleware with panic recovery and structured JSON logging.

```go
import "github.com/transactrx/ras-utils/raslogging"

// Set up structured JSON logger (reads LOG_LEVEL env var)
raslogging.SetUpLogger()

// Logging middleware with panic recovery
logger := slog.Default()
loggingMw := raslogging.LoggingMiddleware(logger, "/health", "/ready") // skip paths optional
```

### rasevents

Event publishing via NATS with sync/async support, worker pools, graceful shutdown, and observability hooks.

```go
import "github.com/transactrx/ras-utils/rasevents"

// Option 1: Use global functions with package-level handler
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

// Option 2: Create independent handler instances
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

**Observability hooks:**

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

**Testing:**

```go
// Inject mock client for testing
rasevents.SetNatsClient(mockClient)
defer rasevents.ResetNatsClient()

// Or with handler instances
handler := rasevents.NewEventsHandler(cfg, mockClient)
```

**Environment variables:**
- `EVENTS_DEFAULT_NAMESPACE` - Default namespace (default: "PatientNotification")
- `EVENTS_SUBJECT` - Base NATS subject (default: "trx.eventscollector.collect")
- `EVENTS_TIMEOUT_SECONDS` - Request timeout in seconds (default: 60)
- `EVENTS_WORKER_POOL_SIZE` - Async worker count (default: 50)
- `EVENTS_QUEUE_SIZE` - Async queue size (default: 1000)

### rasstack

Middleware composition utility for chaining HTTP middleware.

```go
import "github.com/transactrx/ras-utils/rasstack"

// Compose multiple middleware
stack := rasstack.CreateStack(
    raslogging.LoggingMiddleware(logger),
    authMiddleware,
    rateLimitMiddleware,
)

http.Handle("/", stack(myHandler))
```

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

Event publishing via NATS with sync/async support, worker pools, and context cancellation.

```go
import "github.com/transactrx/ras-utils/rasevents"

// Optional: Initialize with custom config (otherwise uses defaults + env vars)
rasevents.Init(&rasevents.Config{
    DefaultNamespace: "MyService",
    Subject:          "trx.eventscollector.collect",
    Timeout:          30 * time.Second,
    WorkerPoolSize:   20,
    EventQueueSize:   500,
})

// Send event synchronously (returns error)
err := rasevents.SendEvent("PatientNotification", "Email", payload)

// Send with context for cancellation/timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
err := rasevents.SendEventWithContext(ctx, "PatientNotification", "SMS", payload)

// Send asynchronously (fire-and-forget, returns true if queued)
queued := rasevents.SendEventAsync("PatientNotification", "Email", payload)

// For testing: inject a mock client
rasevents.SetNatsClient(mockClient)
defer rasevents.ResetNatsClient()

// Graceful shutdown
defer rasevents.StopEventWorkerPool()
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

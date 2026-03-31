# ras-utils

Go utility library providing shared helper functions for the Clinical+ ecosystem.

## Installation

```bash
go get github.com/transactrx/ras-utils
```

## Packages

### cache

Generic in-memory key-value cache with TTL expiration and thread-safe operations.

```go
import "github.com/transactrx/ras-utils/cache"

// Create a cache (expired items removed on access)
c := cache.NewCache[string, User]()

// Create a cache with background cleanup (removes expired items periodically)
c := cache.NewCacheWithCleanup[string, User](5 * time.Minute)
defer c.Stop() // stop the cleanup goroutine when done

// Set with TTL
c.Set("user:123", user, 5*time.Minute)

// Get (returns zero value and false if expired/missing)
user, ok := c.Get("user:123")

// Cache-through pattern: fetch from source if not cached
user, ok := c.TryGet("user:123", func() (cache.CacheItem[User], bool) {
    user, err := db.GetUser(123)
    if err != nil {
        return cache.CacheItem[User]{}, false
    }
    return cache.NewCacheItem(user, time.Now().Add(5*time.Minute)), true
})

// Delete and Clear
c.Delete("user:123")
c.Clear()
```

### config

Database configuration and environment variable helpers.

```go
import "github.com/transactrx/ras-utils/config"

// Environment variables with defaults
host := config.GetEnvironmentVariableOrDefault("DB_HOST", "localhost")
port := config.GetEnvironmentVariableOrDefaultInt("DB_PORT", 5432)
timeout := config.GetEnvironmentVariableOrDefaultDuration("DB_TIMEOUT", "30s")

// Required environment variables (panics if missing)
apiKey := config.GetEnvironmentVariableOrPanic("API_KEY", "API_KEY is required")

// Database connection pool
cfg := &config.DBConfig{
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

pool, err := config.InitDbPool(ctx, cfg)
readOnlyPool, err := config.InitReadOnlyDbPool(ctx, cfg)
```

### conversion

Type conversion helpers for PostgreSQL (pgx/pgtype). Converts nullable Go types to pgtype equivalents with proper null handling.

```go
import "github.com/transactrx/ras-utils/conversion"

// Convert nullable Go types to pgtype (logs errors, returns invalid on failure)
pgText := conversion.ConvertToPgtypeString(stringPtr)
pgInt8 := conversion.ConvertToPgtypeInt8(int64Ptr)
pgInt2 := conversion.ConvertToPgtypeInt2(int32Ptr)
pgBool := conversion.ConvertToPgtypeBool(boolPtr)
pgTime := conversion.ConvertToPgtypeTimestamp(timePtr)

// Error-returning variants for explicit error handling
pgText, err := conversion.TryConvertToPgtypeString(stringPtr)
pgInt8, err := conversion.TryConvertToPgtypeInt8(int64Ptr)
pgInt2, err := conversion.TryConvertToPgtypeInt2(int32Ptr)
pgBool, err := conversion.TryConvertToPgtypeBool(boolPtr)
pgTime, err := conversion.TryConvertToPgtypeTimestamp(timePtr)
```

### middleware

HTTP middleware utilities for logging and middleware composition.

```go
import "github.com/transactrx/ras-utils/middleware"

// Set up structured JSON logger (reads LOG_LEVEL env var)
middleware.SetUpLogger()

// Logging middleware with panic recovery
logger := slog.Default()
loggingMw := middleware.LoggingMiddleware(logger, "/health", "/ready") // skip paths optional

// Compose multiple middleware
stack := middleware.CreateStack(
    middleware.LoggingMiddleware(logger),
    authMiddleware,
    rateLimitMiddleware,
)

http.Handle("/", stack(myHandler))
```

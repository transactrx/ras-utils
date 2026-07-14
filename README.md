# ras-utils

Go utility library providing shared helper functions for the Clinical+ ecosystem.

## Installation

```bash
go get github.com/transactrx/ras-utils
```

## Versioning

This library uses semantic versioning (`vX.Y.Z`). Tags are automatically created when PRs are merged into `main`, with the version increment determined by the branch prefix:

| Branch Prefix | Version Change | Example |
|---------------|----------------|---------|
| `major/*` | Increment major, reset minor and patch | v1.2.3 → v2.0.0 |
| `minor/*` | Increment minor, reset patch | v1.2.3 → v1.3.0 |
| `build/*` | Increment patch | v1.2.3 → v1.2.4 |
| `feature/*` | Increment patch (alias for build) | v1.2.3 → v1.2.4 |

# Packages

| Package | Description | README |
|---------|-------------|--------|
| [rascache](#rascache) | Generic TTL cache with singleflight | [README](rascache/README.md) |
| [rasconfig](#rasconfig) | Database pools and env helpers | [README](rasconfig/README.md) |
| [rasconversion](#rasconversion) | pgtype ↔ Go type conversions | [README](rasconversion/README.md) |
| [rasevents](#rasevents) | NATS event publishing | [README](rasevents/README.md) |
| [rashttp](#rashttp) | HTTP request/response helpers | [README](rashttp/README.md) |
| [raslocation](#raslocation) | Operating hours and scheduling | [README](raslocation/README.md) |
| [raslogging](#raslogging) | HTTP logging middleware | [README](raslogging/README.md) |
| [rasstack](#rasstack) | Middleware composition | [README](rasstack/README.md) |
| [rastime](#rastime) | TimeOfDay, DateRange, and schedule types | [README](rastime/README.md) |
| [rasvalidation](#rasvalidation) | UUID, email, phone, NPI validators | [README](rasvalidation/README.md) |
| [rasworker](#rasworker) | Worker pool with graceful shutdown | [README](rasworker/README.md) |

## rascache

Generic in-memory key-value cache with TTL expiration and thread-safe operations. Supports both local time and UTC-based expiration via functional options.

**Singleflight protection:** `GetOrStore` uses singleflight internally to prevent thundering herd on cache miss. When multiple goroutines request the same missing key simultaneously, only one executes the fetch callback while others wait and share the result.

```go
import "github.com/transactrx/ras-utils/rascache"

// Create a cache using local time (expired items removed on access)
c := rascache.NewCache[string, User]()

// Create a cache using UTC time for expiration
c := rascache.NewCache[string, User](rascache.WithUTC())

// Create a cache with background cleanup (removes expired items periodically)
c := rascache.NewCache[string, User](rascache.WithCleanup(5 * time.Minute))
defer c.Stop() // stop the cleanup goroutine when done

// Combine options
c := rascache.NewCache[string, User](rascache.WithCleanup(5*time.Minute), rascache.WithUTC())
defer c.Stop()

// Set with expiration time (assumes server time or UTC depending on cache init options)
c.Set("user:123", user, time.Now().Add(5*time.Minute))

// Get (returns zero value and false if expired/missing)
user, ok := c.Get("user:123")

// Cache-through pattern: fetch from source if not cached
// Safe for concurrent use - only one goroutine fetches on cache miss
user, ok := c.GetOrStore("user:123", func() (User, time.Time, bool) {
    user, err := db.GetUser(123)
    if err != nil {
        return User{}, time.Time{}, false
    }
    return user, time.Now().Add(5*time.Minute), true
})

// Delete and Clear
c.Delete("user:123")
c.Clear()
```

**Token caching pattern:** For OAuth tokens or similar credentials with expiry, cache at half the token lifetime to allow refresh before expiration:

```go
tokenCache := rascache.NewCache[string, string]()

func GetAuthToken() (string, bool) {
    return tokenCache.GetOrStore("auth_token", func() (string, time.Time, bool) {
        token, err := fetchTokenFromAuthServer()
        if err != nil {
            return "", time.Time{}, false
        }
        // Cache for half the token lifetime
        expiry := time.Now().Add(time.Duration(token.ExpiresIn/2) * time.Second)
        return token.AccessToken, expiry, true
    })
}
```

## rasconfig

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

// Snowflake connection pool (JWT authentication with private key)
sfCfg := &rasconfig.SnowflakeDBConfig{
    Host:                  "account.snowflakecomputing.com",
    Port:                  443,
    Database:              "MY_DB",
    Schema:                "PUBLIC",
    Warehouse:             "MY_WAREHOUSE",
    User:                  "my_user",
    PrivateKey:            "base64-encoded-pkcs8-private-key",
    MaxConnections:        10,
    MaxIdleConnections:    5,
    MaxConnectionLifetime: 30 * time.Minute,
}

db, err := rasconfig.NewSnowflakePool(ctx, sfCfg)
if err != nil {
    log.Fatal(err)
}
defer db.Close()

// PrivateKey must be base64-encoded PKCS8 format (PEM or DER).
// Generate with: openssl genrsa 2048 | openssl pkcs8 -topk8 -nocrypt | base64 -w0
```

## rasconversion

Type conversion helpers for PostgreSQL (pgx/pgtype). Converts nullable Go types to pgtype equivalents with proper null handling, and vice versa.

```go
import "github.com/transactrx/ras-utils/rasconversion"
```

### Conversion Functions

| pgtype | Go Type | To pgtype | From pgtype | OrDefault |
|--------|---------|-----------|-------------|-----------|
| `Text` | `*string` | `ConvertToPgtypeString` | `ConvertFromPgtypeText` | `ConvertFromPgtypeTextOrDefault` |
| `Int2` | `*int32` / `*int16` | `ConvertToPgtypeInt2` | `ConvertFromPgtypeInt2` | `ConvertFromPgtypeInt2OrDefault` |
| `Int4` | `*int32` | `ConvertToPgtypeInt4` | `ConvertFromPgtypeInt4` | `ConvertFromPgtypeInt4OrDefault` |
| `Int8` | `*int64` | `ConvertToPgtypeInt8` | `ConvertFromPgtypeInt8` | `ConvertFromPgtypeInt8OrDefault` |
| `Bool` | `*bool` | `ConvertToPgtypeBool` | `ConvertFromPgtypeBool` | `ConvertFromPgtypeBoolOrDefault` |
| `Float4` | `*float32` | `ConvertToPgtypeFloat4` | `ConvertFromPgtypeFloat4` | `ConvertFromPgtypeFloat4OrDefault` |
| `Float8` | `*float64` | `ConvertToPgtypeFloat8` | `ConvertFromPgtypeFloat8` | `ConvertFromPgtypeFloat8OrDefault` |
| `Numeric` | `*float64` | `ConvertToPgtypeNumeric` | `ConvertFromPgtypeNumeric` | `ConvertFromPgtypeNumericOrDefault` |
| `Numeric` | `*big.Float` | — | `ConvertFromPgtypeNumericToBigFloat` | — |
| `Timestamp` | `*time.Time` | `ConvertToPgtypeTimestamp` | `ConvertFromPgtypeTimestamp` | `ConvertFromPgtypeTimestampOrDefault` |
| `Timestamptz` | `*time.Time` | `ConvertToPgtypeTimestamptz` | `ConvertFromPgtypeTimestamptz` | `ConvertFromPgtypeTimestamptzOrDefault` |
| `Date` | `*time.Time` | `ConvertToPgtypeDate` | `ConvertFromPgtypeDate` | `ConvertFromPgtypeDateOrDefault` |
| `Time` | `*time.Time` | `ConvertToPgtypeTime` | `ConvertFromPgtypeTime` | — |
| `Time` | `*rastime.TimeOfDay` | `ConvertTimeOfDayToPgtypeTime` | `ConvertPgtypeTimeToTimeOfDay` | — |
| `Interval` | `*time.Duration` | `ConvertToPgtypeInterval` | `ConvertFromPgtypeInterval` | `ConvertFromPgtypeIntervalOrDefault` |
| `UUID` | `*uuid.UUID` | `ConvertToPgtypeUUID` | `ConvertFromPgtypeUUID` | `ConvertFromPgtypeUUIDOrDefault` |
| `UUID` | `*string` | `ConvertToPgtypeUUIDFromString` | `ConvertFromPgtypeUUIDToString` | `ConvertFromPgtypeUUIDToStringOrDefault` |
| `JSONB` | `any` / `[]byte` | `ConvertToPgtypeJSONB` | `ConvertFromPgtypeJSONB[T]` | — |

### Behavior

**To pgtype functions:**
- Return `Valid: false` if input pointer is nil
- Log errors and return `Valid: false` on conversion failure

**From pgtype functions:**
- Return `nil` if pgtype has `Valid: false`

**OrDefault functions:**
- Return the provided default value if pgtype has `Valid: false`
- Useful when you want a value type instead of a pointer

**Error-returning variants:** All "To pgtype" functions have a `TryConvert*` variant that returns an error instead of logging:

```go
pgText, err := rasconversion.TryConvertToPgtypeString(stringPtr)
pgInt4, err := rasconversion.TryConvertToPgtypeInt4(int32Ptr)
pgNumeric, err := rasconversion.TryConvertToPgtypeNumeric(float64Ptr)
pgInterval, err := rasconversion.TryConvertToPgtypeInterval(durationPtr)
jsonBytes, err := rasconversion.TryConvertToPgtypeJSONB(anyValue)
// ... etc for all ConvertTo* functions
```

### JSONB Example

```go
// Marshal any Go value to JSONB bytes
type UserPrefs struct {
    Theme    string `json:"theme"`
    Timezone string `json:"timezone"`
}
prefs := UserPrefs{Theme: "dark", Timezone: "America/New_York"}
jsonBytes := rasconversion.ConvertToPgtypeJSONB(prefs)

// Unmarshal JSONB bytes to a typed struct (generic)
result := rasconversion.ConvertFromPgtypeJSONB[UserPrefs](jsonBytes)
if result != nil {
    fmt.Println(result.Theme) // "dark"
}

// With error handling
result, err := rasconversion.TryConvertFromPgtypeJSONB[UserPrefs](jsonBytes)
```

### Interval Example

```go
// Duration to pgtype.Interval
d := 2*time.Hour + 30*time.Minute
pgInterval := rasconversion.ConvertToPgtypeInterval(&d)

// pgtype.Interval to Duration
duration := rasconversion.ConvertFromPgtypeInterval(pgInterval)

// With default fallback
duration := rasconversion.ConvertFromPgtypeIntervalOrDefault(pgInterval, time.Hour)
```

## rastime

Time-of-day utilities for schedule management. Provides `TimeOfDay` (hour/minute without date), `TimeRange` (start/end windows), `DateRange` (date/time periods), and `DayOfWeek` constants with validation, comparison, and JSON serialization.

```go
import "github.com/transactrx/ras-utils/rastime"

// Create TimeOfDay with validation
tod, err := rastime.NewTimeOfDay(9, 30)
if err != nil {
    // invalid hour/minute
}

// Parse from strings (24-hour or 12-hour format)
tod, _ := rastime.ParseTimeOfDay("14:30")
tod, _ := rastime.ParseTimeOfDay("2:30 PM")

// Extract from time.Time
tod := rastime.TimeOfDayFromTime(time.Now())

// Comparison
if tod.Before(other) { }
if tod.After(other) { }
if tod.Equal(other) { }
if tod.Between(start, end) { }  // inclusive start, exclusive end

// Arithmetic (wraps at midnight)
later := tod.AddMinutes(90)
earlier := tod.AddMinutes(-30)

// Round to slot boundaries
slot := tod.RoundUpToNextSlot(15)  // 9:07 -> 9:15

// Convert to minutes since midnight
minutes := tod.ToMinutes()

// TimeRange for operating windows
tr := rastime.TimeRange{Start: start, End: end}
if tr.Contains(tod) { }
duration := tr.Duration()  // time.Duration

// Check for overlapping ranges
if tr.Overlaps(otherRanges) { }

// DayOfWeek constants with String() support
day := rastime.MON
fmt.Println(day.String())  // "Monday"
fmt.Println(day.Short())   // "Mon"

// JSON marshaling (TimeOfDay serializes as "HH:MM" string)
data, _ := json.Marshal(tod)  // "09:30"

// DateRange for date/time periods
dr := rastime.CalendarYear(2025)              // Jan 1 2025 to Jan 1 2026
dr := rastime.RollingYearFrom(firstAlert)     // 1 year from a date
dr, err := rastime.NewDateRange(start, end)   // with validation

if dr.Contains(t) { }           // half-open [Start, End)
if dr.ContainsInclusive(t) { }  // closed [Start, End]
if dr.Overlaps(other) { }       // check overlap
nextPeriod := dr.NextAnnualPeriod()  // shift forward 1 year
```

## raslocation

Location-specific time and scheduling utilities for the Clinical+ ecosystem. Handles operating hours, timezone-aware open/close checks, and schedule merging.

```go
import "github.com/transactrx/ras-utils/raslocation"
import "github.com/transactrx/ras-utils/rastime"

// Create default hours (same hours all 7 days)
start, _ := rastime.NewTimeOfDay(9, 0)
end, _ := rastime.NewTimeOfDay(17, 0)
hours := raslocation.NewDefaultLocationHours(true, start, end)

// Check if open at a specific time
if hours.IsOpenAt(time.Now()) {
    // location is open
}

// Check with timezone conversion
if hours.IsOpenAtZone(time.Now().UTC(), "America/New_York") {
    // open in New York time
}

// Get minutes until close (0 if not open)
minutes := hours.MinutesUntilClose(time.Now())

// Get total weekly open minutes
totalMinutes := hours.WeeklyOpenMinutes()

// Find next open window
nextOpen := hours.GetNextOpenWindow(time.Now())
if !nextOpen.IsZero() {
    fmt.Printf("Opens at %v\n", nextOpen)
}

// Merge campaign overrides onto base hours
// Only replaces entries where DayOfWeek AND IsOpen match
effective := baseHours.Merge(campaignOverrides)

// Validate for overlapping time ranges
if err := hours.Validate(); err != nil {
    log.Printf("Invalid schedule: %v", err)
}

// Get open days only
openDays := hours.GetOpenDays()

// Human-readable schedule
fmt.Println(hours.String()) // "[Mon 09:00-17:00 Tue 09:00-17:00 ...]"
```

## raslogging

HTTP request logging middleware with panic recovery and structured JSON logging.

```go
import "github.com/transactrx/ras-utils/raslogging"

// Set up structured JSON logger (reads LOG_LEVEL env var)
raslogging.SetUpLogger()

// Logging middleware with panic recovery
logger := slog.Default()
loggingMw := raslogging.LoggingMiddleware(logger, "/health", "/ready") // skip paths optional
```

## rasevents

Event publishing via NATS with sync/async support, worker pools, graceful shutdown, and observability hooks.

### Environment Variables


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
- `EVENTS_DEFAULT_NAMESPACE` - Default namespace (required)
- `EVENTS_SUBJECT` - Base NATS subject (required)
- `EVENTS_TIMEOUT_SECONDS` - Request timeout in seconds (default: 60)
- `EVENTS_WORKER_POOL_SIZE` - Async worker count (default: 50)
- `EVENTS_QUEUE_SIZE` - Async queue size (default: 1000)

## rashttp

HTTP helper functions for request parsing, response writing, and common patterns.

```go
import "github.com/transactrx/ras-utils/rashttp"

// Request helpers
ip := rashttp.GetClientIP(r)              // extracts from X-Forwarded-For, X-Real-IP, or RemoteAddr
url := rashttp.GetFullRequestURL(r)       // reconstructs full URL including scheme/host from proxied requests
token := rashttp.GetBearerToken(r)        // extracts bearer token from Authorization header
isHtmx := rashttp.IsHTMX(r)               // checks HX-Request header
isAjax := rashttp.IsAjax(r)               // checks X-Requested-With header

// Query parameter parsing with defaults
page := rashttp.QueryInt(r, "page", 1)
sort := rashttp.QueryString(r, "sort", "created_at")

// JSON request body decoding (with size limit)
type CreateUserRequest struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}
payload, err := rashttp.DecodeJSON[CreateUserRequest](r, rashttp.DefaultMaxBodySize)

// Response helpers — generic
rashttp.WriteJSON(w, http.StatusOK, data)
rashttp.WriteError(w, http.StatusBadRequest, "invalid input")

// Response helpers — status shorthands
rashttp.OK(w, data)                              // 200
rashttp.Created(w, data)                          // 201
rashttp.Accepted(w, data)                         // 202
rashttp.NoContent(w)                              // 204
rashttp.BadRequest(w, "missing field")            // 400
rashttp.Unauthorized(w, "invalid token")          // 401
rashttp.Forbidden(w, "not allowed")               // 403
rashttp.NotFound(w, "resource not found")         // 404
rashttp.Conflict(w, "already exists")             // 409
rashttp.UnprocessableEntity(w, "validation error")// 422
rashttp.TooManyRequests(w, "rate limited")        // 429
rashttp.InternalServerError(w, "unexpected error")// 500
rashttp.ServiceUnavailable(w, "try again later")  // 503

// Health check handler
http.Handle("/health", rashttp.HealthHandler(func() error {
    return db.Ping() // returns 200 if nil, 503 if error
}))
```

## rasstack

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

## rasauth

OAuth2 client credentials token acquisition. Supports Basic Auth header or form body credentials, configurable scopes, and custom parameters for different identity providers.

```go
import "github.com/transactrx/ras-utils/rasauth"

// Basic Auth (credentials in Authorization header)
token, err := rasauth.GetToken(rasauth.AuthConfig{
    ClientID:     "my-client",
    ClientSecret: "my-secret",
    TokenURL:     "https://auth.example.com/oauth/token",
    UseBasicAuth: true,
})

// Form body credentials with scopes
token, err := rasauth.GetToken(rasauth.AuthConfig{
    ClientID:     "my-client",
    ClientSecret: "my-secret",
    TokenURL:     "https://auth.example.com/oauth/token",
    Scopes:       []string{"openid", "profile"},
})

// With extra parameters (e.g., audience for Auth0)
token, err := rasauth.GetToken(rasauth.AuthConfig{
    ClientID:     "my-client",
    ClientSecret: "my-secret",
    TokenURL:     "https://auth.example.com/oauth/token",
    UseBasicAuth: true,
    ExtraParams: map[string]string{
        "audience": "https://api.example.com",
    },
})

// Custom grant type and timeout
token, err := rasauth.GetToken(rasauth.AuthConfig{
    ClientID:     "my-client",
    ClientSecret: "my-secret",
    TokenURL:     "https://auth.example.com/oauth/token",
    GrantType:    "refresh_token",
    Timeout:      5 * time.Second,
    ExtraParams: map[string]string{
        "refresh_token": "existing-refresh-token",
    },
})

// Access token response fields
fmt.Println(token.AccessToken)   // "eyJ..."
fmt.Println(token.ExpiresIn)     // 3600
fmt.Println(token.TokenType)     // "Bearer"
fmt.Println(token.RefreshToken)  // optional
fmt.Println(token.IDToken)       // optional (OpenID Connect)
fmt.Println(token.Scope)         // optional

// Error handling
token, err := rasauth.GetToken(config)
if err != nil {
    if authErr, ok := err.(*rasauth.AuthError); ok {
        log.Printf("Auth failed: %d %s", authErr.StatusCode, authErr.Message)
    }
}

// Legacy helpers (for backwards compatibility)
token, err := rasauth.GetCISToken(clientID, clientSecret, tokenURL)  // form body + openid scope
token, err := rasauth.GetJWTToken(clientID, clientSecret, tokenURL)  // Basic Auth
```

## rasvalidation

Common validation helpers for strings, identifiers, and dates.

```go
import "github.com/transactrx/ras-utils/rasvalidation"

// UUID validation (non-empty, valid format, non-nil)
if rasvalidation.IsValidUUID("550e8400-e29b-41d4-a716-446655440000") {
    // valid
}
rasvalidation.IsValidUUID("")                                      // false
rasvalidation.IsValidUUID("00000000-0000-0000-0000-000000000000")  // false (nil UUID)

// Email validation (RFC 5322)
rasvalidation.IsValidEmail("user@example.com")     // true
rasvalidation.IsValidEmail("user+tag@example.com") // true
rasvalidation.IsValidEmail("invalid")              // false

// US phone number (10 digits, various formats)
rasvalidation.IsValidUSPhone("5551234567")      // true
rasvalidation.IsValidUSPhone("555-123-4567")    // true
rasvalidation.IsValidUSPhone("(555)123-4567")   // true
rasvalidation.IsValidUSPhone("1-555-123-4567")  // true

// US ZIP code (5 or 5+4 format)
rasvalidation.IsValidUSZip("12345")      // true
rasvalidation.IsValidUSZip("12345-6789") // true

// NPI validation (10 digits with Luhn checksum)
rasvalidation.IsValidNPI("1234567893")       // true (valid checksum)
rasvalidation.IsValidNPI("1234567890")       // false (invalid checksum)
rasvalidation.IsValidNPIFormat("1234567890") // true (format only, no checksum)
rasvalidation.IsValidNPIChecksum("1234567893") // true (checksum only)

// Date validation
rasvalidation.IsValidISO8601Date("2024-01-15")    // true (YYYY-MM-DD)
rasvalidation.IsValidMMDDYYYYDate("01/15/2024")   // true (MM/DD/YYYY)
rasvalidation.IsValidDateString("15-01-2024", "02-01-2006") // custom layout
```

## rasworker

Generic worker pool for concurrent job execution with graceful shutdown and error handling.

```go
import "github.com/transactrx/ras-utils/rasworker"

// Create pool with 10 workers and queue size of 100
pool := rasworker.NewPool(10, 100)
pool.Start()

// Submit jobs (returns false if queue is full)
ok := pool.Submit(func(ctx context.Context) error {
    // do work
    return nil
})

// Graceful shutdown - drains queue before returning
err := pool.Shutdown(context.Background())

// Shutdown with timeout - cancels in-flight jobs if deadline exceeded
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
if err := pool.Shutdown(ctx); err != nil {
    log.Printf("Shutdown timeout: %v", err)
}
```

**Error handling:**

```go
// Create pool with error handler
pool := rasworker.NewPoolWithErrorHandler(10, 100, func(err error) {
    log.Printf("Job failed: %v", err)
    metrics.IncrCounter("worker_errors")
})

// Or add handlers after creation (thread-safe, can be called after Start)
pool := rasworker.NewPool(10, 100)
pool.AddErrorHandler(func(err error) {
    slog.Error("job error", "error", err)
})
pool.AddErrorHandler(func(err error) {
    alerting.Notify(err)
})
pool.Start()
```

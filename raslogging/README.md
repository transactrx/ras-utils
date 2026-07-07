# raslogging

HTTP request logging middleware with panic recovery and structured JSON logging.

## Features

- Structured JSON logging via `log/slog`
- Request/response logging with timing
- Panic recovery middleware
- Configurable path exclusions
- LOG_LEVEL environment variable support

## Installation

```go
import "github.com/transactrx/ras-utils/raslogging"
```

## Usage

### Logger Setup

```go
// Set up structured JSON logger (reads LOG_LEVEL env var)
raslogging.SetUpLogger()
```

The `LOG_LEVEL` environment variable accepts: `debug`, `info`, `warn`, `error` (default: `info`).

### Logging Middleware

```go
logger := slog.Default()

// Basic usage
loggingMw := raslogging.LoggingMiddleware(logger)

// With path exclusions (e.g., health checks)
loggingMw := raslogging.LoggingMiddleware(logger, "/health", "/ready")

// Apply to routes
mux.Handle("/api/users", loggingMw(usersHandler))
```

### Middleware Stack

```go
import "github.com/transactrx/ras-utils/rasstack"

stack := rasstack.CreateStack(
    raslogging.LoggingMiddleware(logger),
    authMiddleware,
)

mux.Handle("/", stack(myHandler))
```

## Log Output

Each request logs a JSON object with:

```json
{
  "time": "2024-01-15T10:30:00Z",
  "level": "INFO",
  "msg": "request",
  "method": "GET",
  "path": "/api/users",
  "status": 200,
  "duration_ms": 45,
  "client_ip": "192.168.1.1"
}
```

## API Reference

### Functions

- `SetUpLogger()` - Configure global slog logger with JSON output
- `LoggingMiddleware(logger *slog.Logger, skipPaths ...string) func(http.Handler) http.Handler`

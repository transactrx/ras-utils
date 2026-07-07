# rasstack

Middleware composition utility for chaining HTTP middleware.

## Features

- Simple middleware chaining
- Standard `http.Handler` compatibility
- Clean, readable middleware stacks

## Installation

```go
import "github.com/transactrx/ras-utils/rasstack"
```

## Usage

### Creating a Middleware Stack

```go
// Compose multiple middleware
stack := rasstack.CreateStack(
    raslogging.LoggingMiddleware(logger),
    authMiddleware,
    rateLimitMiddleware,
)

// Apply to a handler
http.Handle("/", stack(myHandler))
```

### Common Patterns

```go
// Different stacks for different route groups
publicStack := rasstack.CreateStack(
    raslogging.LoggingMiddleware(logger),
    rateLimitMiddleware,
)

authStack := rasstack.CreateStack(
    raslogging.LoggingMiddleware(logger),
    authMiddleware,
    rateLimitMiddleware,
)

adminStack := rasstack.CreateStack(
    raslogging.LoggingMiddleware(logger),
    authMiddleware,
    adminOnlyMiddleware,
)

mux.Handle("/public", publicStack(publicHandler))
mux.Handle("/api", authStack(apiHandler))
mux.Handle("/admin", adminStack(adminHandler))
```

### Middleware Signature

Middleware must follow the standard Go pattern:

```go
func myMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // before
        next.ServeHTTP(w, r)
        // after
    })
}
```

## API Reference

### Types

- `Middleware` - Function type `func(http.Handler) http.Handler`

### Functions

- `CreateStack(middlewares ...Middleware) Middleware` - Compose middleware into a single middleware

# rashttp

HTTP helper functions for request parsing, response writing, and common patterns.

## Features

- Request parsing (client IP, bearer token, query params)
- Generic JSON request body decoding
- Status-specific response helpers
- Health check handler

## Installation

```go
import "github.com/transactrx/ras-utils/rashttp"
```

## Usage

### Request Helpers

```go
// Client information
ip := rashttp.GetClientIP(r)              // extracts from X-Forwarded-For, X-Real-IP, or RemoteAddr
url := rashttp.GetFullRequestURL(r)       // reconstructs full URL including scheme/host from proxied requests
token := rashttp.GetBearerToken(r)        // extracts bearer token from Authorization header

// Request type detection
isHtmx := rashttp.IsHTMX(r)               // checks HX-Request header
isAjax := rashttp.IsAjax(r)               // checks X-Requested-With header

// Query parameter parsing with defaults
page := rashttp.QueryInt(r, "page", 1)
sort := rashttp.QueryString(r, "sort", "created_at")
```

### JSON Request Body Decoding

```go
type CreateUserRequest struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}
payload, err := rashttp.DecodeJSON[CreateUserRequest](r, rashttp.DefaultMaxBodySize)
```

### Response Helpers

```go
// Generic
rashttp.WriteJSON(w, http.StatusOK, data)
rashttp.WriteError(w, http.StatusBadRequest, "invalid input")

// Status shorthands
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
```

### Health Check Handler

```go
http.Handle("/health", rashttp.HealthHandler(func() error {
    return db.Ping() // returns 200 if nil, 503 if error
}))
```

## API Reference

### Constants

- `DefaultMaxBodySize` - Default maximum request body size for JSON decoding

### Request Functions

- `GetClientIP(r *http.Request) string`
- `GetFullRequestURL(r *http.Request) string`
- `GetBearerToken(r *http.Request) string`
- `IsHTMX(r *http.Request) bool`
- `IsAjax(r *http.Request) bool`
- `QueryInt(r *http.Request, key string, defaultValue int) int`
- `QueryString(r *http.Request, key string, defaultValue string) string`
- `DecodeJSON[T any](r *http.Request, maxSize int64) (T, error)`

### Response Functions

- `WriteJSON(w http.ResponseWriter, status int, data any)`
- `WriteError(w http.ResponseWriter, status int, message string)`
- `OK(w http.ResponseWriter, data any)`
- `Created(w http.ResponseWriter, data any)`
- `Accepted(w http.ResponseWriter, data any)`
- `NoContent(w http.ResponseWriter)`
- `BadRequest(w http.ResponseWriter, message string)`
- `Unauthorized(w http.ResponseWriter, message string)`
- `Forbidden(w http.ResponseWriter, message string)`
- `NotFound(w http.ResponseWriter, message string)`
- `Conflict(w http.ResponseWriter, message string)`
- `UnprocessableEntity(w http.ResponseWriter, message string)`
- `TooManyRequests(w http.ResponseWriter, message string)`
- `InternalServerError(w http.ResponseWriter, message string)`
- `ServiceUnavailable(w http.ResponseWriter, message string)`
- `HealthHandler(check func() error) http.Handler`

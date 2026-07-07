# rascache

Generic in-memory key-value cache with TTL expiration and thread-safe operations.

## Features

- Generic type support (`Cache[K, V]`)
- Thread-safe operations with `sync.RWMutex`
- TTL-based expiration with absolute expiry times
- Optional background cleanup goroutine
- Local time or UTC-based expiration
- Singleflight protection on cache miss (prevents thundering herd)

## Installation

```go
import "github.com/transactrx/ras-utils/rascache"
```

## Usage

### Basic Cache

```go
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
```

### Operations

```go
// Set with expiration time
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

### Token Caching Pattern

For OAuth tokens or similar credentials with expiry, cache at half the token lifetime to allow refresh before expiration:

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

## API Reference

### Types

- `Cache[K comparable, V any]` - Generic cache type
- `Option` - Functional option for cache configuration

### Functions

- `NewCache[K, V](opts ...Option) *Cache[K, V]` - Create a new cache
- `WithUTC() Option` - Use UTC time for expiration checks
- `WithCleanup(interval time.Duration) Option` - Enable background cleanup

### Methods

- `Set(key K, value V, expiry time.Time)` - Store a value with expiration
- `Get(key K) (V, bool)` - Retrieve a value
- `GetOrStore(key K, fn func() (V, time.Time, bool)) (V, bool)` - Get or compute value
- `Delete(key K)` - Remove a value
- `Clear()` - Remove all values
- `Stop()` - Stop background cleanup goroutine

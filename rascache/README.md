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

// Cache-through with error handling (TTL as duration)
user, err := c.GetOrStoreWithError("user:123", func() (User, time.Duration, error) {
    user, err := db.GetUser(123)
    if err != nil {
        return User{}, 0, err
    }
    return user, 5*time.Minute, nil
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

// Or using GetOrStoreWithError with TTL duration
func GetAuthTokenWithError() (string, error) {
    return tokenCache.GetOrStoreWithError("auth_token", func() (string, time.Duration, error) {
        token, err := fetchTokenFromAuthServer()
        if err != nil {
            return "", 0, err
        }
        // Cache for half the token lifetime
        ttl := time.Duration(token.ExpiresIn/2) * time.Second
        return token.AccessToken, ttl, nil
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
- `GetOrStore(key K, fn StoreCacheCallback[V]) (V, bool)` - Get or fetch value with absolute expiry
- `GetOrStoreWithError(key K, fn StoreCacheCallbackWithError[V]) (V, error)` - Get or fetch value with TTL duration and error handling
- `Delete(key K)` - Remove a value
- `Clear()` - Remove all values
- `Stop()` - Stop background cleanup goroutine

### Callback Types

- `StoreCacheCallback[T] func() (T, time.Time, bool)` - Returns value, absolute expiry time, success flag
- `StoreCacheCallbackWithError[T] func() (T, time.Duration, error)` - Returns value, TTL duration, error

## Notes

### Time Mode and TTL

The cache's time mode (`WithUTC()` or default local time) affects how expiration is calculated:

- **`GetOrStore`**: You provide an absolute `time.Time` expiry, so ensure it matches your cache's time mode (use `time.Now().UTC()` for UTC caches)
- **`GetOrStoreWithError`**: You provide a `time.Duration` TTL, which is added to the cache's internal `now()` (UTC or local depending on configuration)

### Singleflight Behavior

Both `GetOrStore` and `GetOrStoreWithError` use singleflight to deduplicate concurrent fetches for the same key. If the fetch operation fails (returns `false` or an error), **all waiting goroutines receive that same failure**. If you need retry logic for transient errors, implement it outside these methods or in your callback.

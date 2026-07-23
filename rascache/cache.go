// Package rascache provides a generic in-memory key-value cache with TTL expiration.
//
// The cache is thread-safe and supports both local time and UTC-based expiration
// via functional options. Optional background cleanup can remove expired items
// periodically without waiting for access.
//
// Basic usage:
//
//	c := rascache.NewCache[string, User]()
//	c.Set("user:123", user, time.Now().Add(5*time.Minute))
//	if user, ok := c.Get("user:123"); ok {
//	    // use user
//	}
package rascache

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

// ICacheable defines the interface for cache items with expiration behavior.
type ICacheable[T any] interface {
	isExpired() bool
	getValue() T
	getExpiration() time.Time
	getTtl() time.Duration
}

// CacheItem represents an item stored in the cache with its associated TTL based on server time.
type CacheItem[T any] struct {
	value  T
	expiry time.Time
}

// CacheItemUTC represents an item stored in the cache with its associated TTL based on UTC time.
type CacheItemUTC[T any] struct {
	value  T
	expiry time.Time
}

// isExpired reports whether the cache item has expired using UTC time.
func (ci *CacheItemUTC[T]) isExpired() bool {
	return ci.expiry.Before(time.Now().UTC())
}

// isExpired reports whether the cache item has expired using local time.
func (ci *CacheItem[T]) isExpired() bool {
	return ci.expiry.Before(time.Now())
}

// getValue returns the cached value.
func (ci *CacheItemUTC[T]) getValue() T {
	return ci.value
}

// getValue returns the cached value.
func (ci *CacheItem[T]) getValue() T {
	return ci.value
}

// getExpiration returns the expiration time.
func (ci *CacheItemUTC[T]) getExpiration() time.Time {
	return ci.expiry
}

// getExpiration returns the expiration time.
func (ci *CacheItem[T]) getExpiration() time.Time {
	return ci.expiry
}

// getTtl returns the remaining time until expiration.
func (ci *CacheItemUTC[T]) getTtl() time.Duration {
	return ci.expiry.Sub(time.Now().UTC())
}

// getTtl returns the remaining time until expiration.
func (ci *CacheItem[T]) getTtl() time.Duration {
	return time.Until(ci.expiry)
}

// NewCacheItem creates a new [CacheItem] that uses local time for expiration checks.
func NewCacheItem[T any](value T, expiry time.Time) ICacheable[T] {
	return &CacheItem[T]{
		value:  value,
		expiry: expiry,
	}
}

// NewCacheItemUTC creates a new [CacheItemUTC] that uses UTC time for expiration checks.
func NewCacheItemUTC[T any](value T, expiry time.Time) ICacheable[T] {
	return &CacheItemUTC[T]{
		value:  value,
		expiry: expiry,
	}
}

// Cache is a generic in-memory key-value store with expiry support.
// It is safe for concurrent use.
type Cache[K comparable, T any] struct {
	data    map[K]ICacheable[T]              // stores cache items
	mu      sync.RWMutex                     // managing concurrent access
	stopCh  chan struct{}                    // signal to stop background cleanup
	newItem func(T, time.Time) ICacheable[T] // factory for creating cache items
	group   singleflight.Group               // deduplicates concurrent fetches in GetOrStore
	useUTC  bool                             // time mode for TTL calculations
}

// cacheConfig holds configuration for a Cache instance.
type cacheConfig struct {
	useUTC          bool
	cleanupInterval time.Duration
}

// CacheOption configures a Cache.
type CacheOption func(*cacheConfig)

// WithUTC configures the cache to use UTC time for expiration checks.
func WithUTC() CacheOption {
	return func(c *cacheConfig) {
		c.useUTC = true
	}
}

// WithCleanup configures the cache to run background cleanup at the given interval.
func WithCleanup(interval time.Duration) CacheOption {
	return func(c *cacheConfig) {
		c.cleanupInterval = interval
	}
}

// NewCache creates and initializes a new [Cache] instance.
// Use [WithUTC] for UTC-based expiration or [WithCleanup] for background cleanup.
func NewCache[K comparable, T any](opts ...CacheOption) *Cache[K, T] {
	cfg := &cacheConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	c := &Cache[K, T]{
		data:   make(map[K]ICacheable[T]),
		useUTC: cfg.useUTC,
	}

	if cfg.useUTC {
		c.newItem = func(v T, exp time.Time) ICacheable[T] {
			return &CacheItemUTC[T]{value: v, expiry: exp}
		}
	} else {
		c.newItem = func(v T, exp time.Time) ICacheable[T] {
			return &CacheItem[T]{value: v, expiry: exp}
		}
	}

	if cfg.cleanupInterval > 0 {
		c.stopCh = make(chan struct{})
		go c.startCleanup(cfg.cleanupInterval)
	}

	return c
}

// Deprecated: Use NewCache(WithCleanup(interval)) instead.
func NewCacheWithCleanup[K comparable, T any](cleanupInterval time.Duration) *Cache[K, T] {
	return NewCache[K, T](WithCleanup(cleanupInterval))
}

// Stop terminates the background cleanup goroutine.
// It is safe to call on caches without cleanup enabled (no-op).
func (c *Cache[K, T]) Stop() {
	if c.stopCh != nil {
		close(c.stopCh)
	}
}

// startCleanup runs the background cleanup loop.
func (c *Cache[K, T]) startCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.deleteExpired()
		case <-c.stopCh:
			return
		}
	}
}

// deleteExpired removes all expired items from the cache.
func (c *Cache[K, T]) deleteExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for key, item := range c.data {
		if item.isExpired() {
			delete(c.data, key)
		}
	}
}

// Set adds or updates a key-value pair in the cache with the given expiration.
func (c *Cache[K, T]) Set(key K, value T, expiresOn time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = c.newItem(value, expiresOn)
}

func zeroVal[T any]() T {
	var zero T
	return zero
}

// Get retrieves the value associated with the given key from the cache.
// It also checks for expiry and removes expired items.
func (c *Cache[K, T]) Get(key K) (T, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	item, ok := c.data[key]
	if !ok {
		return zeroVal[T](), false
	}
	// item found - check for expiry
	if item.isExpired() {
		// remove entry from cache if time is beyond the expiry
		delete(c.data, key)
		return zeroVal[T](), false
	}
	return item.getValue(), true
}

// Delete removes a key-value pair from the cache.
func (c *Cache[K, T]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
}

// Clear removes all key-value pairs from the cache.
func (c *Cache[K, T]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[K]ICacheable[T])
}

// now returns the current time in the cache's configured time mode.
func (c *Cache[K, T]) now() time.Time {
	if c.useUTC {
		return time.Now().UTC()
	}
	return time.Now()
}

// StoreCacheOperation is a function that fetches a value to cache on miss.
// Returns the cacheable item and true on success, or zero value and false on failure.
// Deprecated: Use GetOrStore with StoreCacheCallback instead.
type StoreCacheOperation[T any] func() (CacheItem[T], bool)

// StoreCacheCallback is a function that fetches a value to cache on miss.
// It returns the value, its absolute expiration time, and true on success.
// On failure, it returns zero values and false; the value will not be cached.
//
// Example:
//
//	val, ok := cache.GetOrStore("user:123", func() (User, time.Time, bool) {
//	    user, err := db.GetUser(123)
//	    if err != nil {
//	        return User{}, time.Time{}, false
//	    }
//	    return user, time.Now().Add(5*time.Minute), true
//	})
type StoreCacheCallback[T any] func() (T, time.Time, bool)

// StoreCacheCallbackWithError is a function that fetches a value to cache on miss.
// It returns the value, a TTL duration, and an error. On success (nil error),
// the value is cached with expiration set to now + TTL. On failure, the error
// is returned and nothing is cached.
//
// Example:
//
//	val, err := cache.GetOrStoreWithError("user:123", func() (User, time.Duration, error) {
//	    user, err := db.GetUser(123)
//	    if err != nil {
//	        return User{}, 0, err
//	    }
//	    return user, 5*time.Minute, nil
//	})
type StoreCacheCallbackWithError[T any] func() (T, time.Duration, error)

// Deprecated: Use GetOrStore with StoreCacheCallback instead.
func (c *Cache[K, T]) TryGet(key K, storeOperation StoreCacheOperation[T]) (T, bool) {
	return c.GetOrStore(key, func() (T, time.Time, bool) {
		storeResult, successful := storeOperation()
		return storeResult.getValue(), storeResult.getExpiration(), successful
	})
}

// getOrStoreInternal is the shared singleflight logic for GetOrStore variants.
// The fetch callback returns (value, expiry, error) — nil error means success.
func (c *Cache[K, T]) getOrStoreInternal(key K, fetch func() (T, time.Time, error)) (T, error) {
	// Fast path: return cached value if present
	if cachedValue, isCached := c.Get(key); isCached {
		return cachedValue, nil
	}

	// Slow path: use singleflight to ensure only one goroutine fetches
	keyStr := fmt.Sprintf("%v", key)
	resultIface, err, _ := c.group.Do(keyStr, func() (any, error) {
		// Double-check cache inside singleflight - another goroutine may have populated it
		if cachedValue, isCached := c.Get(key); isCached {
			return cachedValue, nil
		}

		newValue, expiry, err := fetch()
		if err != nil {
			return nil, err
		}
		c.Set(key, newValue, expiry)
		return newValue, nil
	})

	if err != nil {
		return zeroVal[T](), err
	}
	return resultIface.(T), nil
}

// GetOrStore retrieves a cached value or fetches and stores it on miss.
// If the key exists and is not expired, returns the cached value immediately.
// Otherwise, calls storeOperation to fetch the value; on success, caches it
// with the returned expiration time.
//
// Concurrent calls for the same key share a single fetch operation via
// singleflight, preventing thundering herd on cache miss.
//
// Returns the value and true on success, or zero value and false if the
// store operation fails.
func (c *Cache[K, T]) GetOrStore(key K, storeOperation StoreCacheCallback[T]) (T, bool) {
	val, err := c.getOrStoreInternal(key, func() (T, time.Time, error) {
		v, exp, ok := storeOperation()
		if !ok {
			return zeroVal[T](), time.Time{}, errStoreFailed
		}
		return v, exp, nil
	})
	return val, err == nil
}

// errStoreFailed is a sentinel error for GetOrStore's bool-based callback.
var errStoreFailed = errors.New("store operation failed")

// GetOrStoreWithError retrieves a cached value or fetches and stores it on miss.
// If the key exists and is not expired, returns the cached value immediately.
// Otherwise, calls storeOperation to fetch the value; on success (nil error),
// caches it with expiration set to now + TTL duration, using the cache's
// configured time mode (local or UTC via WithUTC).
//
// Concurrent calls for the same key share a single fetch operation via
// singleflight, preventing thundering herd on cache miss.
//
// Returns the value and nil on success, or zero value and the error from
// storeOperation on failure.
func (c *Cache[K, T]) GetOrStoreWithError(key K, storeOperation StoreCacheCallbackWithError[T]) (T, error) {
	return c.getOrStoreInternal(key, func() (T, time.Time, error) {
		v, ttl, err := storeOperation()
		if err != nil {
			return zeroVal[T](), time.Time{}, err
		}
		return v, c.now().Add(ttl), nil
	})
}

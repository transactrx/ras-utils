package rascache

import (
	"sync"
	"testing"
	"time"
)

func TestNewCache(t *testing.T) {
	c := NewCache[string, int]()
	if c == nil {
		t.Fatal("NewCache returned nil")
	}
	if c.data == nil {
		t.Fatal("NewCache did not initialize data map")
	}
}

func TestCache_SetAndGet(t *testing.T) {
	c := NewCache[string, string]()

	c.Set("key1", "value1", time.Hour)

	val, ok := c.Get("key1")
	if !ok {
		t.Fatal("expected key1 to exist")
	}
	if val != "value1" {
		t.Errorf("expected value1, got %s", val)
	}
}

func TestCache_GetNonExistent(t *testing.T) {
	c := NewCache[string, string]()

	val, ok := c.Get("nonexistent")
	if ok {
		t.Fatal("expected nonexistent key to return false")
	}
	if val != "" {
		t.Errorf("expected zero value, got %s", val)
	}
}

func TestCache_Expiry(t *testing.T) {
	c := NewCache[string, string]()

	c.Set("expires", "soon", 10*time.Millisecond)

	// Should exist immediately
	val, ok := c.Get("expires")
	if !ok {
		t.Fatal("expected key to exist before expiry")
	}
	if val != "soon" {
		t.Errorf("expected 'soon', got %s", val)
	}

	// Wait for expiry
	time.Sleep(20 * time.Millisecond)

	// Should be expired now
	val, ok = c.Get("expires")
	if ok {
		t.Fatal("expected key to be expired")
	}
	if val != "" {
		t.Errorf("expected zero value after expiry, got %s", val)
	}
}

func TestCache_Delete(t *testing.T) {
	c := NewCache[string, string]()

	c.Set("key", "value", time.Hour)
	c.Delete("key")

	_, ok := c.Get("key")
	if ok {
		t.Fatal("expected key to be deleted")
	}
}

func TestCache_Clear(t *testing.T) {
	c := NewCache[string, string]()

	c.Set("key1", "value1", time.Hour)
	c.Set("key2", "value2", time.Hour)
	c.Set("key3", "value3", time.Hour)

	c.Clear()

	for _, key := range []string{"key1", "key2", "key3"} {
		if _, ok := c.Get(key); ok {
			t.Errorf("expected %s to be cleared", key)
		}
	}
}

func TestCache_GetOrStore_CacheHit(t *testing.T) {
	c := NewCache[string, string]()
	c.Set("key", "cached_value", time.Hour)

	callCount := 0
	val, ok := c.GetOrStore("key", func() (ICacheable[string], bool) {
		callCount++
		return NewCacheItem("fetched_value", time.Now().Add(time.Hour)), true
	})

	if !ok {
		t.Fatal("expected GetOrStore to succeed")
	}
	if val != "cached_value" {
		t.Errorf("expected cached_value, got %s", val)
	}
	if callCount != 0 {
		t.Errorf("expected store function not to be called on cache hit, called %d times", callCount)
	}
}

func TestCache_GetOrStore_CacheMiss(t *testing.T) {
	c := NewCache[string, string]()

	callCount := 0
	val, ok := c.GetOrStore("key", func() (ICacheable[string], bool) {
		callCount++
		return NewCacheItem("fetched_value", time.Now().Add(time.Hour)), true
	})

	if !ok {
		t.Fatal("expected GetOrStore to succeed")
	}
	if val != "fetched_value" {
		t.Errorf("expected fetched_value, got %s", val)
	}
	if callCount != 1 {
		t.Errorf("expected store function to be called once, called %d times", callCount)
	}

	// Verify value is now cached
	val, ok = c.Get("key")
	if !ok {
		t.Fatal("expected value to be cached after GetOrStore")
	}
	if val != "fetched_value" {
		t.Errorf("expected fetched_value in cache, got %s", val)
	}
}

func TestCache_GetOrStore_StoreOperationFails(t *testing.T) {
	c := NewCache[string, string]()

	val, ok := c.GetOrStore("key", func() (ICacheable[string], bool) {
		return nil, false
	})

	if ok {
		t.Fatal("expected GetOrStore to fail when store operation fails")
	}
	if val != "" {
		t.Errorf("expected zero value, got %s", val)
	}
}

// TestCache_TryGet_Deprecated tests the deprecated TryGet still works
func TestCache_TryGet_CacheHit(t *testing.T) {
	c := NewCache[string, string]()
	c.Set("key", "cached_value", time.Hour)

	callCount := 0
	val, ok := c.TryGet("key", func() (CacheItem[string], bool) {
		callCount++
		return CacheItem[string]{value: "fetched_value", expiry: time.Now().Add(time.Hour)}, true
	})

	if !ok {
		t.Fatal("expected TryGet to succeed")
	}
	if val != "cached_value" {
		t.Errorf("expected cached_value, got %s", val)
	}
	if callCount != 0 {
		t.Errorf("expected store function not to be called on cache hit, called %d times", callCount)
	}
}

func TestCache_TryGet_CacheMiss(t *testing.T) {
	c := NewCache[string, string]()

	callCount := 0
	val, ok := c.TryGet("key", func() (CacheItem[string], bool) {
		callCount++
		return CacheItem[string]{value: "fetched_value", expiry: time.Now().Add(time.Hour)}, true
	})

	if !ok {
		t.Fatal("expected TryGet to succeed")
	}
	if val != "fetched_value" {
		t.Errorf("expected fetched_value, got %s", val)
	}
	if callCount != 1 {
		t.Errorf("expected store function to be called once, called %d times", callCount)
	}

	// Verify value is now cached
	val, ok = c.Get("key")
	if !ok {
		t.Fatal("expected value to be cached after TryGet")
	}
	if val != "fetched_value" {
		t.Errorf("expected fetched_value in cache, got %s", val)
	}
}

func TestCache_TryGet_StoreOperationFails(t *testing.T) {
	c := NewCache[string, string]()

	val, ok := c.TryGet("key", func() (CacheItem[string], bool) {
		return CacheItem[string]{}, false
	})

	if ok {
		t.Fatal("expected TryGet to fail when store operation fails")
	}
	if val != "" {
		t.Errorf("expected zero value, got %s", val)
	}
}

func TestCache_ConcurrentAccess(t *testing.T) {
	c := NewCache[int, int]()
	var wg sync.WaitGroup

	// Spawn multiple goroutines doing concurrent reads and writes
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			c.Set(n, n*2, time.Hour)
			c.Get(n)
			c.Delete(n)
		}(i)
	}

	wg.Wait()
}

func TestNewCacheItem(t *testing.T) {
	expiry := time.Now().Add(time.Hour)
	item := NewCacheItem("test", expiry)

	if item.getValue() != "test" {
		t.Errorf("expected value 'test', got %s", item.getValue())
	}
	if !item.getExpiration().Equal(expiry) {
		t.Errorf("expected expiry %v, got %v", expiry, item.getExpiration())
	}
}

func TestNewCacheWithCleanup(t *testing.T) {
	c := NewCacheWithCleanup[string, string](50 * time.Millisecond)
	defer c.Stop()

	if c == nil {
		t.Fatal("NewCacheWithCleanup returned nil")
	}
	if c.data == nil {
		t.Fatal("NewCacheWithCleanup did not initialize data map")
	}
	if c.stopCh == nil {
		t.Fatal("NewCacheWithCleanup did not initialize stopCh")
	}
}

func TestCacheWithCleanup_BackgroundCleanup(t *testing.T) {
	c := NewCacheWithCleanup[string, string](50 * time.Millisecond)
	defer c.Stop()

	// Add items with short TTL
	c.Set("expires1", "value1", 20*time.Millisecond)
	c.Set("expires2", "value2", 20*time.Millisecond)
	c.Set("stays", "value3", time.Hour)

	// Verify items exist
	if _, ok := c.Get("expires1"); !ok {
		t.Fatal("expected expires1 to exist initially")
	}
	if _, ok := c.Get("expires2"); !ok {
		t.Fatal("expected expires2 to exist initially")
	}

	// Wait for expiry and cleanup cycle
	time.Sleep(100 * time.Millisecond)

	// Expired items should be cleaned up by background goroutine
	c.mu.RLock()
	_, exists1 := c.data["expires1"]
	_, exists2 := c.data["expires2"]
	_, exists3 := c.data["stays"]
	c.mu.RUnlock()

	if exists1 {
		t.Error("expected expires1 to be cleaned up")
	}
	if exists2 {
		t.Error("expected expires2 to be cleaned up")
	}
	if !exists3 {
		t.Error("expected stays to still exist")
	}
}

func TestCache_Stop(t *testing.T) {
	t.Run("stops cleanup goroutine", func(t *testing.T) {
		c := NewCacheWithCleanup[string, string](10 * time.Millisecond)

		// Add item that would expire
		c.Set("key", "value", 5*time.Millisecond)

		// Stop the cleanup
		c.Stop()

		// Wait past expiry and what would be cleanup time
		time.Sleep(30 * time.Millisecond)

		// The important thing is Stop() doesn't panic and the goroutine exits cleanly
		// Get will still return false due to expiry check on access
		_, ok := c.Get("key")
		if ok {
			t.Error("expected expired key to return false")
		}
	})

	t.Run("safe to call on regular cache", func(t *testing.T) {
		c := NewCache[string, string]()

		// Should not panic
		c.Stop()
	})

	t.Run("documents that calling Stop twice panics", func(t *testing.T) {
		c := NewCacheWithCleanup[string, string](50 * time.Millisecond)
		c.Stop()

		// Calling Stop() twice will panic due to closing closed channel
		// This documents the current behavior
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected Stop() to panic on second call")
			}
		}()
		c.Stop()
	})
}

func TestCache_DeleteExpired(t *testing.T) {
	c := NewCache[string, string]()

	// Add mix of expired and valid items
	c.mu.Lock()
	c.data["expired1"] = &CacheItem[string]{value: "v1", expiry: time.Now().Add(-time.Hour)}
	c.data["expired2"] = &CacheItem[string]{value: "v2", expiry: time.Now().Add(-time.Minute)}
	c.data["valid1"] = &CacheItem[string]{value: "v3", expiry: time.Now().Add(time.Hour)}
	c.data["valid2"] = &CacheItem[string]{value: "v4", expiry: time.Now().Add(time.Minute)}
	c.mu.Unlock()

	c.deleteExpired()

	c.mu.RLock()
	defer c.mu.RUnlock()

	if _, exists := c.data["expired1"]; exists {
		t.Error("expected expired1 to be deleted")
	}
	if _, exists := c.data["expired2"]; exists {
		t.Error("expected expired2 to be deleted")
	}
	if _, exists := c.data["valid1"]; !exists {
		t.Error("expected valid1 to still exist")
	}
	if _, exists := c.data["valid2"]; !exists {
		t.Error("expected valid2 to still exist")
	}
}

func TestCacheWithCleanup_ConcurrentAccess(t *testing.T) {
	c := NewCacheWithCleanup[int, int](10 * time.Millisecond)
	defer c.Stop()

	var wg sync.WaitGroup

	// Spawn multiple goroutines doing concurrent operations while cleanup runs
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			c.Set(n, n*2, 50*time.Millisecond)
			c.Get(n)
			c.Delete(n)
		}(i)
	}

	wg.Wait()
}

func TestNewCacheUTC(t *testing.T) {
	c := NewCacheUTC[string, int]()
	if c == nil {
		t.Fatal("NewCacheUTC returned nil")
	}
	if c.data == nil {
		t.Fatal("NewCacheUTC did not initialize data map")
	}
}

func TestCacheUTC_SetAndGet(t *testing.T) {
	c := NewCacheUTC[string, string]()

	c.Set("key1", "value1", time.Hour)

	val, ok := c.Get("key1")
	if !ok {
		t.Fatal("expected key1 to exist")
	}
	if val != "value1" {
		t.Errorf("expected value1, got %s", val)
	}
}

func TestCacheUTC_Expiry(t *testing.T) {
	c := NewCacheUTC[string, string]()

	c.Set("expires", "soon", 10*time.Millisecond)

	val, ok := c.Get("expires")
	if !ok {
		t.Fatal("expected key to exist before expiry")
	}
	if val != "soon" {
		t.Errorf("expected 'soon', got %s", val)
	}

	time.Sleep(20 * time.Millisecond)

	val, ok = c.Get("expires")
	if ok {
		t.Fatal("expected key to be expired")
	}
	if val != "" {
		t.Errorf("expected zero value after expiry, got %s", val)
	}
}

func TestNewCacheItemUTC(t *testing.T) {
	expiry := time.Now().UTC().Add(time.Hour)
	item := NewCacheItemUTC("test", expiry)

	if item.getValue() != "test" {
		t.Errorf("expected value 'test', got %s", item.getValue())
	}
	if !item.getExpiration().Equal(expiry) {
		t.Errorf("expected expiry %v, got %v", expiry, item.getExpiration())
	}
}

func TestCacheItemUTC_getTtl(t *testing.T) {
	expiry := time.Now().UTC().Add(time.Hour)
	item := NewCacheItemUTC("test", expiry)

	ttl := item.getTtl()
	if ttl < 59*time.Minute || ttl > time.Hour {
		t.Errorf("expected TTL around 1 hour, got %v", ttl)
	}
}

func TestCacheUTC_GetOrStore(t *testing.T) {
	c := NewCacheUTC[string, string]()

	val, ok := c.GetOrStore("key", func() (ICacheable[string], bool) {
		return NewCacheItemUTC("fetched", time.Now().UTC().Add(time.Hour)), true
	})

	if !ok {
		t.Fatal("expected GetOrStore to succeed")
	}
	if val != "fetched" {
		t.Errorf("expected 'fetched', got %s", val)
	}

	// Verify cached
	val, ok = c.Get("key")
	if !ok {
		t.Fatal("expected value to be cached")
	}
	if val != "fetched" {
		t.Errorf("expected 'fetched' in cache, got %s", val)
	}
}

func TestCacheUTC_TryGet(t *testing.T) {
	c := NewCacheUTC[string, string]()

	val, ok := c.TryGet("key", func() (CacheItem[string], bool) {
		return CacheItem[string]{value: "fetched", expiry: time.Now().UTC().Add(time.Hour)}, true
	})

	if !ok {
		t.Fatal("expected TryGet to succeed")
	}
	if val != "fetched" {
		t.Errorf("expected 'fetched', got %s", val)
	}

	// Verify cached
	val, ok = c.Get("key")
	if !ok {
		t.Fatal("expected value to be cached")
	}
	if val != "fetched" {
		t.Errorf("expected 'fetched' in cache, got %s", val)
	}
}

func TestNewCacheWithCleanupUTC(t *testing.T) {
	c := NewCacheWithCleanupUTC[string, string](50 * time.Millisecond)
	defer c.Stop()

	if c == nil {
		t.Fatal("NewCacheWithCleanupUTC returned nil")
	}
	if c.data == nil {
		t.Fatal("NewCacheWithCleanupUTC did not initialize data map")
	}
	if c.stopCh == nil {
		t.Fatal("NewCacheWithCleanupUTC did not initialize stopCh")
	}
}

func TestCacheWithCleanupUTC_BackgroundCleanup(t *testing.T) {
	c := NewCacheWithCleanupUTC[string, string](50 * time.Millisecond)
	defer c.Stop()

	c.Set("expires1", "value1", 20*time.Millisecond)
	c.Set("expires2", "value2", 20*time.Millisecond)
	c.Set("stays", "value3", time.Hour)

	if _, ok := c.Get("expires1"); !ok {
		t.Fatal("expected expires1 to exist initially")
	}
	if _, ok := c.Get("expires2"); !ok {
		t.Fatal("expected expires2 to exist initially")
	}

	time.Sleep(100 * time.Millisecond)

	c.mu.RLock()
	_, exists1 := c.data["expires1"]
	_, exists2 := c.data["expires2"]
	_, exists3 := c.data["stays"]
	c.mu.RUnlock()

	if exists1 {
		t.Error("expected expires1 to be cleaned up")
	}
	if exists2 {
		t.Error("expected expires2 to be cleaned up")
	}
	if !exists3 {
		t.Error("expected stays to still exist")
	}
}

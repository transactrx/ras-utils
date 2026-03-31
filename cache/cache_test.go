package cache

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

func TestCache_TryGet_CacheHit(t *testing.T) {
	c := NewCache[string, string]()
	c.Set("key", "cached_value", time.Hour)

	callCount := 0
	val, ok := c.TryGet("key", func() (CacheItem[string], bool) {
		callCount++
		return NewCacheItem("fetched_value", time.Now().Add(time.Hour)), true
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
		return NewCacheItem("fetched_value", time.Now().Add(time.Hour)), true
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

	if item.value != "test" {
		t.Errorf("expected value 'test', got %s", item.value)
	}
	if !item.expiry.Equal(expiry) {
		t.Errorf("expected expiry %v, got %v", expiry, item.expiry)
	}
}

package auth

import (
	"testing"
	"time"
)

type CacheValue struct {
	val int
}

func TestSetInsertsValue(t *testing.T) {
	cache := NewCache()
	data := &CacheValue{0}
	key := "key"
	cache.Set(key, data, 0)

	v, _, ok := cache.Get(key, false)
	if !ok || v.(*CacheValue) != data {
		t.Errorf("Cache has incorrect value: %v != %v", data, v)
	}
}

func TestGetValueWithMultipleTypes(t *testing.T) {
	cache := NewCache()
	data := &CacheValue{0}
	key := "key"
	cache.Set(key, data, 0)

	v, _, ok := cache.Get("key", false)
	if !ok || v.(*CacheValue) != data {
		t.Errorf("Cache has incorrect value for \"key\": %v != %v", data, v)
	}

	v, _, ok = cache.Get(string([]byte{'k', 'e', 'y'}), false)
	if !ok || v.(*CacheValue) != data {
		t.Errorf("Cache has incorrect value for []byte {'k','e','y'}: %v != %v", data, v)
	}
}

func TestSetWithOldKeyUpdatesValue(t *testing.T) {
	cache := NewCache()
	emptyValue := &CacheValue{0}
	key := "key1"
	cache.Set(key, emptyValue, 0)
	someValue := &CacheValue{20}
	cache.Set(key, someValue, 0)

	v, _, ok := cache.Get(key, false)
	if !ok || v.(*CacheValue) != someValue {
		t.Errorf("Cache has incorrect value: %v != %v", someValue, v)
	}
}

func TestGetNonExistent(t *testing.T) {
	cache := NewCache()

	if _, _, ok := cache.Get("crap", false); ok {
		t.Error("Cache returned a crap value after no inserts.")
	}
}

func TestDelete(t *testing.T) {
	cache := NewCache()
	value := &CacheValue{1}
	key := "key"

	if cache.Delete(key) {
		t.Error("Item unexpectedly already in cache.")
	}

	cache.Set(key, value, 0)

	if !cache.Delete(key) {
		t.Error("Expected item to be in cache.")
	}

	if sz := cache.Size(); sz != 0 {
		t.Errorf("cache.Size() = %v, expected 0", sz)
	}

	if _, _, ok := cache.Get(key, false); ok {
		t.Error("Cache returned a value after deletion.")
	}
}

func TestClear(t *testing.T) {
	cache := NewCache()
	value := &CacheValue{1}
	key := "key"

	cache.Set(key, value, 0)
	cache.Clear()

	if sz := cache.Size(); sz != 0 {
		t.Errorf("cache.Size() = %v, expected 0 after Clear()", sz)
	}
}

func TestCacheIsExpired(t *testing.T) {
	const expiredTime = 2

	cache := NewCache()

	cache.Set("key", &CacheValue{1}, expiredTime)

	// The least recently used one should have been evicted.
	if _, _, ok := cache.Get("key", false); !ok {
		t.Error("The element was evicted.")
	}

	time.Sleep(time.Second)

	// The least recently used one should have been evicted.
	if _, _, ok := cache.Get("key", false); !ok {
		t.Error("The element was evicted.")
	}

	time.Sleep(time.Second)

	// The least recently used one should have been evicted.
	if _, _, ok := cache.Get("key", false); ok {
		t.Error("The element was not evicted.")
	}

	// replace test
	cache.Set("key", &CacheValue{1}, expiredTime)

	// The least recently used one should have been evicted.
	if _, _, ok := cache.Get("key", false); !ok {
		t.Error("The element was evicted.")
	}

	time.Sleep(time.Second)

	// The least recently used one should have been evicted.
	if _, _, ok := cache.Get("key", false); !ok {
		t.Error("The element was evicted.")
	}

	time.Sleep(time.Second)

	// The least recently used one should have been evicted.
	if _, _, ok := cache.Get("key", false); ok {
		t.Error("The element was not evicted.")
	}
}

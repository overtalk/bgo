// Copyright 2012, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cache_test

import (
	"encoding/json"
	"testing"
	"time"

	lru "github.com/overtalk/bgo/3rdparty/lrucache"
)

type CacheValue struct {
	val int
}

func TestInitialState(t *testing.T) {
	cache := lru.NewLRUCache(5, 0)
	l, sz, c, _ := cache.Stats()
	if l != 0 {
		t.Errorf("length = %v, want 0", l)
	}
	if sz != 0 {
		t.Errorf("size = %v, want 0", sz)
	}
	if c != 5 {
		t.Errorf("capacity = %v, want 5", c)
	}
}

func TestSetInsertsValue(t *testing.T) {
	cache := lru.NewLRUCache(100, 0)
	data := &CacheValue{0}
	key := "key"
	cache.Set(key, data)

	v, ok := cache.Get(key)
	if !ok || v.(*CacheValue) != data {
		t.Errorf("Cache has incorrect value: %v != %v", data, v)
	}

	k := cache.Keys()
	if len(k) != 1 || k[0] != key {
		t.Errorf("Cache.Keys() returned incorrect values: %v", k)
	}
	values := cache.Items()
	if len(values) != 1 || values[0].Key != key {
		t.Errorf("Cache.Values() returned incorrect values: %v", values)
	}
}

func TestSetIfAbsent(t *testing.T) {
	cache := lru.NewLRUCache(100, 0)
	data := &CacheValue{0}
	key := "key"
	cache.SetIfAbsent(key, data)

	v, ok := cache.Get(key)
	if !ok || v.(*CacheValue) != data {
		t.Errorf("Cache has incorrect value: %v != %v", data, v)
	}

	v, ok = cache.SetIfAbsent(key, &CacheValue{1})
	if ok {
		t.Errorf("Cache has incorrect value: %v != %v", data, v)
	}

	v, ok = cache.Get(key)
	if !ok || v.(*CacheValue) != data {
		t.Errorf("Cache has incorrect value: %v != %v", data, v)
	}
}

func TestGetValueWithMultipleTypes(t *testing.T) {
	cache := lru.NewLRUCache(100, 0)
	data := &CacheValue{0}
	key := "key"
	cache.Set(key, data)

	v, ok := cache.Get("key")
	if !ok || v.(*CacheValue) != data {
		t.Errorf("Cache has incorrect value for \"key\": %v != %v", data, v)
	}

	v, ok = cache.Get(string([]byte{'k', 'e', 'y'}))
	if !ok || v.(*CacheValue) != data {
		t.Errorf("Cache has incorrect value for []byte {'k','e','y'}: %v != %v", data, v)
	}
}

func TestSetUpdatesSize(t *testing.T) {
	cache := lru.NewLRUCache(100, 0)
	emptyValue := &CacheValue{0}
	key := "key1"
	cache.Set(key, emptyValue)
	if sz := cache.Size(); sz != 1 {
		t.Errorf("cache.Size() = %v, expected 1", sz)
	}
	someValue := &CacheValue{20}
	key = "key2"
	cache.Set(key, someValue)
	if sz := cache.Size(); sz != 2 {
		t.Errorf("cache.Size() = %v, expected 2", sz)
	}
}

func TestSetWithOldKeyUpdatesValue(t *testing.T) {
	cache := lru.NewLRUCache(100, 0)
	emptyValue := &CacheValue{0}
	key := "key1"
	cache.Set(key, emptyValue)
	someValue := &CacheValue{20}
	cache.Set(key, someValue)

	v, ok := cache.Get(key)
	if !ok || v.(*CacheValue) != someValue {
		t.Errorf("Cache has incorrect value: %v != %v", someValue, v)
	}
}

func TestSetWithOldKeyUpdatesSize(t *testing.T) {
	cache := lru.NewLRUCache(100, 0)
	emptyValue := &CacheValue{0}
	key := "key1"
	cache.Set(key, emptyValue)

	if sz := cache.Size(); sz != 1 {
		t.Errorf("cache.Size() = %v, expected 1", sz)
	}

	someValue := &CacheValue{20}
	cache.Set(key, someValue)
	if sz := cache.Size(); sz != 1 {
		t.Errorf("cache.Size() = %v, expected 1", sz)
	}
}

func TestGetNonExistent(t *testing.T) {
	cache := lru.NewLRUCache(100, 0)

	if cache.IsExisted("crap") {
		t.Error("Cache returned a crap value after no inserts.")
	}
	if _, ok := cache.Get("crap"); ok {
		t.Error("Cache returned a crap value after no inserts.")
	}
}

func TestPeek(t *testing.T) {
	cache := lru.NewLRUCache(2, 0)
	val1 := &CacheValue{1}
	cache.Set("key1", val1)
	val2 := &CacheValue{1}
	cache.Set("key2", val2)
	// Make key1 the most recent.
	cache.Get("key1")
	// Peek key2.
	if v, ok := cache.Peek("key2"); ok && v.(*CacheValue) != val2 {
		t.Errorf("key2 received: %v, want %v", v, val2)
	}
	// Push key2 out
	cache.Set("key3", &CacheValue{1})
	if v, ok := cache.Peek("key2"); ok {
		t.Errorf("key2 received: %v, want absent", v)
	}
}

func TestDelete(t *testing.T) {
	cache := lru.NewLRUCache(100, 0)
	value := &CacheValue{1}
	key := "key"

	if cache.Delete(key) {
		t.Error("Item unexpectedly already in cache.")
	}

	cache.Set(key, value)

	if !cache.Delete(key) {
		t.Error("Expected item to be in cache.")
	}

	if sz := cache.Size(); sz != 0 {
		t.Errorf("cache.Size() = %v, expected 0", sz)
	}

	if _, ok := cache.Get(key); ok {
		t.Error("Cache returned a value after deletion.")
	}
}

func TestClear(t *testing.T) {
	cache := lru.NewLRUCache(100, 0)
	value := &CacheValue{1}
	key := "key"

	cache.Set(key, value)
	cache.Clear()

	if sz := cache.Size(); sz != 0 {
		t.Errorf("cache.Size() = %v, expected 0 after Clear()", sz)
	}
}

func TestCapacityIsObeyed(t *testing.T) {
	size := int64(3)
	cache := lru.NewLRUCache(100, 0)
	cache.SetCapacity(size)
	value := &CacheValue{1}

	// Insert up to the cache's capacity.
	cache.Set("key1", value)
	cache.Set("key2", value)
	cache.Set("key3", value)
	if sz := cache.Size(); sz != size {
		t.Errorf("cache.Size() = %v, expected %v", sz, size)
	}
	// Insert one more; something should be evicted to make room.
	cache.Set("key4", value)
	if sz := cache.Size(); sz != size {
		t.Errorf("post-evict cache.Size() = %v, expected %v", sz, size)
	}

	// Check json stats
	data := cache.StatsJSON()
	m := make(map[string]interface{})
	if err := json.Unmarshal([]byte(data), &m); err != nil {
		t.Errorf("cache.StatsJSON() returned bad json data: %v %v", data, err)
	}
	if m["Size"].(float64) != float64(size) {
		t.Errorf("cache.StatsJSON() returned bad size: %v", m)
	}

	// Check various other stats
	if l := cache.Length(); l != size {
		t.Errorf("cache.StatsJSON() returned bad length: %v", l)
	}
	if s := cache.Size(); s != size {
		t.Errorf("cache.StatsJSON() returned bad size: %v", s)
	}
	if c := cache.Capacity(); c != size {
		t.Errorf("cache.StatsJSON() returned bad length: %v", c)
	}

	// checks StatsJSON on nil
	cache = nil
	if s := cache.StatsJSON(); s != "{}" {
		t.Errorf("cache.StatsJSON() on nil object returned %v", s)
	}
}

func TestLRUIsEvicted(t *testing.T) {
	size := int64(3)
	cache := lru.NewLRUCache(size, 0)

	cache.Set("key1", &CacheValue{1})
	cache.Set("key2", &CacheValue{1})
	cache.Set("key3", &CacheValue{1})
	// lru: [key3, key2, key1]

	// Look up the elements. This will rearrange the LRU ordering.
	cache.Get("key3")
	beforeKey2 := time.Now()
	cache.Get("key2")
	afterKey2 := time.Now()
	cache.Get("key1")
	// lru: [key1, key2, key3]

	cache.Set("key0", &CacheValue{1})
	// lru: [key0, key1, key2]

	// The least recently used one should have been evicted.
	if _, ok := cache.Get("key3"); ok {
		t.Error("Least recently used element was not evicted.")
	}

	// Check oldest
	if o := cache.Oldest(); o.Before(beforeKey2) || o.After(afterKey2) {
		t.Errorf("cache.Oldest returned an unexpected value: got %v, expected a value between %v and %v", o, beforeKey2, afterKey2)
	}

	// Check newest
	if o := cache.Newest(); o.Before(afterKey2) {
		t.Errorf("cache.Newest returned an unexpected value: got %v, expected a value after %v", o, afterKey2)
	}
}

func TestLRUIsExpired(t *testing.T) {
	cache := lru.NewLRUCache(10, time.Second)

	cache.Set("key1", &CacheValue{1})
	cache.SetWithTTL("key3", &CacheValue{1}, 300*time.Millisecond)

	// The least recently used one should have been evicted.
	if _, ok := cache.Get("key1"); !ok {
		t.Error("Last element was evicted.")
	}

	time.Sleep(500 * time.Millisecond)

	cache.Set("key2", &CacheValue{1})

	// key1 expired, key2 will replace it.
	// so the size is still 1
	if cache.Size() != 2 {
		t.Error("Last element was not evicted.")
	}
	// The least recently used one should have been evicted.
	if _, ok := cache.Get("key1"); !ok {
		t.Error("Last element was missing.")
	}
	if _, ok := cache.Get("key2"); !ok {
		t.Error("New element was missing.")
	}
}

func TestLRUExpiredGet(t *testing.T) {
	cache := lru.NewLRUCache(10, time.Second)

	cache.SetWithTTL("key1", &CacheValue{1}, 300*time.Millisecond)
	cache.Set("key2", &CacheValue{1})

	// The least recently used one should have been evicted.
	if _, ok := cache.Get("key1"); !ok {
		t.Error("Last element was evicted.")
	}

	time.Sleep(500 * time.Millisecond)

	// The least recently used one should have been evicted.
	if _, ok := cache.Get("key1"); ok {
		t.Error("Last element was not evicted.")
	}
	if _, ok := cache.Get("key2"); !ok {
		t.Error("Last element was evicted.")
	}
}

func TestLRUExpiredSet(t *testing.T) {
	cache := lru.NewLRUCache(10, time.Second)

	cache.Set("key1", &CacheValue{1})
	if _, ok := cache.Get("key1"); !ok {
		t.Error("The element-key1 was evicted.")
	}

	cache.SetExpired("key1")
	if _, ok := cache.Get("key1"); ok {
		t.Error("The element-key1 was not evicted.")
	}
}

func TestLRUSize(t *testing.T) {
	cache := lru.NewLRUCache(2, 0)

	cache.Set("key1", &CacheValue{1})
	cache.Set("key2", &CacheValue{1})
	cache.Set("key3", &CacheValue{1})

	if sz := cache.Size(); sz != 2 {
		t.Errorf("cache.Size() = %v, expected 2", sz)
	}
}

func TestLRURandomGet(t *testing.T) {
	cache := lru.NewLRUCache(10, time.Second)

	cache.Set("key1", &CacheValue{1})
	cache.Set("key2", &CacheValue{2})
	cache.Set("key3", &CacheValue{3})
	cache.Set("key4", &CacheValue{4})
	cache.Set("key5", &CacheValue{5})

	t.Logf("RandomItem:%v", cache.RandomItems(2))
}
// Copyright 2012, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package cache implements a LRU cache.
//
// The implementation borrows heavily from SmallLRUCache
// (originally by Nathan Schrenk). The object maintains a doubly-linked list of
// elements. When an element is accessed, it is promoted to the head of the
// list. When space is needed, the element at the tail of the list
// (the least recently used element) is evicted.
package cache

import (
	"container/list"
	"fmt"
	"sync"
	"time"
)

// LRUCache is a typical LRU cache implementation.  If the cache
// reaches the capacity, the least recently used item is deleted from
// the cache. Note the capacity is not the number of items, but the
// total sum of the Size() of each item.
type LRUCache struct {
	mu sync.Mutex

	// list & table of *entry objects
	list  *list.List
	table map[string]*list.Element

	// Our current size. Obviously a gross simplification and
	// low-grade approximation.
	size int64

	// How much we are limiting the cache to.
	capacity int64

	// how long the last cache will be expired
	// <= 0: won't expired
	ttl time.Duration
}

// NewLRUCache creates a new empty cache with the given capacity.
func NewLRUCache(capacity int64, ttl time.Duration) *LRUCache {
	return &LRUCache{
		list:     list.New(),
		table:    make(map[string]*list.Element),
		capacity: capacity,
		ttl:      ttl,
	}
}

// Get returns a value from the cache, and marks the entry as most
// recently used.
func (lru *LRUCache) Get(key string) (v interface{}, ok bool) {
	lru.mu.Lock()
	element := lru.table[key]
	if element == nil {
		lru.mu.Unlock()
		return nil, false
	}
	e := element.Value.(*entry)
	now := time.Now()
	if e.expired(now) {
		lru.mu.Unlock()
		return nil, false
	}
	e.assessTime = now
	lru.list.MoveToFront(element)
	lru.mu.Unlock()
	return e.value, true
}

// Peek returns a value from the cache without changing the LRU order.
func (lru *LRUCache) Peek(key string) (v interface{}, ok bool) {
	lru.mu.Lock()
	element := lru.table[key]
	if element == nil {
		lru.mu.Unlock()
		return nil, false
	}
	e := element.Value.(*entry)
	if e.expired(time.Now()) {
		lru.mu.Unlock()
		return nil, false
	}
	lru.mu.Unlock()
	return e.value, true
}

// IsExisted check whether a value is existed in the cache and not expired.
func (lru *LRUCache) IsExisted(key string) (existed bool) {
	lru.mu.Lock()
	if element := lru.table[key]; element != nil {
		e := element.Value.(*entry)
		existed = !e.expired(time.Now())
	}
	lru.mu.Unlock()
	return
}

// Set sets a value in the cache with a TTL.
func (lru *LRUCache) Set(key string, value interface{}) {
	lru.mu.Lock()
	if !lru.replaceOldItem(key, value, lru.ttl) {
		lru.addNew(key, value, lru.ttl)
	}
	lru.mu.Unlock()
}

// SetWithTTL sets a value in the cache with a TTL.
func (lru *LRUCache) SetWithTTL(
	key string, value interface{}, ttl time.Duration) {
	lru.mu.Lock()
	if !lru.replaceOldItem(key, value, ttl) {
		lru.addNew(key, value, ttl)
	}
	lru.mu.Unlock()
}

// SetIfAbsent will set the value in the cache if not present.
// If the value exists in the cache, we don't set it.
func (lru *LRUCache) SetIfAbsent(
	key string, value interface{}) (interface{}, bool) {
	lru.mu.Lock()
	if element := lru.table[key]; element != nil {
		// check whether it's expired
		e := element.Value.(*entry)
		now := time.Now()
		if !e.expired(now) {
			e.ttl = lru.ttl
			e.assessTime = now
			lru.list.MoveToFront(element)
			lru.mu.Unlock()
			return e.value, false
		}
	}
	if !lru.replaceOldItem(key, value, lru.ttl) {
		lru.addNew(key, value, lru.ttl)
	}
	lru.mu.Unlock()
	return value, true
}

// SetExpired will set an entry expired from the cache
// and returns if the entry existed.
func (lru *LRUCache) SetExpired(key string) (ok bool) {
	lru.mu.Lock()
	if element := lru.table[key]; element != nil {
		e := element.Value.(*entry)
		e.ttl = 1
		ok = true
	}
	lru.mu.Unlock()
	return
}

// Delete removes an entry from the cache, and returns if the entry existed.
func (lru *LRUCache) Delete(key string) (deleted bool) {
	lru.mu.Lock()
	if element := lru.table[key]; element != nil {
		lru.list.Remove(element)
		delete(lru.table, key)
		lru.size--
		deleted = true
	}
	lru.mu.Unlock()
	return
}

// Clear will clear the entire cache.
func (lru *LRUCache) Clear() {
	lru.mu.Lock()
	lru.list.Init()
	lru.table = make(map[string]*list.Element)
	lru.size = 0
	lru.mu.Unlock()
}

// SetCapacity will set the capacity of the cache. If the capacity is
// smaller, and the current cache size exceed that capacity, the cache
// will be shrank.
func (lru *LRUCache) SetCapacity(capacity int64) {
	lru.mu.Lock()
	lru.capacity = capacity
	lru.checkCapacity()
	lru.mu.Unlock()
}

func (lru *LRUCache) checkCapacity() {
	for lru.size > lru.capacity {
		delElem := lru.list.Back()
		delValue := delElem.Value.(*entry)
		lru.list.Remove(delElem)
		delete(lru.table, delValue.key)
		lru.size--
	}
}

// Stats returns a few stats on the cache.
func (lru *LRUCache) Stats() (length, size, capacity int64, oldest time.Time) {
	lru.mu.Lock()
	if lastElem := lru.list.Back(); lastElem != nil {
		oldest = lastElem.Value.(*entry).assessTime
	}
	length = int64(lru.list.Len())
	lru.mu.Unlock()
	return length, lru.size, lru.capacity, oldest
}

// StatsJSON returns stats as a JSON object in a string.
func (lru *LRUCache) StatsJSON() string {
	if lru != nil {
		l, s, c, o := lru.Stats()
		return fmt.Sprintf(
			`{"Length":%v, "Size":%v, "Capacity":%v, "OldestAccess":"%v"}`,
			l, s, c, o,
		)
	}
	return "{}"
}

// Length returns how many elements are in the cache
func (lru *LRUCache) Length() int64 {
	lru.mu.Lock()
	val := lru.list.Len()
	lru.mu.Unlock()
	return int64(val)
}

// Size returns the sum of the objects' Size() method.
func (lru *LRUCache) Size() int64 {
	lru.mu.Lock()
	val := lru.size
	lru.mu.Unlock()
	return val
}

// Capacity returns the cache maximum capacity.
func (lru *LRUCache) Capacity() int64 {
	lru.mu.Lock()
	val := lru.capacity
	lru.mu.Unlock()
	return val
}

// FreeSize returns the cache's free capacity.
func (lru *LRUCache) FreeSize() int64 {
	lru.mu.Lock()
	val := lru.capacity - lru.size
	lru.mu.Unlock()
	return val
}

// Oldest returns the insertion time of the oldest element in the cache,
// or a IsZero() time if cache is empty.
func (lru *LRUCache) Oldest() (oldest time.Time) {
	lru.mu.Lock()
	if lastElem := lru.list.Back(); lastElem != nil {
		oldest = lastElem.Value.(*entry).assessTime
	}
	lru.mu.Unlock()
	return
}

// Newest returns the insertion time of the newest element in the cache,
// or a IsZero() time if cache is empty.
func (lru *LRUCache) Newest() (newest time.Time) {
	lru.mu.Lock()
	if firstElem := lru.list.Front(); firstElem != nil {
		newest = firstElem.Value.(*entry).assessTime
	}
	lru.mu.Unlock()
	return
}

// Keys returns all the keys for the cache, ordered from most recently
// used to last recently used.
func (lru *LRUCache) Keys() []string {
	lru.mu.Lock()
	keys := make([]string, 0, lru.list.Len())
	for e := lru.list.Front(); e != nil; e = e.Next() {
		keys = append(keys, e.Value.(*entry).key)
	}
	lru.mu.Unlock()
	return keys
}

// Items returns all the values for the cache, ordered from most recently
// used to last recently used.
func (lru *LRUCache) Items() []Item {
	lru.mu.Lock()
	items := make([]Item, 0, lru.list.Len())
	for e := lru.list.Front(); e != nil; e = e.Next() {
		v := e.Value.(*entry)
		items = append(items, Item{Key: v.key, Value: v.value})
	}
	lru.mu.Unlock()
	return items
}

// RandomItems returns all random values for the cache
func (lru *LRUCache) RandomItems(maxCount int) []Item {
	items := make([]Item, 0, maxCount)
	lru.mu.Lock()
	defer lru.mu.Unlock()
	for key, element := range lru.table {
		if len(items) < maxCount {
			e := element.Value.(*entry)
			items = append(items, Item{Key: key, Value: e.value})
		} else {
			break
		}
	}
	return items
}

func (lru *LRUCache) addNew(key string, value interface{}, ttl time.Duration) {
	lru.table[key] = lru.list.PushFront(&entry{key, value, time.Now(), ttl})
	lru.size++
	lru.checkCapacity()
}

func (lru *LRUCache) replaceOldItem(key string, value interface{}, ttl time.Duration) bool {
	element := lru.table[key]
	// if existed, just replace its value.
	if element != nil {
		e := element.Value.(*entry)
		e.value = value
		e.ttl = ttl
		e.assessTime = time.Now()
		lru.list.MoveToFront(element)
		return true
	}
	// replace expired item or spare one.
	element = lru.list.Back()
	if element == nil {
		return false
	}
	e := element.Value.(*entry)
	now := time.Now()
	if lru.size < lru.capacity && !e.expired(now) {
		return false
	}
	delete(lru.table, e.key)
	e.key = key
	e.value = value
	e.ttl = ttl
	e.assessTime = now
	lru.table[key] = element
	lru.list.MoveToFront(element)
	return true
}

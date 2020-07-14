// Package auth cache is forked from https://bitbucket.org/ckrissun/lrucache
package memcache

import (
	"container/list"
	"sync"
	"time"
)

// Cache is a typical cache implementation.
// When doing get/set/delete action, expired element will be deleted
type Cache struct {
	mu sync.Mutex

	// list & table of *entry objects
	list  *list.List
	table map[string]*list.Element
}

type entry struct {
	key        string
	value      interface{}
	expireTime int64
	accessTime time.Time
}

func (e *entry) expired(now time.Time) bool {
	return (e.expireTime > 0 && now.Unix() >= e.expireTime)
}

// NewCache creates a new empty cache with the given ttl.
func NewCache() *Cache {
	return &Cache{
		list:  list.New(),
		table: make(map[string]*list.Element),
	}
}

// Get returns a value from the cache, and marks the entry as most
// recently used.
func (c *Cache) Get(key string, accessed bool) (interface{}, time.Duration, bool) {
	c.mu.Lock()
	element := c.table[key]
	if element == nil {
		c.mu.Unlock()
		return nil, 0, false
	}
	e := element.Value.(*entry)
	now := time.Now()
	if e.expired(now) {
		c.mu.Unlock()
		return nil, 0, false
	}
	var delaTime time.Duration
	if accessed {
		delaTime = now.Sub(e.accessTime)
		e.accessTime = now
	}
	c.mu.Unlock()
	return e.value, delaTime, true
}

// Set sets a value in the cache.
func (c *Cache) Set(key string, value interface{}, ttl int64) {
	c.mu.Lock()
	if !c.replaceOldItem(key, value, ttl) {
		c.addNew(key, value, ttl)
	}
	c.mu.Unlock()
}

// SetExpiration sets a value's expireTime in the cache.
// ttl: < 0(no expiration)
//      = 0(expiratioin ASAP)
//      > 0(a short while)
func (c *Cache) SetExpiration(key string, ttl int64) {
	c.mu.Lock()
	element := c.table[key]
	if element == nil {
		c.mu.Unlock()
		return
	}
	e := element.Value.(*entry)
	if ttl >= 0 {
		e.expireTime = time.Now().Unix() + ttl
	} else {
		e.expireTime = 0
	}
	c.mu.Unlock()
}

// Delete removes an entry from the cache, and returns if the entry existed.
func (c *Cache) Delete(key string) bool {
	c.mu.Lock()
	element := c.table[key]
	if element == nil {
		c.mu.Unlock()
		return false
	}
	c.list.Remove(element)
	delete(c.table, key)
	c.mu.Unlock()
	return true
}

// Clear will clear the entire cache.
func (c *Cache) Clear() {
	c.mu.Lock()
	c.list.Init()
	c.table = make(map[string]*list.Element)
	c.mu.Unlock()
}

// Size returns how many elements are in the cache
func (c *Cache) Size() int64 {
	c.mu.Lock()
	size := c.list.Len()
	c.mu.Unlock()
	return int64(size)
}

func (c *Cache) updateInplace(element *list.Element, value interface{}, ttl int64) {
	e := element.Value.(*entry)
	e.value = value
	if ttl > 0 {
		e.expireTime = time.Now().Unix() + ttl
	} else {
		e.expireTime = 0
	}
	// new value moved to the front of the list
	c.list.MoveToFront(element)
}

func (c *Cache) addNew(key string, value interface{}, ttl int64) {
	now := time.Now()
	newEntry := &entry{key, value, 0, now}
	if ttl > 0 {
		newEntry.expireTime = now.Unix() + ttl
	}
	// each new entry is put in front of the list
	element := c.list.PushFront(newEntry)
	c.table[key] = element
}

func (c *Cache) replaceOldItem(key string, value interface{}, ttl int64) bool {
	element := c.table[key]
	if element != nil {
		c.updateInplace(element, value, ttl)
		return true
	}
	if ttl <= 0 {
		return false
	}
	element = c.list.Back()
	if element == nil {
		return false
	}
	e := element.Value.(*entry)
	now := time.Now()
	if !e.expired(now) {
		return false
	}
	if e.key != key {
		delete(c.table, e.key)
		e.key = key
		c.table[key] = element
	}
	e.value = value
	e.expireTime = now.Unix() + ttl
	c.list.MoveToFront(element)
	return true
}

func (c *Cache) ExpiredKeys(duration time.Duration) []string {
	var ret []string
	for {
		element := c.list.Back()

		// if empty
		if element == nil {
			break
		}

		// if not expired
		e := element.Value.(*entry)
		t := time.Now().Add(duration)
		if !e.expired(t) {
			break
		}

		ret = append(ret, e.key)
	}
	return ret
}

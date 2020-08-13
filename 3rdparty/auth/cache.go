package auth

import (
	"container/list"
	"sync"
	"time"
)

// entry defines the item stored in cache
type entry struct {
	key   string
	value interface{}
	// ts for expired time, expireTime == 0 means never expired
	expireTime int64
	// last access time
	accessTime time.Time
}

func (this *entry) expired(now time.Time) bool {
	return this.expireTime > 0 && now.Unix() >= this.expireTime
}

type Cache struct {
	sync.Mutex

	// list & table of *entry objects
	list  *list.List
	table map[string]*list.Element
}

func NewCache() *Cache {
	return &Cache{
		list:  list.New(),
		table: make(map[string]*list.Element),
	}
}

// Get returns the value of certain key, and marks the entry as most recently used
// second return value defines the time duration with last access
func (this *Cache) Get(key string, accessed bool) (interface{}, time.Duration, bool) {
	this.Lock()
	defer this.Unlock()

	element, flag := this.table[key]
	if !flag {
		return nil, 0, false
	}

	e := element.Value.(*entry)
	now := time.Now()

	// check weather the key is expired
	if e.expired(now) {
		return nil, 0, false
	}

	var delaTime time.Duration
	if accessed {
		delaTime = now.Sub(e.accessTime)
		e.accessTime = now
	}
	return e.value, delaTime, true
}

// Set sets a value in the cache.
func (this *Cache) Set(key string, value interface{}, ttl int64) {
	this.Lock()
	defer this.Unlock()

	if !this.replaceOldItem(key, value, ttl) {
		this.addNewItem(key, value, ttl)
	}
}

func (this *Cache) SetExpiration(key string, ttl int64) {
	this.Lock()
	defer this.Unlock()

	// check weather the key is exist
	element, flag := this.table[key]
	if !flag {
		return
	}

	// refresh the expired ts
	e := element.Value.(*entry)
	if ttl > 0 {
		e.expireTime = time.Now().Unix() + ttl
	} else {
		e.expireTime = 0
	}
}

func (this *Cache) Delete(key string) bool {
	this.Lock()
	defer this.Unlock()

	// check weather the key is exist
	element, flag := this.table[key]
	if !flag {
		return false
	}

	// del the element
	this.list.Remove(element)
	delete(this.table, key)
	return true
}

func (this *Cache) Clear() {
	this.Lock()
	defer this.Unlock()

	this.list.Init()
	this.table = make(map[string]*list.Element)
}

func (this *Cache) Size() int {
	this.Lock()
	this.Unlock()
	return this.list.Len()
}

// addNewItem add a new element to cache
// make sure the key is new, or the old value will be override
func (this *Cache) addNewItem(key string, value interface{}, ttl int64) {
	now := time.Now()
	newEntry := &entry{
		key:        key,
		value:      value,
		expireTime: 0,
		accessTime: now,
	}

	if ttl > 0 {
		newEntry.expireTime = now.Unix() + ttl
	}

	this.table[key] = this.list.PushFront(newEntry)
}

// not thread-safe
// before call this func, you should call lock()
// this func will not change the number of the cached items
// if the key exist, refresh the value & expired ts & move to the front of list
// if not, and the last element is expired, the new value will replace the last element
func (this *Cache) replaceOldItem(key string, value interface{}, ttl int64) bool {
	element, flag := this.table[key]
	if flag {
		// if the key is existed
		// update the value & move to the front
		this.updateElement(element, value, ttl)
		return true
	}

	if ttl <= 0 {
		return false
	}

	element = this.list.Back()
	if element == nil {
		return false
	}

	e := element.Value.(*entry)
	now := time.Now()
	if !e.expired(now) {
		// this func will not change the number of cached items
		// if the key is not expired
		return false
	}

	// the last element is expired
	// replace the last element with the new value
	if e.key != key {
		delete(this.table, e.key)
		e.key = key
		this.table[key] = element
	}
	e.value = value
	e.expireTime = now.Unix() + ttl
	this.list.MoveToFront(element)
	return true
}

// updateElement will update the value & expired ts & move to the front of the list
func (this *Cache) updateElement(element *list.Element, value interface{}, ttl int64) {
	e := element.Value.(*entry)

	// set value & expired ts
	e.value = value
	if ttl > 0 {
		e.expireTime = time.Now().Unix() + ttl
	} else {
		e.expireTime = 0
	}

	// new value moved to the front of the list
	this.list.MoveToFront(element)
}

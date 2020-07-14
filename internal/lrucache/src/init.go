package clrucache

import (
	"container/list"
	"sync"
	"time"

	"github.com/overtalk/bgo/core"
	"github.com/overtalk/bgo/internal/lrucache"
)

const (
	defaultCapacity = 10000
	defaultTTL      = time.Minute * 20
)

func init() {
	var module ilrucache.ILrucacheModule = new(CLrucacheModule)
	core.GetCore().RegisterModule(ilrucache.ModuleName, module)
}

// LRUCache is a typical LRU cache implementation.  If the cache
// reaches the capacity, the least recently used item is deleted from
// the cache. Note the capacity is not the number of items, but the
// total sum of the Size() of each item.
type CLrucacheModule struct {
	core.Module

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

func (lru *CLrucacheModule) Init() error {
	// creates a new empty cache with the given capacity.
	lru.list = list.New()
	lru.table = make(map[string]*list.Element)
	lru.capacity = defaultCapacity
	lru.ttl = defaultTTL
	return nil
}

package ilrucache

import (
	"time"

	"github.com/overtalk/bgo/core"
)

const ModuleName = "internal.lrucache"

// Item is what is stored in the cache
type Item struct {
	Key   string
	Value interface{}
}

type ILrucacheModule interface {
	core.IModule

	SetDefaultTTL(ttl time.Duration)
	Get(key string) (v interface{}, ok bool)
	Peek(key string) (v interface{}, ok bool)
	IsExisted(key string) (existed bool)
	Set(key string, value interface{})
	SetWithTTL(key string, value interface{}, ttl time.Duration)
	SetIfAbsent(key string, value interface{}) (interface{}, bool)
	SetExpired(key string) (ok bool)
	Delete(key string) (deleted bool)
	Clear()
	SetCapacity(capacity int64)
	Stats() (length, size, capacity int64, oldest time.Time)
	StatsJSON() string
	Length() int64
	Size() int64
	Capacity() int64
	FreeSize() int64
	Oldest() (oldest time.Time)
	Newest() (newest time.Time)
	Keys() []string
	Items() []Item
	RandomItems(maxCount int) []Item
}

package cache

import "time"

// Item is what is stored in the cache
type Item struct {
	Key   string
	Value interface{}
}

// entry is the struct to store
type entry struct {
	key        string
	value      interface{}
	assessTime time.Time
	ttl        time.Duration
}

// if the entry is accessed
// the expired will be carry forward
func (e *entry) expired(t time.Time) bool {
	// ttl <= 0: won't expired
	return e.ttl > 0 && t.Sub(e.assessTime) >= e.ttl
}

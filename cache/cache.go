// Package cache implements a concurrent TTL cache.
package cache

import (
	"sync"
	"time"
)

// Fetches data for a given key, and returns the data as a byte slice and the TTL as a time.Duration.
// A Getter may be called concurrently with different keys, but will NOT be called concurrently with the same key.
type Getter func(key string) (data []byte, ttl time.Duration)

type item struct {
	sync.RWMutex
	data []byte
}

type Cache struct {
	sync.RWMutex
	data   map[string]*item
	getter Getter
}

// Return a new Cache linked to a Getter.
// A Cache will look up data in-memory, or fetch it using its Getter.
// In-memory cached items will be purged after the TTL expires.
func New(g Getter) Cache {
	c := Cache{}
	c.data = make(map[string]*item)
	c.getter = g
	return c
}

// Get looks up the cached value, and if necessary, performs a remote fetch to get the new value and TTL.
// If another fetch for the same key is in progress, we just wait for that to complete and read the result.
// The boolean return value represents whether this Get was a cache hit.
func (c *Cache) Get(key string) ([]byte, bool) {
	c.RLock()
	val, ok := c.data[key]
	c.RUnlock()
	if ok {
		val.RLock()
		defer val.RUnlock()
		return val.data, ok
	}
	c.Lock()
	val = new(item)
	c.data[key] = val
	c.Unlock()
	val.Lock()
	defer val.Unlock()
	data := val.data
	if data != nil {
		return data, true
	}
	data, ttl := c.getter(key)
	val.data = data
	c.RemoveAfter(ttl, key)
	return data, false
}

func (c *Cache) Remove(key string) {
	c.Lock()
	delete(c.data, key)
	c.Unlock()
}

func (c *Cache) RemoveAfter(ttl time.Duration, key string) {
	time.AfterFunc(ttl, func() {
		c.Remove(key)
	})
}

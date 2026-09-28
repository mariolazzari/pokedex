package pokecache

import (
	"sync"
	"time"
)

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

type Cache struct {
	values map[string]cacheEntry
	mutex  sync.Mutex
}

func NewCache(ttl time.Duration) *Cache {
	cache := &Cache{
		values: make(map[string]cacheEntry),
	}

	go readLoop(cache, ttl)

	return cache
}

func (c *Cache) Add(key string, val []byte) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.values[key] = cacheEntry{
		createdAt: time.Now(),
		val:       val,
	}
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	v, ok := c.values[key]
	if !ok {
		return nil, false
	}

	return v.val, true
}

func readLoop(c *Cache, ttl time.Duration) {
	ticker := time.NewTicker(ttl)
	defer ticker.Stop()

	for range ticker.C {
		c.mutex.Lock()

		for key, entry := range c.values {
			if time.Since(entry.createdAt) >= ttl {
				delete(c.values, key)
			}
		}
		c.mutex.Unlock()
	}

}

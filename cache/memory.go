package cache

import "sync"

// MemoryCache is an implementation of Cache that stores responses in an in-memory map.
type MemoryCache struct {
	mu    sync.RWMutex
	items map[string][]byte
}

// Get returns the []byte representation of the response and true if present, false if not
func (c *MemoryCache) Get(key string) (resp []byte, ok bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	resp, ok = c.items[key]
	return resp, ok
}

// Set saves response resp to the cache with key
func (c *MemoryCache) Set(key string, resp []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = resp
}

// Delete removes key from the cache
func (c *MemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// NewMemoryCache returns a new Cache that will store items in an in-memory map
func NewMemoryCache() *MemoryCache {
	c := &MemoryCache{items: map[string][]byte{}}
	return c
}

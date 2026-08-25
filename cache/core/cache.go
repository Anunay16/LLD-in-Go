package core

import (
	"sync"

	"github.com/lld/cache/eviction"
	"github.com/lld/cache/storage"
)

// Cache defines the public interface for the cache.
type Cache interface {
	Get(key string) (any, bool)
	Put(key string, value any)
	Delete(key string) bool
	Len() int
	Capacity() int
	Clear()
}

// SimpleCache is a thread-safe cache orchestrating storage and an eviction policy.
type SimpleCache struct {
	mu       sync.RWMutex
	capacity int
	storage  storage.Storage
	eviction eviction.Policy
}

// NewCache creates a new SimpleCache instance with the given capacity, eviction policy, and storage backend.
func NewCache(capacity int, policy eviction.Policy, store storage.Storage) *SimpleCache {
	if store == nil {
		store = storage.NewMemoryStorage()
	}
	return &SimpleCache{
		capacity: capacity,
		storage:  store,
		eviction: policy,
	}
}

// Get retrieves a value from the cache and updates its eviction access metadata.
func (c *SimpleCache) Get(key string) (any, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	val, exists := c.storage.Get(key)
	if !exists {
		return nil, false
	}

	c.eviction.KeyAccessed(key)
	return val, true
}

// Put inserts or updates a key-value pair, evicting an entry if capacity is reached.
func (c *SimpleCache) Put(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If key already exists, update and notify access
	if _, exists := c.storage.Get(key); exists {
		c.storage.Set(key, value)
		c.eviction.KeyAccessed(key)
		return
	}

	// If cache is at full capacity, evict an item first
	if c.storage.Len() >= c.capacity && c.capacity > 0 {
		victim, err := c.eviction.Evict()
		if err == nil {
			c.storage.Delete(victim)
		}
	}

	// Insert new entry
	c.storage.Set(key, value)
	c.eviction.KeyAdded(key)
}

// Delete removes a key-value pair and notifies the eviction policy.
func (c *SimpleCache) Delete(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if deleted := c.storage.Delete(key); deleted {
		c.eviction.KeyRemoved(key)
		return true
	}
	return false
}

// Len returns the current number of items in the cache.
func (c *SimpleCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.storage.Len()
}

// Capacity returns the maximum capacity of the cache.
func (c *SimpleCache) Capacity() int {
	return c.capacity
}

// Clear removes all items from the cache.
func (c *SimpleCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.storage.Clear()
}

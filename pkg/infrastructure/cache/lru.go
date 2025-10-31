package cache

import (
	"container/list"
	"sync"
	"time"
)

// LRUCache implements an LRU (Least Recently Used) cache with TTL support
type LRUCache struct {
	mu         sync.RWMutex
	capacity   int
	defaultTTL time.Duration

	items map[string]*list.Element
	list  *list.List

	// Stats
	hits      uint64
	misses    uint64
	evictions uint64
}

// lruEntry holds the cache entry data
type lruEntry struct {
	key       string
	value     interface{}
	expiresAt time.Time
}

// NewLRUCache creates a new LRU cache
func NewLRUCache(capacity int, defaultTTL time.Duration) *LRUCache {
	return &LRUCache{
		capacity:   capacity,
		defaultTTL: defaultTTL,
		items:      make(map[string]*list.Element),
		list:       list.New(),
	}
}

// Get retrieves a value from the cache
func (c *LRUCache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, exists := c.items[key]
	if !exists {
		c.misses++
		return nil, false
	}

	entry := elem.Value.(*lruEntry)

	// Check expiration
	if !entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt) {
		c.removeElement(elem)
		c.misses++
		return nil, false
	}

	// Move to front (most recently used)
	c.list.MoveToFront(elem)
	c.hits++

	return entry.value, true
}

// Set stores a value in the cache with default TTL
func (c *LRUCache) Set(key string, value interface{}) {
	c.SetWithTTL(key, value, c.defaultTTL)
}

// SetWithTTL stores a value with a specific TTL
func (c *LRUCache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	expiresAt := time.Time{}
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}

	// Update existing entry
	if elem, exists := c.items[key]; exists {
		c.list.MoveToFront(elem)
		entry := elem.Value.(*lruEntry)
		entry.value = value
		entry.expiresAt = expiresAt
		return
	}

	// Add new entry
	entry := &lruEntry{
		key:       key,
		value:     value,
		expiresAt: expiresAt,
	}

	elem := c.list.PushFront(entry)
	c.items[key] = elem

	// Evict if over capacity
	if c.list.Len() > c.capacity {
		c.evictOldest()
	}
}

// Delete removes a value from the cache
func (c *LRUCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, exists := c.items[key]; exists {
		c.removeElement(elem)
	}
}

// Clear removes all values from the cache
func (c *LRUCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*list.Element)
	c.list = list.New()
	c.hits = 0
	c.misses = 0
	c.evictions = 0
}

// Size returns the number of items in the cache
func (c *LRUCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.list.Len()
}

// Stats returns cache statistics
func (c *LRUCache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := c.hits + c.misses
	hitRate := float64(0)
	if total > 0 {
		hitRate = float64(c.hits) / float64(total)
	}

	return CacheStats{
		Hits:      c.hits,
		Misses:    c.misses,
		Evictions: c.evictions,
		Size:      c.list.Len(),
		Capacity:  c.capacity,
		HitRate:   hitRate,
	}
}

// evictOldest removes the least recently used item
func (c *LRUCache) evictOldest() {
	elem := c.list.Back()
	if elem != nil {
		c.removeElement(elem)
		c.evictions++
	}
}

// removeElement removes an element from both the list and map
func (c *LRUCache) removeElement(elem *list.Element) {
	c.list.Remove(elem)
	entry := elem.Value.(*lruEntry)
	delete(c.items, entry.key)
}

// CleanupExpired removes expired entries
func (c *LRUCache) CleanupExpired() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	count := 0
	now := time.Now()

	// Walk through list and remove expired items
	for elem := c.list.Front(); elem != nil; {
		next := elem.Next()
		entry := elem.Value.(*lruEntry)

		if !entry.expiresAt.IsZero() && now.After(entry.expiresAt) {
			c.removeElement(elem)
			count++
		}

		elem = next
	}

	return count
}

// StartCleanupRoutine starts a background goroutine to clean up expired entries
func (c *LRUCache) StartCleanupRoutine(interval time.Duration) chan struct{} {
	stopCh := make(chan struct{})

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				c.CleanupExpired()
			case <-stopCh:
				return
			}
		}
	}()

	return stopCh
}

// Keys returns all keys in the cache (most recent first)
func (c *LRUCache) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]string, 0, c.list.Len())
	for elem := c.list.Front(); elem != nil; elem = elem.Next() {
		entry := elem.Value.(*lruEntry)
		keys = append(keys, entry.key)
	}

	return keys
}

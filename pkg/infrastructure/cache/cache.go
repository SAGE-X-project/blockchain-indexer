package cache

import (
	"sync"
	"time"
)

// Cache is a generic interface for caching implementations
type Cache interface {
	// Get retrieves a value from the cache
	Get(key string) (interface{}, bool)

	// Set stores a value in the cache
	Set(key string, value interface{})

	// SetWithTTL stores a value with a specific TTL
	SetWithTTL(key string, value interface{}, ttl time.Duration)

	// Delete removes a value from the cache
	Delete(key string)

	// Clear removes all values from the cache
	Clear()

	// Size returns the number of items in the cache
	Size() int

	// Stats returns cache statistics
	Stats() CacheStats
}

// CacheStats holds cache statistics
type CacheStats struct {
	Hits       uint64
	Misses     uint64
	Evictions  uint64
	Size       int
	Capacity   int
	HitRate    float64
	MemoryUsed uint64
}

// Entry represents a cache entry with TTL
type Entry struct {
	Value      interface{}
	ExpiresAt  time.Time
	AccessedAt time.Time
}

// IsExpired checks if the entry has expired
func (e *Entry) IsExpired() bool {
	if e.ExpiresAt.IsZero() {
		return false
	}
	return time.Now().After(e.ExpiresAt)
}

// MemoryCache is a simple in-memory cache with TTL support
type MemoryCache struct {
	mu         sync.RWMutex
	entries    map[string]*Entry
	defaultTTL time.Duration
	maxSize    int

	// Stats
	hits      uint64
	misses    uint64
	evictions uint64
}

// NewMemoryCache creates a new memory cache
func NewMemoryCache(maxSize int, defaultTTL time.Duration) *MemoryCache {
	return &MemoryCache{
		entries:    make(map[string]*Entry),
		defaultTTL: defaultTTL,
		maxSize:    maxSize,
	}
}

// Get retrieves a value from the cache
func (c *MemoryCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		c.misses++
		return nil, false
	}

	if entry.IsExpired() {
		c.misses++
		// Don't delete here to avoid write lock, will be cleaned up later
		return nil, false
	}

	entry.AccessedAt = time.Now()
	c.hits++
	return entry.Value, true
}

// Set stores a value in the cache with default TTL
func (c *MemoryCache) Set(key string, value interface{}) {
	c.SetWithTTL(key, value, c.defaultTTL)
}

// SetWithTTL stores a value with a specific TTL
func (c *MemoryCache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Evict if at capacity and key doesn't exist
	if _, exists := c.entries[key]; !exists && len(c.entries) >= c.maxSize {
		c.evictOldest()
	}

	expiresAt := time.Time{}
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}

	c.entries[key] = &Entry{
		Value:      value,
		ExpiresAt:  expiresAt,
		AccessedAt: time.Now(),
	}
}

// Delete removes a value from the cache
func (c *MemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, key)
}

// Clear removes all values from the cache
func (c *MemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*Entry)
	c.hits = 0
	c.misses = 0
	c.evictions = 0
}

// Size returns the number of items in the cache
func (c *MemoryCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.entries)
}

// Stats returns cache statistics
func (c *MemoryCache) Stats() CacheStats {
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
		Size:      len(c.entries),
		Capacity:  c.maxSize,
		HitRate:   hitRate,
	}
}

// evictOldest evicts the oldest accessed entry
func (c *MemoryCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range c.entries {
		if oldestTime.IsZero() || entry.AccessedAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.AccessedAt
		}
	}

	if oldestKey != "" {
		delete(c.entries, oldestKey)
		c.evictions++
	}
}

// CleanupExpired removes expired entries
func (c *MemoryCache) CleanupExpired() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	count := 0
	for key, entry := range c.entries {
		if entry.IsExpired() {
			delete(c.entries, key)
			count++
		}
	}

	return count
}

// StartCleanupRoutine starts a background goroutine to clean up expired entries
func (c *MemoryCache) StartCleanupRoutine(interval time.Duration) chan struct{} {
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

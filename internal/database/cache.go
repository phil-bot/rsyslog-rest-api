package database

import (
	"fmt"
	"sync"
	"time"
)

const metaCacheTTL = 60 * time.Second

type cacheEntry struct {
	value     interface{}
	expiresAt time.Time
}

// MetaCache is a simple TTL cache for meta query results.
// Keyed by column + WHERE clause + args to support filtered meta queries.
type MetaCache struct {
	mu      sync.Mutex
	entries map[string]cacheEntry
}

// NewMetaCache creates a new empty MetaCache.
func NewMetaCache() *MetaCache {
	return &MetaCache{entries: make(map[string]cacheEntry)}
}

// Get retrieves a cached value. Returns (nil, false) if not found or expired.
func (c *MetaCache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.entries[key]
	if !ok || time.Now().After(e.expiresAt) {
		delete(c.entries, key)
		return nil, false
	}
	return e.value, true
}

// Set stores a value with the default TTL.
func (c *MetaCache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = cacheEntry{value: value, expiresAt: time.Now().Add(metaCacheTTL)}
}

// CacheKey generates a deterministic cache key from column, WHERE clause, and args.
func CacheKey(column, whereClause string, args []interface{}) string {
	return fmt.Sprintf("%s|%s|%v", column, whereClause, args)
}

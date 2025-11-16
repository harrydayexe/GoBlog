package content

import (
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/harrydayexe/GoBlog/pkg/models"
)

// Cache wraps Ristretto cache for blog posts
type Cache struct {
	cache *ristretto.Cache
	ttl   time.Duration
}

// NewCache creates a new cache instance
func NewCache(maxSizeMB int64, ttl time.Duration) (*Cache, error) {
	// Configure Ristretto
	// NumCounters: 10x the max number of items we expect (assuming 1KB per post)
	// MaxCost: max size in bytes
	// BufferItems: 64 is the recommended value
	numCounters := int64(maxSizeMB * 1024 * 10) // 10x items for better hit rate
	maxCost := maxSizeMB * 1024 * 1024          // Convert MB to bytes

	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: numCounters,
		MaxCost:     maxCost,
		BufferItems: 64,
	})
	if err != nil {
		return nil, err
	}

	return &Cache{
		cache: cache,
		ttl:   ttl,
	}, nil
}

// Get retrieves a post from cache by slug
func (c *Cache) Get(slug string) *models.Post {
	if c.cache == nil {
		return nil
	}

	val, found := c.cache.Get(slug)
	if !found {
		return nil
	}

	post, ok := val.(*models.Post)
	if !ok {
		return nil
	}

	return post
}

// Set stores a post in cache
func (c *Cache) Set(slug string, post *models.Post) {
	if c.cache == nil {
		return
	}

	// Estimate cost as the size of the content + metadata
	// This is a rough estimate - in production you might want to be more precise
	cost := int64(len(post.Content) + len(post.Title) + len(post.Description))

	c.cache.SetWithTTL(slug, post, cost, c.ttl)
}

// Clear clears all cached posts
func (c *Cache) Clear() {
	if c.cache == nil {
		return
	}

	c.cache.Clear()
}

// Close closes the cache
func (c *Cache) Close() {
	if c.cache != nil {
		c.cache.Close()
	}
}

// Stats returns cache statistics
func (c *Cache) Stats() *CacheStats {
	if c.cache == nil {
		return &CacheStats{}
	}

	metrics := c.cache.Metrics
	return &CacheStats{
		Hits:      metrics.Hits(),
		Misses:    metrics.Misses(),
		KeysAdded: metrics.KeysAdded(),
		HitRatio:  metrics.Ratio(),
	}
}

// CacheStats represents cache statistics
type CacheStats struct {
	Hits      uint64
	Misses    uint64
	KeysAdded uint64
	HitRatio  float64
}

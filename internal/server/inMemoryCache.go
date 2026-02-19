package server

import (
	"context"
	"sync"

	"github.com/harrydayexe/GoBlog/v2/pkg/generator"
)

// InMemoryCache is an in-memory implementation of the Cache interface.
// It stores a single GeneratedBlog instance in memory using a mutex
// for concurrent access protection.
//
// All methods are safe for concurrent use by multiple goroutines.
// The cache has no expiration or size limits - content remains cached
// until explicitly cleared or replaced.
type InMemoryCache struct {
	mu   sync.RWMutex
	blog *generator.GeneratedBlog
}

// NewInMemoryCache creates a new InMemoryCache instance.
// The cache is initially empty.
func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{}
}

// Get retrieves the cached blog content.
// It returns nil if no content is currently cached.
// Get is safe for concurrent use and uses a read lock to allow
// multiple concurrent reads.
func (m *InMemoryCache) Get(ctx context.Context) (*generator.GeneratedBlog, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.blog, nil
}

// Set stores the provided blog content in the cache.
// Any previously cached content is replaced.
// Set is safe for concurrent use and uses a write lock to ensure
// exclusive access during updates.
func (m *InMemoryCache) Set(ctx context.Context, blog *generator.GeneratedBlog) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.blog = blog
	return nil
}

// Clear removes all cached content.
// After calling Clear, Get will return nil until Set is called again.
// Clear is safe for concurrent use and uses a write lock to ensure
// exclusive access during the operation.
func (m *InMemoryCache) Clear(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.blog = nil
	return nil
}

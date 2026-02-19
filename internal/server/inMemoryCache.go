package server

import (
	"context"
	"sync"

	"github.com/harrydayexe/GoBlog/v2/pkg/generator"
)

type InMemoryCache struct {
	mu   sync.RWMutex
	blog *generator.GeneratedBlog
}

func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{}
}

func (m *InMemoryCache) Get(ctx context.Context) (*generator.GeneratedBlog, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.blog, nil
}

func (m *InMemoryCache) Set(ctx context.Context, blog *generator.GeneratedBlog) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.blog = blog
	return nil
}

func (m *InMemoryCache) Clear(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.blog = nil
	return nil
}

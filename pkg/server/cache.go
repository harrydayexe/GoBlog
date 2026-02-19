package server

import (
	"context"

	"github.com/harrydayexe/GoBlog/v2/pkg/generator"
)

// Cache provides an interface for caching generated blog content.
// Implementations must be safe for concurrent use by multiple goroutines.
//
// The cache stores a single GeneratedBlog instance and provides methods
// to get, set, and clear the cached content. This is typically used to
// avoid regenerating blog content on every request.
type Cache interface {
	// Get retrieves the cached blog content.
	// It returns nil if no content is cached.
	// Get must be safe to call concurrently with other Cache methods.
	Get(ctx context.Context) (*generator.GeneratedBlog, error)

	// Set stores the provided blog content in the cache.
	// Any previously cached content is replaced.
	// Set must be safe to call concurrently with other Cache methods.
	Set(ctx context.Context, blog *generator.GeneratedBlog) error

	// Clear removes all cached content.
	// After calling Clear, Get will return nil until Set is called again.
	// Clear must be safe to call concurrently with other Cache methods.
	Clear(ctx context.Context) error
}

package server

import (
	"time"
)

// Options configures the blog server
type Options struct {
	// ContentPath is the path to markdown posts (required)
	ContentPath string

	// Cache settings
	EnableCache bool
	CacheMaxMB  int64
	CacheTTL    time.Duration

	// Search settings
	EnableSearch    bool
	SearchIndexPath string
	RebuildIndex    bool

	// Blog settings
	PostsPerPage int

	// Logging
	Verbose bool
}

// DefaultOptions returns default server options
func DefaultOptions() Options {
	return Options{
		ContentPath:     "./posts",
		EnableCache:     true,
		CacheMaxMB:      100,
		CacheTTL:        60 * time.Minute,
		EnableSearch:    true,
		SearchIndexPath: "./blog.bleve",
		RebuildIndex:    false,
		PostsPerPage:    10,
		Verbose:         false,
	}
}

// Validate checks if options are valid
func (o Options) Validate() error {
	if o.ContentPath == "" {
		return ErrInvalidContentPath
	}

	if o.EnableCache && o.CacheMaxMB < 1 {
		return ErrInvalidCacheSize
	}

	if o.PostsPerPage < 1 {
		return ErrInvalidPostsPerPage
	}

	return nil
}

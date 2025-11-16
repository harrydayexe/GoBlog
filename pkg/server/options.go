package server

import (
	"time"

	"github.com/caarlos0/env/v11"
)

// Options configures the blog server
type Options struct {
	// ContentPath is the path to markdown posts (required)
	ContentPath string `env:"GOBLOG_CONTENT_PATH" envDefault:"./posts"`

	// Cache settings
	EnableCache bool          `env:"GOBLOG_CACHE_ENABLED" envDefault:"true"`
	CacheMaxMB  int64         `env:"GOBLOG_CACHE_MAX_MB" envDefault:"100"`
	CacheTTL    time.Duration `env:"GOBLOG_CACHE_TTL" envDefault:"60m"`

	// Search settings
	EnableSearch    bool   `env:"GOBLOG_SEARCH_ENABLED" envDefault:"true"`
	SearchIndexPath string `env:"GOBLOG_SEARCH_INDEX_PATH" envDefault:"./blog.bleve"`
	RebuildIndex    bool   `env:"GOBLOG_REBUILD_INDEX" envDefault:"false"`

	// Blog settings
	PostsPerPage int `env:"GOBLOG_POSTS_PER_PAGE" envDefault:"10"`

	// Logging
	Verbose bool `env:"GOBLOG_VERBOSE" envDefault:"false"`
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

// LoadFromEnv loads options from environment variables
// Falls back to defaults for any unset variables
func LoadFromEnv() (Options, error) {
	opts := Options{}
	if err := env.Parse(&opts); err != nil {
		return Options{}, err
	}
	return opts, nil
}

// MustLoadFromEnv loads options from environment variables
// Panics if parsing fails
func MustLoadFromEnv() Options {
	opts, err := LoadFromEnv()
	if err != nil {
		panic(err)
	}
	return opts
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

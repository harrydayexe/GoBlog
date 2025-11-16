package server

import (
	"net/http"
	"time"

	"github.com/caarlos0/env/v11"
)

// EventHandlers contains callbacks for blog events
type EventHandlers struct {
	// OnPostView is called when a post is viewed
	OnPostView func(slug string, r *http.Request)

	// OnSearch is called when a search is performed
	OnSearch func(query string, resultCount int, r *http.Request)

	// OnError is called when an error occurs
	OnError func(err error, r *http.Request)

	// OnReload is called when content is reloaded
	OnReload func(postCount int)
}

// CORSConfig configures Cross-Origin Resource Sharing
type CORSConfig struct {
	// AllowedOrigins is a list of allowed origins (e.g., ["https://example.com"])
	// Use ["*"] to allow all origins (not recommended for production)
	AllowedOrigins []string

	// AllowedMethods is a list of allowed HTTP methods
	// Defaults to ["GET", "POST", "OPTIONS"]
	AllowedMethods []string

	// AllowedHeaders is a list of allowed headers
	// Defaults to ["Content-Type"]
	AllowedHeaders []string

	// AllowCredentials indicates whether credentials are allowed
	AllowCredentials bool

	// MaxAge indicates how long (in seconds) the results of a preflight request can be cached
	MaxAge int
}

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

	// File watching
	WatchFiles bool `env:"GOBLOG_WATCH_FILES" envDefault:"false"`

	// Logging
	Verbose bool `env:"GOBLOG_VERBOSE" envDefault:"false"`

	// Middleware - custom middleware to wrap all blog routes
	Middleware []func(http.Handler) http.Handler

	// CustomCSS - path to custom CSS file or inline CSS to inject
	CustomCSS string

	// CustomTemplates - custom template overrides (not implemented yet, reserved for future)
	CustomTemplates map[string]string

	// EventHandlers - callbacks for blog events
	EventHandlers *EventHandlers

	// CORS - Cross-Origin Resource Sharing configuration
	CORS *CORSConfig

	// Feed configuration
	FeedTitle       string // Title for RSS/Atom feeds
	FeedDescription string // Description for RSS/Atom feeds
	FeedAuthor      string // Author for RSS/Atom feeds
	FeedAuthorEmail string // Author email for Atom feeds
	SiteURL         string // Base URL of the site (e.g., "https://example.com")
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
		WatchFiles:      false,
		Verbose:         false,
		Middleware:      []func(http.Handler) http.Handler{},
		EventHandlers:   nil,
		CORS:            nil,
		FeedTitle:       "My Blog",
		FeedDescription: "A blog powered by GoBlog",
		FeedAuthor:      "",
		FeedAuthorEmail: "",
		SiteURL:         "http://localhost:8080",
	}
}

// Fluent API methods for configuring Options

// WithContentPath sets the content path
func (o Options) WithContentPath(path string) Options {
	o.ContentPath = path
	return o
}

// WithCache configures caching
func (o Options) WithCache(enabled bool, maxMB int64, ttl time.Duration) Options {
	o.EnableCache = enabled
	o.CacheMaxMB = maxMB
	o.CacheTTL = ttl
	return o
}

// WithSearch configures search
func (o Options) WithSearch(enabled bool, indexPath string, rebuild bool) Options {
	o.EnableSearch = enabled
	o.SearchIndexPath = indexPath
	o.RebuildIndex = rebuild
	return o
}

// WithPostsPerPage sets posts per page
func (o Options) WithPostsPerPage(count int) Options {
	o.PostsPerPage = count
	return o
}

// WithFileWatching enables or disables file watching
func (o Options) WithFileWatching(enabled bool) Options {
	o.WatchFiles = enabled
	return o
}

// WithVerbose enables or disables verbose logging
func (o Options) WithVerbose(enabled bool) Options {
	o.Verbose = enabled
	return o
}

// WithMiddleware adds middleware to the options
func (o Options) WithMiddleware(middleware ...func(http.Handler) http.Handler) Options {
	o.Middleware = append(o.Middleware, middleware...)
	return o
}

// WithCustomCSS sets custom CSS
func (o Options) WithCustomCSS(css string) Options {
	o.CustomCSS = css
	return o
}

// WithEventHandlers sets event handlers
func (o Options) WithEventHandlers(handlers *EventHandlers) Options {
	o.EventHandlers = handlers
	return o
}

// WithCORS sets CORS configuration
func (o Options) WithCORS(cors *CORSConfig) Options {
	o.CORS = cors
	return o
}

// WithFeed sets feed configuration
func (o Options) WithFeed(title, description, author, authorEmail, siteURL string) Options {
	o.FeedTitle = title
	o.FeedDescription = description
	o.FeedAuthor = author
	o.FeedAuthorEmail = authorEmail
	o.SiteURL = siteURL
	return o
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

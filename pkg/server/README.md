# GoBlogServ SDK

The GoBlogServ SDK allows you to embed blog functionality into your own Go applications.

## Installation

```bash
go get github.com/harrydayexe/GoBlog/pkg/server
```

## Quick Start

```go
package main

import (
    "log"
    "net/http"
    "github.com/harrydayexe/GoBlog/pkg/server"
)

func main() {
    // Create server with default options
    blogServer, err := server.New(server.DefaultOptions())
    if err != nil {
        log.Fatal(err)
    }
    defer blogServer.Close()

    // Create router and attach blog routes
    mux := http.NewServeMux()
    blogServer.AttachRoutes(mux, "/blog")

    // Start server
    http.ListenAndServe(":8080", mux)
}
```

## Environment Variables

GoBlogServ supports configuration via environment variables, making it perfect for Docker and cloud deployments.

### Loading from Environment

```go
// Load configuration from environment variables
opts, err := server.LoadFromEnv()
if err != nil {
    log.Fatal(err)
}

// Or panic if configuration is invalid
opts := server.MustLoadFromEnv()

blogServer, _ := server.New(opts)
```

### Available Variables

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `GOBLOG_CONTENT_PATH` | string | `./posts` | Path to markdown posts |
| `GOBLOG_CACHE_ENABLED` | bool | `true` | Enable Ristretto cache |
| `GOBLOG_CACHE_MAX_MB` | int64 | `100` | Max cache size in MB |
| `GOBLOG_CACHE_TTL` | duration | `60m` | Cache TTL (e.g., 30m, 1h) |
| `GOBLOG_SEARCH_ENABLED` | bool | `true` | Enable Bleve search |
| `GOBLOG_SEARCH_INDEX_PATH` | string | `./blog.bleve` | Search index path |
| `GOBLOG_REBUILD_INDEX` | bool | `false` | Rebuild index on startup |
| `GOBLOG_POSTS_PER_PAGE` | int | `10` | Posts per page |
| `GOBLOG_VERBOSE` | bool | `false` | Verbose logging |

### Docker Example

```dockerfile
FROM your-app-base

ENV GOBLOG_CONTENT_PATH=/app/posts \
    GOBLOG_CACHE_MAX_MB=200 \
    GOBLOG_VERBOSE=true

CMD ["/app/yourapp"]
```

```yaml
# docker-compose.yml
services:
  blog:
    image: your-blog-app
    environment:
      GOBLOG_CONTENT_PATH: /app/posts
      GOBLOG_CACHE_ENABLED: "true"
      GOBLOG_VERBOSE: "true"
    volumes:
      - ./posts:/app/posts:ro
```

## API Reference

### Creating a Server

```go
func New(opts Options) (*Server, error)
```

Creates a new blog server with the given options.

**Example:**
```go
opts := server.DefaultOptions()
opts.ContentPath = "./posts"
opts.EnableCache = true
opts.EnableSearch = true

blogServer, err := server.New(opts)
```

### Attaching Routes

```go
func (s *Server) AttachRoutes(mux *http.ServeMux, basePath string)
```

Attaches blog routes to your HTTP mux at the specified base path.

**Routes created:**
- `GET {basePath}` - Blog index with pagination and search
- `GET {basePath}/posts/{slug}` - Individual post view
- `GET {basePath}/tags/{tag}` - Posts filtered by tag
- `GET {basePath}/search` - Search endpoint (returns HTMX partial)

**Example:**
```go
mux := http.NewServeMux()
blogServer.AttachRoutes(mux, "/blog")
// Now blog is available at /blog/*
```

### Server Methods

#### Reload

```go
func (s *Server) Reload() error
```

Reloads all content from disk and rebuilds the search index.

**Example:**
```go
// Reload after files change
if err := blogServer.Reload(); err != nil {
    log.Printf("Failed to reload: %v", err)
}
```

#### Close

```go
func (s *Server) Close() error
```

Closes the server and cleans up resources (cache, search index).

**Example:**
```go
defer blogServer.Close()
```

#### Stats

```go
func (s *Server) Stats() Stats
```

Returns server statistics.

**Example:**
```go
stats := blogServer.Stats()
fmt.Printf("Total posts: %d\n", stats.TotalPosts)
fmt.Printf("Cache hit ratio: %.2f\n", stats.CacheHitRatio)
```

#### GetLoader

```go
func (s *Server) GetLoader() *content.Loader
```

Returns the content loader for advanced usage.

**Example:**
```go
loader := blogServer.GetLoader()
posts := loader.GetAll()
posts := loader.GetByTag("golang")
```

#### GetIndex

```go
func (s *Server) GetIndex() *search.Index
```

Returns the search index for advanced usage.

**Example:**
```go
index := blogServer.GetIndex()
results, _ := index.Search("web development", 10)
```

## Options

### DefaultOptions

```go
func DefaultOptions() Options
```

Returns sensible defaults:
- `ContentPath`: `"./posts"`
- `EnableCache`: `true`
- `CacheMaxMB`: `100`
- `CacheTTL`: `60 * time.Minute`
- `EnableSearch`: `true`
- `SearchIndexPath`: `"./blog.bleve"`
- `RebuildIndex`: `false`
- `PostsPerPage`: `10`
- `Verbose`: `false`

### Options Struct

```go
type Options struct {
    ContentPath      string        // Path to markdown posts (required)
    EnableCache      bool          // Enable Ristretto cache
    CacheMaxMB       int64         // Max cache size in MB
    CacheTTL         time.Duration // Cache TTL
    EnableSearch     bool          // Enable Bleve search
    SearchIndexPath  string        // Path to Bleve index
    RebuildIndex     bool          // Rebuild index on startup
    PostsPerPage     int           // Posts per page
    Verbose          bool          // Enable verbose logging
}
```

## Stats Struct

```go
type Stats struct {
    TotalPosts    int       // Total number of published posts
    AllTags       []string  // All unique tags
    CacheHits     uint64    // Cache hit count
    CacheMisses   uint64    // Cache miss count
    CacheHitRatio float64   // Hit ratio (0.0 - 1.0)
    IndexedDocs   uint64    // Number of indexed documents
}
```

## Error Handling

The SDK defines several error types:

```go
var (
    ErrInvalidContentPath  = errors.New("invalid content path")
    ErrInvalidCacheSize    = errors.New("cache size must be at least 1MB")
    ErrInvalidPostsPerPage = errors.New("posts per page must be at least 1")
    ErrServerNotStarted    = errors.New("server not started")
)
```

**Example:**
```go
blogServer, err := server.New(opts)
if errors.Is(err, server.ErrInvalidContentPath) {
    log.Fatal("Content path does not exist")
}
```

## Advanced Examples

### Multiple Blogs

Host multiple independent blogs:

```go
techBlog, _ := server.New(server.Options{
    ContentPath: "./tech-posts",
    EnableSearch: true,
})
techBlog.AttachRoutes(mux, "/tech")

personalBlog, _ := server.New(server.Options{
    ContentPath: "./personal-posts",
    EnableCache: true,
})
personalBlog.AttachRoutes(mux, "/personal")
```

### Custom Statistics Endpoint

```go
mux.HandleFunc("GET /api/blog/stats", func(w http.ResponseWriter, r *http.Request) {
    stats := blogServer.Stats()
    json.NewEncoder(w).Encode(stats)
})
```

### Hot Reload

Watch for file changes and reload:

```go
// Using fsnotify or similar
watcher.Add(opts.ContentPath)
for {
    select {
    case <-watcher.Events:
        log.Println("Reloading blog content...")
        if err := blogServer.Reload(); err != nil {
            log.Printf("Reload failed: %v", err)
        }
    }
}
```

### Disable Features

Run without caching or search:

```go
opts := server.DefaultOptions()
opts.EnableCache = false  // Disable cache
opts.EnableSearch = false // Disable search

blogServer, _ := server.New(opts)
```

## Performance Tuning

### Cache Size

Match cache size to your content:
```go
opts.CacheMaxMB = 200  // 200MB for large blogs
opts.CacheTTL = 30 * time.Minute  // Shorter TTL for frequent updates
```

### Search Index

Rebuild index only when needed:
```go
// First run
opts.RebuildIndex = true

// Subsequent runs
opts.RebuildIndex = false
```

### Pagination

Adjust posts per page:
```go
opts.PostsPerPage = 20  // Show 20 posts per page
```

## Integration Patterns

### With Existing Router

```go
// Your existing application
app := yourframework.New()

// Add blog
blogServer, _ := server.New(server.DefaultOptions())
blogServer.AttachRoutes(app.Router, "/blog")
```

### With Middleware

```go
// Authentication middleware
authMux := http.NewServeMux()
blogServer.AttachRoutes(authMux, "/admin/blog")

mux.Handle("/admin/", requireAuth(authMux))
```

### With Reverse Proxy

```go
// Blog on subdomain
blogServer, _ := server.New(server.DefaultOptions())
mux := http.NewServeMux()
blogServer.AttachRoutes(mux, "/")

// Main app on different port
go http.ListenAndServe(":8080", mainApp)
go http.ListenAndServe(":8081", mux)  // blog.example.com -> :8081
```

## See Also

- [SDK Integration Example](../../examples/sdk-integration/)
- [Standalone Binary](../../cmd/GoBlogServ/)
- [Post Format](../../CLAUDE.md#post-metadata-frontmatter)

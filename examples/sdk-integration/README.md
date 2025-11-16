# SDK Integration Example

This example demonstrates how to integrate GoBlogServ into your own Go application using the SDK.

## Overview

Instead of running GoBlogServ as a standalone binary, you can import it as a library and mount it on your existing HTTP router at any path you choose.

## Features Demonstrated

1. **Custom Application Routes** - Your own handlers for `/`, `/about`, `/contact`
2. **Integrated Blog** - Blog functionality mounted at `/blog/*`
3. **Configuration** - Full control over cache, search, and other settings
4. **Statistics** - Access to blog statistics via the SDK

## Running the Example

```bash
# From this directory
cd examples/sdk-integration

# Make sure you have some posts (or use the example posts)
mkdir -p posts
cp ../posts/*.md posts/

# Run the application
go run main.go
```

Then visit:
- http://localhost:8080/ - Home page
- http://localhost:8080/about - About page
- http://localhost:8080/blog - Blog index
- http://localhost:8080/blog/stats - Blog statistics (JSON)

## Code Walkthrough

### 1. Import the SDK

```go
import "github.com/harrydayexe/GoBlog/pkg/server"
```

### 2. Create Your Router

```go
mux := http.NewServeMux()

// Add your own routes
mux.HandleFunc("GET /", homeHandler)
mux.HandleFunc("GET /about", aboutHandler)
```

### 3. Configure the Blog Server

```go
blogOpts := server.DefaultOptions()
blogOpts.ContentPath = "./posts"
blogOpts.EnableCache = true
blogOpts.EnableSearch = true
blogOpts.Verbose = true

blogServer, err := server.New(blogOpts)
if err != nil {
    log.Fatal(err)
}
defer blogServer.Close()
```

### 4. Attach Blog Routes

```go
// Mount blog at /blog
blogServer.AttachRoutes(mux, "/blog")
```

This creates:
- `GET /blog` - Blog index with search
- `GET /blog/posts/{slug}` - Individual post pages
- `GET /blog/tags/{tag}` - Posts filtered by tag
- `GET /blog/search` - Search endpoint (HTMX)

### 5. Start Your Server

```go
http.ListenAndServe(":8080", mux)
```

## Advanced Usage

### Accessing Internals

You can access the content loader and search index directly:

```go
// Get all posts
posts := blogServer.GetLoader().GetAll()

// Manually search
results, _ := blogServer.GetIndex().Search("golang", 10)

// Reload content
blogServer.Reload()
```

### Custom Statistics Endpoint

```go
mux.HandleFunc("GET /blog/stats", func(w http.ResponseWriter, r *http.Request) {
    stats := blogServer.Stats()
    // Use stats.TotalPosts, stats.CacheHitRatio, etc.
})
```

### Multiple Blogs

You can even mount multiple blog instances:

```go
techBlog, _ := server.New(server.Options{ContentPath: "./tech-posts"})
techBlog.AttachRoutes(mux, "/tech")

personalBlog, _ := server.New(server.Options{ContentPath: "./personal-posts"})
personalBlog.AttachRoutes(mux, "/personal")
```

## Configuration Options

All available options:

```go
type Options struct {
    ContentPath      string        // Path to markdown files (required)
    EnableCache      bool          // Enable Ristretto cache
    CacheMaxMB       int64         // Max cache size in MB
    CacheTTL         time.Duration // How long to cache
    EnableSearch     bool          // Enable Bleve search
    SearchIndexPath  string        // Path to search index
    RebuildIndex     bool          // Rebuild index on startup
    PostsPerPage     int           // Pagination size
    Verbose          bool          // Verbose logging
}
```

## Real-World Use Cases

1. **Personal Website** - Add blog to existing portfolio site
2. **Documentation Site** - Combine docs with a blog/changelog
3. **Company Website** - Add blog to marketing site
4. **Multi-tenant** - Different blogs for different sections
5. **API + Blog** - Serve API and blog from same application

## Performance Tips

1. **Enable Caching** - Set `EnableCache: true` for production
2. **Adjust Cache Size** - Match your content size (default 100MB)
3. **Pre-build Index** - Run with `-rebuild-index` once, then disable
4. **Use CDN** - Serve static assets from CDN, dynamic content from GoBlogServ

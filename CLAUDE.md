# GoBlog - Development Guide

## Project Overview

GoBlog is a two-part system for creating and serving markdown-based blogs:

1. **GoBlogGen**: Unopinionated static site generator that converts markdown posts to HTML
2. **GoBlogServ**: Opinionated web server that dynamically serves posts with HTMX for interactive features

## Current Status (as of 2025-11-14)

### Implemented (~5% complete)
- ✅ Project structure and Go module setup
- ✅ Configuration parser (`internal/gen/config/config.go`) - YAML-based, validates input/output folders
- ✅ Logging system (`internal/gen/log/log.go`) - Colored CLI output with verbose mode
- ✅ CLI entry point (`cmd/GoBlogGen/main.go`) - Basic flag parsing

### Not Yet Implemented
- ❌ Markdown parsing and HTML generation
- ❌ Template system
- ❌ File I/O for posts
- ❌ Static site generation
- ❌ GoBlogServ (nothing started)

### Current Focus
**Phase 1: GoBlogGen Only** - Focusing on static site generation first. GoBlogServ implementation is deferred for future work.

## Architecture Decisions

> **NOTE ON GOBLOGSERV (Future Implementation)**
>
> GoBlogServ has been designed but not yet implemented. When we implement it, the approach will be:
>
> **Core Architecture:**
> - Separate binary from GoBlogGen
> - Reads markdown files directly (no database)
> - Uses stdlib `net/http` router (Go 1.22+)
> - HTMX for dynamic features (search, filtering, pagination)
>
> **Technology Stack:**
> - **HTTP Router**: `net/http` stdlib with enhanced routing
> - **Components**: `github.com/a-h/templ` for type-safe HTML
> - **HTMX Helpers**: `github.com/donseba/go-htmx`
> - **Caching**: `github.com/dgraph-io/ristretto` for in-memory cache
> - **Search**: `github.com/blevesearch/bleve/v2` for full-text search
>
> **Key Features:**
> - Filesystem-only (no database requirement)
> - Live search with HTMX (no page reloads)
> - Tag filtering and pagination via HTMX partials
> - Ristretto caching for parsed markdown
> - Bleve search index for fast queries
>
> **Design Philosophy:**
> - Opinionated (vs GoBlogGen's unopinionated approach)
> - No JavaScript frameworks - pure Go + HTMX
> - Single binary deployment
> - Fast performance through aggressive caching
>
> See "GoBlogServ (Future)" sections below for detailed implementation plans.

### Separation of Concerns
- **GoBlogGen** and **GoBlogServ** are separate binaries
- GoBlogGen can be used standalone for pure static site generation
- GoBlogServ can serve dynamically without pre-generation (reads markdown on-demand)

### Post Metadata (Frontmatter)
Standard YAML frontmatter with these fields:
```yaml
---
title: "Post Title"
date: 2025-01-15
description: "Brief description for SEO and previews"
tags: ["go", "web", "htmx"]
draft: false
---
```

### Technology Stack

#### GoBlogGen Dependencies
- **Markdown Parser**: `github.com/yuin/goldmark`
  - CommonMark compliant (GitHub-compatible)
  - Highly extensible for custom features
  - Active maintenance (Hugo's default parser)

- **Frontmatter**: `go.abhg.dev/goldmark/frontmatter`
  - Integrated with Goldmark (single parse pass)
  - Type-safe struct unmarshaling
  - YAML/TOML/JSON support

- **Syntax Highlighting**: `github.com/yuin/goldmark-highlighting/v2` + `github.com/alecthomas/chroma/v2`
  - 30+ languages supported
  - Customizable themes
  - Integrates seamlessly with Goldmark

- **Templates**: `html/template` (stdlib)
  - User-configurable templates for unopinionated design
  - Runtime flexibility
  - No compilation step needed

#### GoBlogServ Dependencies
- **HTTP Router**: `net/http` (stdlib)
  - Go 1.22+ enhanced routing (method matching, path variables)
  - Zero external dependencies
  - Sufficient for blog use case
  ```go
  mux.HandleFunc("GET /blog/{slug}", getPostHandler)
  slug := r.PathValue("slug")
  ```

- **HTML Components**: `github.com/a-h/templ`
  - Type-safe templates (compile-time checking)
  - Perfect for HTMX partial rendering
  - Compiles to Go code (no runtime parsing)
  - Context-aware escaping

- **HTMX Helpers**: `github.com/donseba/go-htmx`
  - Detect HX-Request headers
  - Handle HX-Trigger responses
  - Simplifies partial rendering logic

- **Caching**: `github.com/dgraph-io/ristretto`
  - High concurrency performance
  - Superior hit rates (TinyLFU admission policy)
  - Cache parsed markdown, rendered templates, search results
  - GC-friendly

- **Search**: `github.com/blevesearch/bleve/v2`
  - No database required (uses BoltDB for persistence)
  - Fast full-text search (50-100ms for millions of docs)
  - 30+ language analyzers
  - Perfect for HTMX live search

### Data Flow

#### GoBlogGen (Static Generation)
```
Markdown files with YAML frontmatter
    ↓
goldmark + goldmark/frontmatter + goldmark-highlighting
    ↓
html/template (user-provided templates)
    ↓
Static HTML/CSS output (./site/)
```

#### GoBlogServ (Dynamic Serving)
```
Markdown files
    ↓
Ristretto cache (parsed content)
    ↓
Bleve index (full-text search)
    ↓
http.ServeMux routing
    ↓
Templ components (HTMX partials)
    ↓
go-htmx helpers (header management)
    ↓
Client browser with HTMX
```

## Configuration Schema

### Current Config (internal/gen/config/config.go)
```yaml
verbose: bool          # Enable debug logging
input_folder: string   # Path to markdown posts (default: ./posts)
output_folder: string  # Path for generated site (default: ./site)
```

### Planned Config Extensions

#### GoBlogGen Config
```yaml
# Site metadata
site:
  title: "My Blog"
  description: "A blog about technology"
  author: "Your Name"
  url: "https://example.com"

# Input/Output
input_folder: "./posts"
output_folder: "./site"

# Templates
templates:
  post: "./templates/post.html"
  index: "./templates/index.html"
  tag: "./templates/tag.html"

# Static assets
static_folder: "./static"

# Pagination
posts_per_page: 10

# Features
enable_syntax_highlighting: true
syntax_theme: "monokai"
enable_toc: true  # Table of contents

# Logging
verbose: false
```

#### GoBlogServ Config
```yaml
# Server settings
server:
  host: "localhost"
  port: 8080

# Content
content_folder: "./posts"  # Read markdown directly

# Cache
cache:
  max_size_mb: 100
  ttl_minutes: 60

# Search
search:
  index_path: "./blog.bleve"
  rebuild_on_start: false

# Blog settings
blog:
  path: "/blog"  # URL path for blog feed
  posts_per_page: 10

# Static files (if serving pre-generated)
static_folder: "./site"

# Logging
verbose: false
```

## Code Organization

### Current Structure
```
/Users/harryday/Developer/GoBlog/
├── cmd/
│   └── GoBlogGen/
│       └── main.go
├── internal/
│   └── gen/
│       ├── config/
│       │   └── config.go
│       └── log/
│           └── log.go
├── pkg/
│   └── api/
├── go.mod
└── README.md
```

### Planned Structure
```
/Users/harryday/Developer/GoBlog/
├── cmd/
│   ├── GoBlogGen/           # Static site generator binary
│   │   └── main.go
│   └── GoBlogServ/          # Web server binary
│       └── main.go
│
├── internal/
│   ├── gen/                 # GoBlogGen internals
│   │   ├── config/
│   │   │   └── config.go    # Configuration parsing
│   │   ├── log/
│   │   │   └── log.go       # Logging utilities
│   │   ├── parser/
│   │   │   ├── parser.go    # Markdown parsing
│   │   │   ├── frontmatter.go
│   │   │   └── post.go      # Post struct & validation
│   │   ├── template/
│   │   │   ├── engine.go    # Template rendering
│   │   │   └── defaults.go  # Default templates
│   │   └── generator/
│   │       ├── generator.go # Site generation orchestration
│   │       ├── posts.go     # Post page generation
│   │       ├── index.go     # Index/list generation
│   │       └── tags.go      # Tag page generation
│   │
│   └── server/              # GoBlogServ internals
│       ├── config/
│       │   └── config.go    # Server configuration
│       ├── content/
│       │   ├── loader.go    # Markdown file loading
│       │   ├── cache.go     # Ristretto cache wrapper
│       │   └── search.go    # Bleve search index
│       ├── handlers/
│       │   ├── posts.go     # Post handlers
│       │   ├── search.go    # Search API handler
│       │   ├── tags.go      # Tag filter handlers
│       │   └── static.go    # Static file serving
│       └── components/      # Templ components
│           ├── layout.templ
│           ├── post.templ
│           ├── postlist.templ
│           ├── search.templ
│           └── pagination.templ
│
├── pkg/
│   └── models/              # Shared models between gen and server
│       ├── post.go          # Post struct
│       └── config.go        # Common config types
│
├── templates/               # Default templates for GoBlogGen
│   ├── post.html
│   ├── index.html
│   └── tag.html
│
├── static/                  # Default static assets
│   ├── css/
│   │   └── style.css
│   └── js/
│       └── htmx.min.js
│
├── examples/                # Example posts and config
│   ├── posts/
│   │   ├── hello-world.md
│   │   └── second-post.md
│   └── config.yaml
│
├── go.mod
├── go.sum
├── README.md
├── CLAUDE.md                # This file
└── LICENSE
```

## Development Workflow

### GoBlogGen Development
1. Write markdown posts in configured input folder
2. Run `go run ./cmd/GoBlogGen -config config.yaml`
3. Generated HTML appears in output folder
4. Upload output folder to any static hosting (GitHub Pages, Netlify, etc.)

### GoBlogServ Development
1. Write markdown posts
2. Run `go run ./cmd/GoBlogServ -config server-config.yaml`
3. Server reads markdown files on-demand
4. Access at http://localhost:8080/blog
5. HTMX provides dynamic search/filtering without page reloads

### Combined Workflow (Recommended for Production)
1. Use GoBlogGen to pre-generate static HTML for SEO
2. Use GoBlogServ to serve the static files with dynamic features on top
3. Best of both worlds: fast static pages + interactive features

## HTMX Integration Patterns

### Search Example
```go
// Handler
func SearchHandler(w http.ResponseWriter, r *http.Request) {
    query := r.FormValue("q")
    results := searchIndex.Search(query)

    if htmx.IsHTMXRequest(r) {
        // Return just the results list
        components.SearchResults(results).Render(r.Context(), w)
    } else {
        // Return full page
        components.SearchPage(results).Render(r.Context(), w)
    }
}
```

```html
<!-- Client-side -->
<input
    type="text"
    name="q"
    hx-get="/api/search"
    hx-trigger="keyup changed delay:300ms"
    hx-target="#search-results"
/>
<div id="search-results"></div>
```

### Tag Filtering Example
```go
func TagFilterHandler(w http.ResponseWriter, r *http.Request) {
    tag := r.PathValue("tag")
    posts := filterPostsByTag(tag)

    if htmx.IsHTMXRequest(r) {
        components.PostList(posts).Render(r.Context(), w)
    } else {
        components.TagPage(tag, posts).Render(r.Context(), w)
    }
}
```

### Pagination Example
```html
<div id="post-list">
    <!-- Post cards here -->
</div>

<button
    hx-get="/blog?page=2"
    hx-target="#post-list"
    hx-swap="beforeend"
>
    Load More
</button>
```

## Testing Approach

### Unit Tests
- Markdown parser tests (various frontmatter formats)
- Template rendering tests
- Configuration validation tests
- Search indexing tests

### Integration Tests
- Full generation pipeline (markdown → HTML)
- Server routing tests
- Cache behavior tests
- HTMX partial rendering tests

### End-to-End Tests
- Generate site from example posts
- Serve site and verify all routes work
- Test search functionality
- Test tag filtering

## Performance Considerations

### GoBlogGen
- Parse all markdown files in parallel
- Cache parsed frontmatter to avoid re-parsing
- Generate pages in parallel where possible
- Compress output HTML/CSS

### GoBlogServ
- **Caching Strategy**:
  - Level 1: Ristretto in-memory cache (hot content)
  - Level 2: Filesystem reads (warm content)
  - Cache invalidation on file modification

- **Search Optimization**:
  - Build Bleve index on startup
  - Keep index in memory for fast queries
  - Update index incrementally when posts change

- **HTMX Benefits**:
  - Reduced bandwidth (partial HTML vs full page)
  - Faster perceived performance
  - Server-side rendering (no JS framework overhead)

## Future Enhancements (Out of Scope for MVP)

### GoBlogGen
- RSS/Atom feed generation
- Sitemap generation
- Image optimization
- Markdown table of contents generation
- Custom shortcodes

### GoBlogServ
- Comments system (optional database)
- Like/reaction system
- View counters
- Admin panel for post management
- Real-time preview mode
- Multi-user support

## Philosophy

### GoBlogGen: Unopinionated
- Users control templates
- Users control styling
- Flexible configuration
- Output is portable HTML/CSS
- No assumptions about hosting

### GoBlogServ: Opinionated
- Structured for performance
- HTMX for interactivity (no complex JS frameworks)
- Filesystem-first (no mandatory database)
- Simple deployment (single binary)
- Reasonable defaults

## Useful Commands

```bash
# Run GoBlogGen
go run ./cmd/GoBlogGen -config config.yaml

# Run GoBlogGen with verbose logging
go run ./cmd/GoBlogGen -config config.yaml -verbose

# Build GoBlogGen binary
go build -o gobloggen ./cmd/GoBlogGen

# Run GoBlogServ
go run ./cmd/GoBlogServ -config server-config.yaml

# Build GoBlogServ binary
go build -o goblogserv ./cmd/GoBlogServ

# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Install Templ CLI (for development)
go install github.com/a-h/templ/cmd/templ@latest

# Generate Templ components
templ generate

# Format Templ files
templ fmt
```

## Resources

- [Goldmark Documentation](https://github.com/yuin/goldmark)
- [Templ Documentation](https://templ.guide/)
- [HTMX Documentation](https://htmx.org/)
- [Bleve Documentation](http://blevesearch.com/docs/)
- [Go HTTP Routing (1.22+)](https://go.dev/blog/routing-enhancements)

## License

MIT License - See LICENSE file for details

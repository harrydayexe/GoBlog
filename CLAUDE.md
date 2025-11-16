# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

GoBlog is a dual-purpose markdown blog system written in Go:

1. **GoBlogGen**: Static site generator that converts markdown posts into HTML
2. **GoBlogServ**: Dynamic web server that serves markdown posts with HTMX-powered interactivity, search (Bleve), and caching (Ristretto)

Both tools share common models (`pkg/models`) but have separate internal implementations.

## Build & Development Commands

### Building Binaries

```bash
# Build both binaries
go build -o gobloggen ./cmd/GoBlogGen
go build -o goblogserv ./cmd/GoBlogServ

# Build specific binary
go build -o bin/gobloggen ./cmd/GoBlogGen
go build -o bin/goblogserv ./cmd/GoBlogServ
```

### Running Locally

```bash
# Run GoBlogGen with config
go run ./cmd/GoBlogGen -config config.yaml

# Run GoBlogGen with verbose logging
go run ./cmd/GoBlogGen -config config.yaml

# Run GoBlogServ with defaults
go run ./cmd/GoBlogServ

# Run GoBlogServ with custom settings
go run ./cmd/GoBlogServ -content ./examples/posts -port 3000 -verbose

# Run GoBlogServ with environment variables
export GOBLOG_CONTENT_PATH=./examples/posts
export GOBLOG_VERBOSE=true
go run ./cmd/GoBlogServ
```

### Testing

```bash
# Run all tests (when implemented)
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for specific package
go test ./pkg/models
go test ./internal/gen/parser
```

### Development Tools

```bash
# Format code
gofmt -w .

# Tidy dependencies
go mod tidy

# Run linting (if golangci-lint installed)
golangci-lint run
```

### Docker

```bash
# Build Docker image
docker build -t goblogserv .

# Run with Docker
docker run -v ./examples/posts:/posts -p 8080:8080 goblogserv

# Run with Docker Compose
docker-compose up
```

### Release Process

```bash
# Tag a new version (triggers release workflow)
git tag v0.1.0
git push origin v0.1.0

# Test GoReleaser locally (requires GoReleaser installed)
goreleaser release --snapshot --clean
```

## Architecture

### Directory Structure

```
cmd/
├── GoBlogGen/    # Static site generator entry point
└── GoBlogServ/   # Web server entry point

internal/
├── gen/          # GoBlogGen internals (not importable by external packages)
│   ├── config/   # Configuration parsing for generator
│   ├── generator/# Core generation orchestration
│   ├── log/      # CLI logging utilities
│   ├── parser/   # Markdown parsing with frontmatter
│   └── template/ # HTML template rendering
└── server/       # GoBlogServ internals
    ├── components/# UI components (likely templ)
    ├── config/   # Server configuration
    ├── content/  # Content loading & caching
    ├── handlers/ # HTTP handlers for posts, tags, search
    └── search/   # Bleve search integration

pkg/              # Public/importable packages
├── api/          # Public API interfaces (if any)
├── models/       # Shared Post model and PostList utilities
└── server/       # Public server package for library usage

examples/
├── posts/        # Example markdown posts
├── sdk-integration/ # Example of using pkg/server as library
├── site/         # Generated static site output
└── static/       # Static assets (CSS, JS, images)

templates/defaults/ # Default HTML templates for GoBlogGen
```

### Key Components

#### Post Model (`pkg/models/post.go`)

The central data structure shared by both tools:

```go
type Post struct {
    // Frontmatter (YAML)
    Title       string
    Date        time.Time
    Description string
    Tags        []string
    Draft       bool

    // Generated
    Slug        string        // URL-friendly
    Content     string        // Rendered HTML
    HTMLContent template.HTML // For templates
    RawContent  string        // Original markdown
    SourcePath  string        // Source file path
}
```

**Important Methods:**
- `Validate()` - Ensures title, date, description are present
- `GenerateSlug()` - Creates URL slug from title or filename
- `IsPublished()` - Returns `!p.Draft`
- `HasTag(tag)` - Case-insensitive tag matching

**PostList helpers:**
- `FilterPublished()` - Excludes drafts
- `FilterByTag(tag)` - Returns posts with tag
- `SortByDate()` - Newest first (bubble sort)
- `GetAllTags()` - Unique tag list

#### GoBlogGen Generator (`internal/gen/generator/generator.go`)

Orchestrates static site generation:

1. **Parse**: Reads markdown files from `input_folder`, extracts YAML frontmatter, renders markdown to HTML
2. **Filter**: Removes draft posts
3. **Generate Posts**: Creates individual HTML pages at `{output_folder}/{blog_path}/{slug}.html`
4. **Generate Index**: Creates paginated index pages (`index.html`, `page2.html`, etc.)
5. **Generate Tag Pages**: Creates tag filter pages at `{output_folder}/{blog_path}/tags/{tag}.html`
6. **Copy Static**: Copies files from `static_folder` to `output_folder`

**Templates** (from `template_dir`):
- `post.html` - Individual post rendering
- `index.html` - Post list with pagination
- `tag.html` - Tag-filtered post list

#### GoBlogServ Server (`pkg/server/server.go`)

Dynamic web server with three layers:

1. **Content Loader** (`internal/server/content/loader.go`)
   - Watches `content_folder` for markdown files
   - Parses frontmatter and renders to HTML
   - Provides `GetAll()`, `GetBySlug()`, `GetAllTags()`

2. **Cache** (`internal/server/content/cache.go`)
   - Ristretto-based in-memory cache
   - Configurable size (MB) and TTL
   - Tracks hits/misses for `/stats` endpoint

3. **Search Index** (`internal/server/search/search.go`)
   - Bleve full-text search
   - Optional rebuild on startup
   - Persisted to disk at `index_path`

**Routes** (attached via `AttachRoutes(mux, basePath)`):
- `GET {basePath}` - Index page
- `GET {basePath}/posts/{slug}` - Individual post
- `GET {basePath}/tags/{tag}` - Tag filter
- `GET {basePath}/search` - HTMX search endpoint

**Standalone binary extras:**
- `GET /health` - Health check
- `GET /stats` - JSON stats (posts, tags, cache metrics, indexed docs)

### Configuration

#### GoBlogGen (`config.yaml`)

```yaml
verbose: false          # Enable debug logging
input_folder: "./posts"
output_folder: "./site"
static_folder: "./static"
template_dir: "./templates/defaults"  # Optional custom templates

site:
  title: "My Blog"
  description: "Blog description"
  author: "Author Name"
  url: "https://example.com"

posts_per_page: 10
blog_path: "/blog"      # URL path for posts
```

#### GoBlogServ (env vars or flags)

**Environment variables:**
- `GOBLOG_CONTENT_PATH` - Path to markdown posts (default: `./posts`)
- `GOBLOG_CACHE_ENABLED` - Enable caching (default: `true`)
- `GOBLOG_CACHE_MAX_MB` - Cache size in MB (default: `100`)
- `GOBLOG_CACHE_TTL` - Cache TTL duration (default: `60m`)
- `GOBLOG_SEARCH_ENABLED` - Enable search (default: `true`)
- `GOBLOG_SEARCH_INDEX_PATH` - Search index path (default: `./blog.bleve`)
- `GOBLOG_REBUILD_INDEX` - Rebuild search index on startup (default: `false`)
- `GOBLOG_POSTS_PER_PAGE` - Posts per page (default: `10`)
- `GOBLOG_VERBOSE` - Enable verbose logging (default: `false`)
- `GOBLOG_PORT` - Server port (default: `8080`)
- `GOBLOG_HOST` - Server host (default: `localhost`)
- `GOBLOG_BASE_PATH` - Base URL path for routes (default: `/blog`)

**Flags override env vars** (see `-help` for full list).

### Markdown Post Format

All posts require YAML frontmatter:

```markdown
---
title: "Post Title"
date: 2025-01-15
description: "Brief description for SEO/preview"
tags: ["tag1", "tag2"]
draft: false
---

# Post Content

Your markdown content here...
```

**Required fields:** `title`, `date`, `description`

**Optional fields:** `tags` (array), `draft` (boolean, default `false`)

## Key Dependencies

- **goldmark**: Markdown parser with frontmatter support (`go.abhg.dev/goldmark/frontmatter`)
- **chroma**: Syntax highlighting for code blocks (`github.com/alecthomas/chroma/v2`)
- **bleve**: Full-text search engine (`github.com/blevesearch/bleve/v2`)
- **ristretto**: High-performance cache (`github.com/dgraph-io/ristretto`)
- **templ**: Go templating for components (`github.com/a-h/templ`) - used in server components
- **go-htmx**: HTMX utilities (`github.com/donseba/go-htmx`)

## Common Workflows

### Adding a New Feature to GoBlogGen

1. Determine if it affects parsing, template rendering, or generation orchestration
2. Add logic to appropriate internal package (`parser`, `template`, or `generator`)
3. Update config struct if new settings needed (`internal/gen/config/config.go`)
4. Test manually with `go run ./cmd/GoBlogGen -config examples/config.yaml`

### Adding a New Feature to GoBlogServ

1. Determine if it's a new route (handler), content processing (loader), or caching/search concern
2. Add handler to `internal/server/handlers/`
3. Register route in `pkg/server/server.go` `AttachRoutes()`
4. Update `pkg/server/options.go` if new configuration needed
5. Test with `go run ./cmd/GoBlogServ -content examples/posts -verbose`

### Using GoBlogServ as a Library

See `examples/sdk-integration/main.go` for reference:

```go
import "github.com/harrydayexe/GoBlog/pkg/server"

// Load options from env
opts, _ := server.LoadFromEnv()

// Create server
blogServer, _ := server.New(opts)
defer blogServer.Close()

// Attach to your mux
mux := http.NewServeMux()
blogServer.AttachRoutes(mux, "/blog")
```

### Modifying Templates

Default templates are embedded in the binary but can be overridden:

1. Copy from `templates/defaults/` to a custom directory
2. Edit `post.html`, `index.html`, or `tag.html`
3. Update `config.yaml` with `template_dir: "./my-templates"`
4. Available template data is documented in README.md (search for "Template Data")

### Release Process

Releases are automated via GoReleaser:

1. Create and push a git tag: `git tag v0.x.x && git push origin v0.x.x`
2. GitHub Actions runs `.github/workflows/release.yml`
3. GoReleaser builds binaries for Linux/macOS/Windows (amd64/arm64)
4. Homebrew tap is updated automatically (`harrydayexe/homebrew-goblog`)
5. Docker images are built and pushed via `.github/workflows/docker.yml`

**Note:** GoBlogServ is commented out in `.goreleaser.yml` line 31 - uncomment when ready for release.

## GitHub Workflows

- **ci.yml**: Runs on PRs, executes tests (go test ./...)
- **release.yml**: Triggered by tags, runs GoReleaser
- **docker.yml**: Builds and pushes Docker images
- **tag.yml**: Automates tagging process
- **claude.yml**: Claude PR Assistant
- **claude-code-review.yml**: Claude Code Review bot

## Important Notes

- **No tests currently exist** - when adding tests, use standard Go testing (`*_test.go` files)
- **Server uses templ for components** - if modifying `internal/server/components/*.templ`, run `templ generate` after editing
- **Static assets** are separate from templates - templates are for HTML structure, static assets (CSS/JS) go in `static_folder`
- **Blog path is configurable** - don't hardcode `/blog`, use `cfg.BlogPath` or `basePath` parameter
- **Posts are cached** - when debugging GoBlogServ, remember cache may return stale content unless `GOBLOG_CACHE_ENABLED=false` or restart server
- **Search index is persistent** - delete `blog.bleve/` directory or use `-rebuild-index` to rebuild
- **Environment variables take precedence order**: CLI flags > env vars > defaults

## Distribution Channels

- **Homebrew**: `brew install harrydayexe/goblog/gobloggen`
- **Direct Download**: GitHub releases (Linux/macOS/Windows binaries)
- **Go Install**: `go install github.com/harrydayexe/GoBlog/cmd/GoBlogGen@latest`
- **Docker**: `docker pull harrydayexe/goblogserv:latest` (for GoBlogServ)
- **Go Library**: Import `github.com/harrydayexe/GoBlog/pkg/server` for SDK usage
- Always run go fmt, go vet and golangci-lint before a commit. Fix any issues which arise
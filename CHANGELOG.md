# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Middleware support for custom authentication, rate limiting, and request logging
- Event callbacks (OnPostView, OnSearch, OnError, OnReload) for analytics integration
- Partial rendering API (RenderPost, RenderPostList) for HTMX integration
- File watching with fsnotify for automatic content reload during development
- RSS 2.0 feed generation for both GoBlogGen and GoBlogServ
- Atom 1.0 feed generation for both GoBlogGen and GoBlogServ
- XML sitemap generation with proper SEO metadata
- Fluent API methods for easier SDK configuration (WithContentPath, WithCache, etc.)
- CORS support with configurable origins, methods, and headers
- Structured error handling with BlogError type and error codes
- Custom CSS injection support via Options
- Comprehensive SECURITY.md with forward-only patching policy
- Comprehensive CONTRIBUTING.md with Conventional Commits enforcement

### Changed
- Enhanced error messages with machine-readable error codes
- Improved feed handler integration in GoBlogServ

### Fixed
- Type conversion in Atom feed HTML content rendering

## [1.0.0] - TBD

Initial stable release of GoBlog.

### Features

**GoBlogGen (Static Site Generator)**
- Parse markdown files with YAML frontmatter
- Generate static HTML sites with customizable templates
- Syntax highlighting support (30+ languages)
- Pagination for blog index pages
- Tag-based post filtering and tag pages
- RSS/Atom feed generation
- XML sitemap generation
- Static asset copying
- Draft post filtering

**GoBlogServ (Dynamic Server)**
- Serve markdown posts dynamically without pre-generation
- In-memory caching with Ristretto for high performance
- Full-text search with Bleve search engine
- HTMX integration for interactive features
- File watching for development hot-reload
- Middleware support for custom handlers
- Event callbacks for monitoring and analytics
- CORS support for cross-origin requests
- Health check endpoint
- Feed and sitemap endpoints

**Embedding SDK (pkg/server)**
- Fluent API for easy configuration
- Embeddable in existing Go applications
- Attach blog routes to existing http.ServeMux
- Customizable content paths and URL prefixes
- Partial rendering for HTMX integration
- Production-ready with caching and search

### Documentation
- Comprehensive README with examples
- SDK documentation in pkg/server/README.md
- Example integration code in examples/
- Security policy (SECURITY.md)
- Contribution guidelines (CONTRIBUTING.md)
- Development guide (CLAUDE.md)

### Dependencies
- Go 1.23+ (toolchain 1.24.7)
- github.com/yuin/goldmark for markdown parsing
- github.com/blevesearch/bleve/v2 for search
- github.com/dgraph-io/ristretto for caching
- github.com/a-h/templ for type-safe templates
- github.com/fsnotify/fsnotify for file watching
- gopkg.in/yaml.v3 for configuration

---

## Release Notes Format

Each release will document:
- **Added**: New features
- **Changed**: Changes to existing functionality
- **Deprecated**: Soon-to-be-removed features
- **Removed**: Removed features
- **Fixed**: Bug fixes
- **Security**: Security fixes (see SECURITY.md for policy)

## Versioning Policy

GoBlog follows [Semantic Versioning](https://semver.org/):

- **MAJOR** (X.0.0): Incompatible API changes
- **MINOR** (0.X.0): New functionality (backwards-compatible)
- **PATCH** (0.0.X): Bug fixes (backwards-compatible)

**Security patches are forward-only** - see SECURITY.md for details.

## Links

- [Unreleased changes](https://github.com/harrydayexe/GoBlog/compare/main...HEAD)
- [All releases](https://github.com/harrydayexe/GoBlog/releases)
- [Security policy](./SECURITY.md)
- [Contributing guide](./CONTRIBUTING.md)

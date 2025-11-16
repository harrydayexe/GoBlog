# API Stability Guarantees

This document outlines GoBlog's commitment to API stability and backward compatibility.

## Semantic Versioning Promise

GoBlog strictly follows [Semantic Versioning 2.0.0](https://semver.org/):

```
MAJOR.MINOR.PATCH
```

- **MAJOR** (X.0.0): Breaking changes - may require code changes
- **MINOR** (0.X.0): New features - backward compatible
- **PATCH** (0.0.X): Bug fixes - backward compatible

### Version Guarantees

**Pre-1.0.0 (Current Status)**
- API may change without notice
- Breaking changes can occur in MINOR versions
- Use at your own risk in production
- Feedback welcome to shape 1.0.0 API

**Post-1.0.0 (Stable Release)**
- **MAJOR** versions only for breaking changes
- **MINOR** versions are safe to upgrade
- **PATCH** versions are always safe to upgrade
- Deprecation warnings before removal (at least one MINOR version)

## What's Considered a Breaking Change

### Breaking (Requires MAJOR Version Bump)

**Struct Changes:**
```go
// ❌ BREAKING: Removing exported fields
type Options struct {
    // ContentPath string  // REMOVED
    EnableCache bool
}

// ❌ BREAKING: Changing field types
type Options struct {
    CacheTTL int  // Was: time.Duration
}

// ❌ BREAKING: Renaming exported fields
type Options struct {
    CacheEnabled bool  // Was: EnableCache
}
```

**Function Signature Changes:**
```go
// ❌ BREAKING: Changing return types
func New(opts Options) *Server  // Was: (*Server, error)

// ❌ BREAKING: Changing parameter types
func New(opts *Options) (*Server, error)  // Was: (opts Options)

// ❌ BREAKING: Removing parameters
func New() (*Server, error)  // Was: New(opts Options)

// ❌ BREAKING: Adding required parameters
func New(opts Options, logger *log.Logger) (*Server, error)
```

**Behavioral Changes:**
```go
// ❌ BREAKING: Changing default behavior
// v1: Default cache TTL was 1 hour
// v2: Default cache TTL is now 10 minutes

// ❌ BREAKING: Removing exported methods
// server.Start() no longer exists

// ❌ BREAKING: Changing error conditions
// GetPost() now returns ErrNotFound instead of nil
```

**Package Changes:**
```go
// ❌ BREAKING: Moving packages
// Was: github.com/harrydayexe/GoBlog/pkg/server
// Now: github.com/harrydayexe/GoBlog/server

// ❌ BREAKING: Renaming packages
// Was: import "github.com/harrydayexe/GoBlog/pkg/models"
// Now: import "github.com/harrydayexe/GoBlog/pkg/types"
```

### Non-Breaking (Safe for MINOR/PATCH)

**Additions:**
```go
// ✅ SAFE: Adding new exported fields (with defaults)
type Options struct {
    EnableCache bool
    EnableSearch bool  // NEW - defaults to false
}

// ✅ SAFE: Adding new methods
func (s *Server) GetMetrics() Metrics  // NEW

// ✅ SAFE: Adding new functions
func NewWithDefaults() (*Server, error)  // NEW

// ✅ SAFE: Adding variadic options
func New(opts Options, features ...Feature) (*Server, error)
```

**Internal Changes:**
```go
// ✅ SAFE: Changing private fields
type Server struct {
    cache *cache  // Changed implementation
}

// ✅ SAFE: Performance improvements
// Faster search algorithm (same API)

// ✅ SAFE: Bug fixes
// Fixed off-by-one error in pagination
```

**Compatible Changes:**
```go
// ✅ SAFE: Expanding error types (with compatibility)
type BlogError struct {
    Code string  // NEW field
    Message string
    Err error
}
// Still implements error interface

// ✅ SAFE: Adding interface methods to unexported types
// (only if interface is not meant for user implementation)
```

## Stability Levels

GoBlog APIs are classified into three stability levels:

### 1. Stable API (Guaranteed)

**Location:** `pkg/server/`

**Guarantee:** Will not break without MAJOR version bump

**Includes:**
```go
// Core embedding SDK
pkg/server.Server
pkg/server.Options
pkg/server.DefaultOptions()
pkg/server.New()

// Public methods
(*Server).Start()
(*Server).Shutdown()
(*Server).AttachRoutes()
(*Server).RenderPost()
(*Server).RenderPostList()

// Fluent API
(Options).WithContentPath()
(Options).WithCache()
(Options).WithSearch()
// ... all With* methods
```

**Models:**
```go
pkg/models.Post
pkg/models.PostList
```

**Stability Promise:**
- Field additions only (with sensible defaults)
- Method additions only
- Behavior changes only for bug fixes
- Deprecation warnings before removal

### 2. Semi-Stable API (Use with Caution)

**Location:** `internal/`

**Guarantee:** May change in MINOR versions (but rarely)

**Includes:**
```go
// Internal packages
internal/gen/
internal/server/

// These are NOT meant for external use
// But if you import them, expect changes
```

**When We Might Change:**
- Performance optimizations
- Bug fixes that require restructuring
- Security improvements

**Migration Path:**
- Changes documented in CHANGELOG
- Deprecation warnings when feasible
- At least one MINOR version notice

### 3. Experimental API (No Guarantees)

**Markers:**
- Documented as "experimental" in godoc
- May have `Experimental` in function name
- Explicitly marked in README

**Current Experimental Features:**
```go
// None currently
// Future experimental features will be clearly marked
```

**Stability Promise:**
- May change at any time
- May be removed without deprecation
- Use at your own risk
- Feedback appreciated to stabilize

## Deprecation Policy

When we need to remove or change stable APIs:

### 1. Deprecation Notice (MINOR version)

```go
// Deprecated: Use NewWithContext instead.
// This function will be removed in v2.0.0.
func New(opts Options) (*Server, error) {
    return NewWithContext(context.Background(), opts)
}
```

**Guarantees:**
- At least one MINOR version warning
- Documentation updated with migration path
- CHANGELOG entry with alternatives

### 2. Removal (MAJOR version)

```go
// v2.0.0: Old function removed
// Use NewWithContext instead
```

**Timeline:**
- Minimum 6 months between deprecation and removal
- Announced in release notes
- Migration guide provided

## Compatibility Guarantees

### Go Version Compatibility

**Minimum Go Version:**
- GoBlog 1.x: Go 1.23+
- May increase minimum in MAJOR versions only
- MINOR/PATCH versions will not increase minimum Go version

**Toolchain:**
- Toolchain version may increase in MINOR versions
- Only if backward compatible with minimum Go version

### Dependency Compatibility

**Major Dependencies:**
- goldmark, bleve, ristretto, templ
- We will not introduce breaking dependency updates in MINOR versions
- Security patches may require dependency updates (PATCH versions)

**Transitive Dependencies:**
- We use `go.mod` version constraints
- You should vendor if you need strict reproducibility

### Binary Compatibility

**GoBlogGen and GoBlogServ:**
- Command-line flags will not change in incompatible ways
- New flags may be added (opt-in)
- Flag removal requires deprecation period

**Configuration Files:**
```yaml
# config.yaml format
# v1.0: Original format
site:
  title: "My Blog"

# v1.5: Can add new fields (with defaults)
site:
  title: "My Blog"
  favicon: "/favicon.ico"  # NEW (optional)

# v2.0: Can change structure
metadata:  # BREAKING: Was "site"
  title: "My Blog"
```

## Interface Compatibility

### Interfaces You Implement

**If you implement these interfaces, we guarantee:**

```go
// http.Handler - stdlib, never changes
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request)

// Future custom interfaces will be documented
```

**We will NOT:**
- Add methods to interfaces you implement (without MAJOR bump)
- Change method signatures (without MAJOR bump)

### Interfaces We Implement

**You can rely on:**

```go
// Server implements http.Handler
var _ http.Handler = (*Server)(nil)

// BlogError implements error
var _ error = (*BlogError)(nil)
```

## What's NOT Covered

These are **not** considered breaking changes:

### 1. Internal Implementation Details

```go
// We can change these without notice:
// - Cache eviction algorithms
// - Search index format (auto-rebuilt)
// - Template rendering order
// - Goroutine usage
// - Memory allocation patterns
```

### 2. Bugs and Security Fixes

```go
// If current behavior is a bug, we'll fix it:
// - Off-by-one errors
// - Race conditions
// - Memory leaks
// - Security vulnerabilities

// Even if you depend on the buggy behavior
```

### 3. Performance Characteristics

```go
// We may change performance without notice:
// - O(n) → O(log n) is NOT breaking
// - Faster search is NOT breaking
// - Lower memory usage is NOT breaking

// But we won't make things intentionally slower
```

### 4. Error Messages

```go
// Error messages may change:
err := fmt.Errorf("post not found: %s", slug)
// May become:
err := fmt.Errorf("post %q does not exist", slug)

// Don't parse error strings!
// Use error codes instead:
if blogErr, ok := err.(*BlogError); ok {
    if blogErr.Code == "POST_NOT_FOUND" {
        // This is stable
    }
}
```

### 5. Unstable Features

```go
// Features marked "experimental"
// Features in internal/ packages
// Features without tests
```

## Version Support Policy

**Active Support:**
- Latest MAJOR version only
- Security patches: forward-only (see SECURITY.md)
- Bug fixes: latest MINOR version only

**Example:**
```
Current release: 1.5.0
Vulnerability found affecting 1.0.0+

Fix released as: 1.5.1 (not 1.0.1, 1.1.1, etc.)
Users must upgrade to 1.5.1
```

**Long-Term Support:**
- No LTS versions planned
- Always upgrade to latest MAJOR version
- MAJOR versions released only when necessary (years between)

## How to Stay Compatible

### 1. Pin to MAJOR Versions

```go
// go.mod
require github.com/harrydayexe/GoBlog v1.5.0

// Safe to upgrade to any 1.x.x
go get -u github.com/harrydayexe/GoBlog@v1
```

### 2. Watch for Deprecation Warnings

```bash
# Compile-time warnings
go build ./...

# Check for deprecated usage
golangci-lint run
```

### 3. Read CHANGELOG Before Upgrading

```bash
# Before upgrading
git fetch
git log v1.5.0..v1.6.0 -- CHANGELOG.md
```

### 4. Test Before Deploying

```bash
# Test with new version
go test ./...

# Integration tests
# End-to-end tests
```

## Breaking Change Examples (When 2.0 Comes)

Here are hypothetical examples of what might justify a 2.0:

### Scenario 1: Context Everywhere

```go
// v1.x
func (s *Server) GetPost(slug string) (*Post, error)

// v2.0: Add context for cancellation
func (s *Server) GetPost(ctx context.Context, slug string) (*Post, error)
```

**Migration:**
```go
// Old code:
post, err := server.GetPost("my-post")

// New code:
post, err := server.GetPost(ctx, "my-post")
```

### Scenario 2: Options Restructure

```go
// v1.x
type Options struct {
    ContentPath string
    EnableCache bool
    CacheMaxMB int64
    CacheTTL time.Duration
    EnableSearch bool
    SearchIndexPath string
}

// v2.0: Nested configuration
type Options struct {
    ContentPath string
    Cache CacheOptions
    Search SearchOptions
}

type CacheOptions struct {
    Enabled bool
    MaxMB int64
    TTL time.Duration
}
```

**Migration Guide:**
- Would be provided in 2.0.0 release notes
- Automated migration tool where feasible

## Questions?

- **API clarification**: Open a [GitHub Discussion](https://github.com/harrydayexe/GoBlog/discussions)
- **Stability concerns**: Open an [Issue](https://github.com/harrydayexe/GoBlog/issues)
- **Breaking change proposals**: See CONTRIBUTING.md

## References

- [Semantic Versioning Specification](https://semver.org/)
- [Go 1 Compatibility Promise](https://go.dev/doc/go1compat)
- [Keep a Changelog](https://keepachangelog.com/)
- [GoBlog Security Policy](./SECURITY.md)
- [GoBlog Contributing Guide](./CONTRIBUTING.md)

---

*This API stability policy takes effect with GoBlog 1.0.0*

*Last updated: 2025-01-16*

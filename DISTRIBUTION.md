# GoBlog Distribution Strategy

This document outlines how GoBlogGen and GoBlogServ are packaged and distributed to users.

## GoBlogGen (Static Site Generator)

GoBlogGen is distributed as a command-line tool through multiple channels.

### Distribution Channels

#### 1. GitHub Releases (Primary)
Automated via GoReleaser on each tagged release.

**Platforms:**
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64)

**Artifacts:**
- Compressed archives (tar.gz for Unix, zip for Windows)
- SHA256 checksums
- Auto-generated changelog

#### 2. Homebrew Tap (Recommended)
Custom tap for easy installation on macOS and Linux.

**Setup Required:**
1. Create `harrydayexe/homebrew-goblog` repository
2. GoReleaser automatically pushes formula updates on each release
3. Users install via: `brew install harrydayexe/goblog/gobloggen`

**Benefits:**
- One-command installation
- Automatic updates via `brew upgrade`
- No manual PATH configuration

**Status:** Configured in `.goreleaser.yml`, requires tap repository creation

#### 3. Homebrew Core (Long-term Goal)
Official Homebrew repository for broader reach.

**Requirements:**
- Stable project (6+ months of releases)
- Notable userbase (typically 50+ GitHub stars)
- Well-documented and tested
- Active maintenance
- No security issues

**Process:**
1. Build userbase with custom tap first
2. After meeting criteria, submit PR to `homebrew/homebrew-core`
3. Homebrew maintainers review and approve

**Timeline:** 6-12 months after initial release

#### 4. Go Install (For Go Developers)
Direct installation from source.

```bash
go install github.com/harrydayexe/GoBlog/cmd/GoBlogGen@latest
```

**Benefits:**
- Always latest version
- No pre-built binaries needed
- Familiar to Go developers

### Installation Instructions

The README.md should include these methods in order of preference:

1. Homebrew (simplest for most users)
2. Direct binary download (for users without Homebrew)
3. Go install (for developers)

## GoBlogServ (Dynamic Server)

GoBlogServ has dual distribution: as a Go library and as a standalone binary/container.

### Distribution Model 1: Go Library/Package

**Purpose:** For developers who want to integrate blog functionality into existing Go applications.

**Usage:**
```go
import "github.com/harrydayexe/GoBlog/pkg/server"

func main() {
    blogServer := server.New(server.Config{
        ContentFolder: "./posts",
        CacheSize: 100,
    })

    // Option A: Mount to existing router
    mux.Handle("/blog/", blogServer.Routes())

    // Option B: Use individual handlers
    mux.HandleFunc("GET /blog/{slug}", blogServer.PostHandler)
    mux.HandleFunc("GET /api/search", blogServer.SearchHandler)
}
```

**Package Structure:**
```
pkg/
└── server/
    ├── server.go        # Main server struct and config
    ├── handlers.go      # HTTP handlers
    ├── middleware.go    # HTMX detection, caching
    ├── cache.go         # Ristretto wrapper
    ├── search.go        # Bleve integration
    └── components/      # Templ components (compiled to .go)
        ├── layout.templ
        ├── post.templ
        └── search.templ
```

**Documentation:**
- GoDoc for all exported types
- Example applications in `examples/integration/`
- Integration guide in docs

### Distribution Model 2: Standalone Binary + Docker

**Purpose:** For users who want a turnkey blog server without custom Go code.

#### Standalone Binary

Distributed via same channels as GoBlogGen:
- GitHub Releases
- Homebrew tap: `brew install harrydayexe/goblog/goblogserv`
- Go install: `go install github.com/harrydayexe/GoBlog/cmd/GoBlogServ@latest`

**Usage:**
```bash
# Run with config file
goblogserv -config server.yaml

# Run with environment variables
export GOBLOG_CONTENT_FOLDER=./posts
export GOBLOG_PORT=8080
goblogserv

# Run with flags
goblogserv -content ./posts -port 8080
```

#### Docker Container

**Registry:** Docker Hub

**Images:**
- `harrydayexe/goblogserv:latest`
- `harrydayexe/goblogserv:v1.2.3`
- `harrydayexe/goblogserv:v1` (major version tag)

**Dockerfile:**
```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /build
COPY . .
RUN go build -o goblogserv ./cmd/GoBlogServ

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /build/goblogserv /usr/local/bin/
EXPOSE 8080
ENTRYPOINT ["goblogserv"]
```

**Usage:**
```bash
# Basic usage
docker run -v ./posts:/posts -p 8080:8080 harrydayexe/goblogserv

# With config
docker run -v ./config.yaml:/config.yaml \
           -v ./posts:/posts \
           -p 8080:8080 \
           harrydayexe/goblogserv -config /config.yaml

# Docker Compose (sidecar pattern)
version: '3.8'
services:
  app:
    build: .
    ports:
      - "3000:3000"

  blog:
    image: harrydayexe/goblogserv:latest
    volumes:
      - ./posts:/posts
    environment:
      - GOBLOG_CONTENT_FOLDER=/posts
      - GOBLOG_PORT=8080
    ports:
      - "8080:8080"
```

**Kubernetes/Sidecar Pattern:**
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: app-with-blog
spec:
  containers:
  - name: main-app
    image: my-app:latest
    ports:
    - containerPort: 3000

  - name: blog
    image: harrydayexe/goblogserv:latest
    ports:
    - containerPort: 8080
    volumeMounts:
    - name: blog-posts
      mountPath: /posts
    env:
    - name: GOBLOG_CONTENT_FOLDER
      value: /posts

  volumes:
  - name: blog-posts
    configMap:
      name: blog-posts
```

### Docker Image Building

**Automated via GitHub Actions:**

The Docker build workflow (`.github/workflows/docker.yml`) automatically builds and pushes images to Docker Hub when a version tag is pushed.

**Triggers:**
- Push of version tags (`v*`) - Builds release images
- Pull requests - Builds PR snapshot images for testing
- Manual workflow dispatch

**Required Secrets:**
- `DOCKERHUB_USERNAME` - Your Docker Hub username
- `DOCKERHUB_TOKEN` - Docker Hub access token (create at https://hub.docker.com/settings/security)

**Platforms:**
- `linux/amd64`
- `linux/arm64`

**Generated Tags:**

*Release builds (from version tags):*
- `harrydayexe/goblogserv:v1.2.3` (specific version)
- `harrydayexe/goblogserv:v1.2` (minor version)
- `harrydayexe/goblogserv:v1` (major version)
- `harrydayexe/goblogserv:latest` (latest release)

*PR snapshot builds (from pull requests):*
- `harrydayexe/goblogserv:pr-123-abc1234` (PR number + short commit SHA)
  - Example: `docker pull harrydayexe/goblogserv:pr-42-a1b2c3d`
  - These are temporary builds for testing PRs
  - Not tagged as `latest`
  - May be removed after PR is merged/closed

## Combined Distribution Strategy

### Recommended Approach

**For most users:**
1. Use GoBlogGen to generate static site (install via Homebrew)
2. Deploy static files to any host (GitHub Pages, Netlify, etc.)

**For users wanting dynamic features:**
1. Use GoBlogServ Docker container as sidecar
2. Serves markdown files with HTMX-powered search/filtering
3. No build step required

**For developers:**
1. Import `pkg/server` into their Go application
2. Customize handlers and templates as needed
3. Full control over routing and integration

### Repository Setup Needed

1. **Create Homebrew tap:** `harrydayexe/homebrew-goblog`
2. **Configure Docker registry:** Enable GHCR in repo settings
3. **Add GitHub Actions workflow:** Docker build on tag push
4. **Update README.md:** Installation instructions for all methods

## Next Steps

1. Create `homebrew-goblog` repository
2. Add Dockerfile for GoBlogServ
3. Create GitHub Actions workflow for Docker builds
4. Write comprehensive README.md
5. Add example applications
6. Create documentation site (using GoBlog itself!)

## Testing Distribution

Before first release:

1. Test GoReleaser locally: `goreleaser release --snapshot --clean`
2. Test Homebrew formula locally: `brew install --build-from-source ./Formula/gobloggen.rb`
3. Test Docker build: `docker build -t goblogserv:test .`
4. Test Docker run: `docker run -v ./posts:/posts -p 8080:8080 goblogserv:test`
5. Verify all installation methods in README work

## Version Numbering

Follow Semantic Versioning (SemVer):

- **v0.x.x**: Pre-1.0 development (current phase)
- **v1.0.0**: First stable release (GoBlogGen complete)
- **v1.1.0**: GoBlogServ library release
- **v1.2.0**: GoBlogServ Docker release
- **v2.0.0**: Breaking API changes (if needed)

## Release Process

1. Update CHANGELOG.md
2. Create git tag: `git tag -a v1.0.0 -m "Release v1.0.0"`
3. Push tag: `git push origin v1.0.0`
4. GoReleaser builds binaries and updates Homebrew tap
5. GitHub Actions builds and pushes Docker images
6. Verify all distribution channels work
7. Announce release on GitHub Discussions/social media

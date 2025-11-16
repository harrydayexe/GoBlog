# GoBlogServ Docker Image

Opinionated web server for serving markdown-based blogs with HTMX-powered interactivity.

## Features

- Real-time search powered by Bleve
- Tag filtering and pagination
- HTMX for dynamic updates without page reloads
- In-memory caching with Ristretto
- Single binary deployment
- Multi-platform support (linux/amd64, linux/arm64)

## Quick Start

### Basic Usage

```bash
docker run -v ./posts:/posts -p 8080:8080 harrydayexe/goblogserv:latest
```

Visit http://localhost:8080 to view your blog.

### With Custom Configuration

```bash
docker run \
  -v ./config.yaml:/app/config.yaml \
  -v ./posts:/posts \
  -p 8080:8080 \
  harrydayexe/goblogserv:latest \
  -config /app/config.yaml
```

## Docker Compose

### Standalone

```yaml
version: '3.8'
services:
  blog:
    image: harrydayexe/goblogserv:latest
    volumes:
      - ./posts:/posts
    ports:
      - "8080:8080"
    environment:
      - TZ=America/New_York
```

### Sidecar Pattern

```yaml
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

## Configuration

### Environment Variables

- `GOBLOG_CONTENT_FOLDER` - Path to markdown posts (default: `/posts`)
- `GOBLOG_PORT` - Server port (default: `8080`)
- `GOBLOG_HOST` - Server host (default: `localhost`)
- `TZ` - Timezone for the container

### Command-line Flags

```bash
docker run harrydayexe/goblogserv:latest -help
```

Available flags:
- `-content` - Path to markdown posts
- `-port` - Server port
- `-config` - Path to YAML configuration file
- `-verbose` - Enable debug logging

## Post Format

Create markdown files with YAML frontmatter:

```markdown
---
title: "Hello World"
date: 2025-01-15
description: "My first blog post"
tags: ["intro", "blogging"]
draft: false
---

# Hello World

This is my first blog post!
```

## Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: goblogserv
spec:
  replicas: 2
  selector:
    matchLabels:
      app: goblogserv
  template:
    metadata:
      labels:
        app: goblogserv
    spec:
      containers:
      - name: goblogserv
        image: harrydayexe/goblogserv:latest
        ports:
        - containerPort: 8080
        volumeMounts:
        - name: blog-posts
          mountPath: /posts
        env:
        - name: GOBLOG_CONTENT_FOLDER
          value: /posts
        - name: GOBLOG_PORT
          value: "8080"
      volumes:
      - name: blog-posts
        configMap:
          name: blog-posts
---
apiVersion: v1
kind: Service
metadata:
  name: goblogserv
spec:
  selector:
    app: goblogserv
  ports:
  - port: 80
    targetPort: 8080
  type: LoadBalancer
```

## Health Check

The container includes a health check endpoint at `/health`:

```bash
curl http://localhost:8080/health
```

## Security

- Runs as non-root user (UID 1000)
- Minimal Alpine-based image
- No unnecessary packages
- Compiled with `-ldflags="-s -w"` for smaller binary size

## Volumes

- `/posts` - Mount your markdown posts here

## Ports

- `8080` - HTTP server (configurable)

## Links

- **GitHub**: https://github.com/harrydayexe/GoBlog
- **Documentation**: https://github.com/harrydayexe/GoBlog/wiki
- **Issues**: https://github.com/harrydayexe/GoBlog/issues

## License

MIT License - See [LICENSE](https://github.com/harrydayexe/GoBlog/blob/main/LICENSE) for details.

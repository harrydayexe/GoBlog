# GoBlog

[![Go Reference](https://pkg.go.dev/badge/github.com/harrydayexe/GoBlog.svg)](https://pkg.go.dev/github.com/harrydayexe/GoBlog/v2)
[![Go Report Card](https://goreportcard.com/badge/github.com/harrydayexe/GoBlog/v2)](https://goreportcard.com/report/github.com/harrydayexe/GoBlog/v2)
[![Test](https://github.com/harrydayexe/GoBlog/actions/workflows/test.yml/badge.svg?event=push)](https://github.com/harrydayexe/GoBlog/actions/workflows/test.yml)
[![License: MIT](https://img.shields.io/badge/License-GPL3-yellow.svg)](LICENSE)
[![DockerHub](https://img.shields.io/docker/v/harrydayexe/goblog?sort=semver)](https://hub.docker.com/repository/docker/harrydayexe/goblog/general)

GoBlog is a blog generation and serving system for creating static blog feeds from Markdown files. It is available as a CLI tool, a Docker image, and an embeddable Go library.

## CLI

Install the `goblog` binary:

```bash
go install github.com/harrydayexe/GoBlog/v2/cmd/goblog@latest
```

```bash
# Generate static files
goblog generate posts/ output/

# Serve locally
goblog serve posts/
```

### `generate` flags

| Flag | Short | Default | Description |
|---|---|---|---|
| `--raw` | `-r` | `false` | Output raw HTML without template wrapping |
| `--disable-tags` | `-T` | `false` | Disable tag tracking and tag page generation |
| `--disable-reading-time` | | `false` | Disable reading time estimation on posts |
| `--root-path` | `-p` | `/` | Blog root path for subdirectory deployment |
| `--template-dir` | `-t` | built-in | Path to a custom template directory |

### `serve` flags

| Flag | Short | Default | Description |
|---|---|---|---|
| `--port` | `-P` | `8080` | TCP port to listen on |
| `--host` | `-H` | all interfaces | Host address to bind to |
| `--disable-tags` | `-T` | `false` | Disable tag tracking and tag page generation |
| `--disable-reading-time` | | `false` | Disable reading time estimation on posts |
| `--root-path` | `-p` | `/` | Blog root path for subdirectory deployment |
| `--template-dir` | `-t` | built-in | Path to a custom template directory |
| `--watch` | `-w` | `false` | Watch the posts directory and regenerate on changes |

## Docker

The official image is [`harrydayexe/goblog`](https://hub.docker.com/repository/docker/harrydayexe/goblog/general). It runs `goblog serve /posts` by default and exposes port `8080`. File watching is off by default; pass `--watch` to enable it.

Mount your Markdown posts directory to `/posts`:

```bash
docker run -v ./posts:/posts -p 8080:8080 harrydayexe/goblog
```

To watch for post changes and reload automatically:

```bash
docker run -v ./posts:/posts -p 8080:8080 harrydayexe/goblog /posts --watch
```

Pass any `serve` flags after the image name — re-supply the posts path as the first argument:

```bash
docker run -v ./posts:/posts -p 9000:9000 harrydayexe/goblog /posts --port 9000
docker run -v ./posts:/posts -p 8080:8080 harrydayexe/goblog /posts --root-path /blog/
```

For custom templates, mount your template directory and use `--template-dir`:

```bash
docker run \
  -v ./posts:/posts \
  -v ./mytheme:/mytheme \
  -p 8080:8080 \
  harrydayexe/goblog /posts --template-dir /mytheme
```

## Library

Add GoBlog as a dependency:

```bash
go get github.com/harrydayexe/GoBlog/v2
```

The main packages are:

| Package | Summary |
|---|---|
| [`pkg/parser`](https://pkg.go.dev/github.com/harrydayexe/GoBlog/v2/pkg/parser) | Parse Markdown + YAML frontmatter into `Post` objects |
| [`pkg/generator`](https://pkg.go.dev/github.com/harrydayexe/GoBlog/v2/pkg/generator) | Convert a posts directory into a `GeneratedBlog` in memory |
| [`pkg/outputter`](https://pkg.go.dev/github.com/harrydayexe/GoBlog/v2/pkg/outputter) | Write a `GeneratedBlog` to disk or a custom destination |
| [`pkg/server`](https://pkg.go.dev/github.com/harrydayexe/GoBlog/v2/pkg/server) | Embeddable HTTP server with atomic live-reload |
| [`pkg/config`](https://pkg.go.dev/github.com/harrydayexe/GoBlog/v2/pkg/config) | Functional options for generator, outputter, and server |
| [`pkg/models`](https://pkg.go.dev/github.com/harrydayexe/GoBlog/v2/pkg/models) | Core data types: `Post`, `PostList`, template data structs |
| [`pkg/templates`](https://pkg.go.dev/github.com/harrydayexe/GoBlog/v2/pkg/templates) | Embedded default templates (`templates.Default`) |

A minimal generate-and-write example:

```go
package main

import (
    "context"
    "os"

    "github.com/harrydayexe/GoBlog/v2/pkg/generator"
    "github.com/harrydayexe/GoBlog/v2/pkg/outputter"
    "github.com/harrydayexe/GoBlog/v2/pkg/templates"
)

func main() {
    fsys := os.DirFS("posts/")
    renderer, err := generator.NewTemplateRenderer(templates.Default)
    if err != nil {
        panic(err)
    }

    gen := generator.New(fsys, renderer)
    blog, err := gen.Generate(context.Background())
    if err != nil {
        panic(err)
    }

    writer := outputter.NewDirectoryWriter("output/")
    writer.HandleGeneratedBlog(context.Background(), blog)
}
```

### Logger injection

Every component accepts a structured [`log/slog`](https://pkg.go.dev/log/slog) logger via `config.WithLogger`. When not supplied, each component falls back to `slog.Default()` at construction time.

```go
logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

// Generator and outputter: wrap in WithBaseOption
gen := generator.New(fsys, renderer,
    config.WithBaseOption(config.WithLogger(logger)),
)
writer := outputter.NewDirectoryWriter("output/",
    config.WithBaseOption(config.WithLogger(logger)),
)

// Server: embed in BaseServerOption
cfg := config.ServerConfig{
    Server: []config.BaseServerOption{
        config.WithPort(8080),
        {BaseOption: config.WithLogger(logger)},
    },
}
srv, err := server.New(nil, postsFS, cfg)

// Watcher: wrap in WithBaseWatcherOption
w, err := watcher.New("posts/", config.WithBaseWatcherOption(config.WithLogger(logger)))

// Parser: use parser.WithLogger directly
p := parser.New(parser.WithLogger(logger))
```

Full API documentation, including all config options and template data types, is at [pkg.go.dev/github.com/harrydayexe/GoBlog/v2](https://pkg.go.dev/github.com/harrydayexe/GoBlog/v2).

## Contributing

Contributions are welcome. See [CONTRIBUTING.md](CONTRIBUTING.md) for how to set up the project, run tests, and submit pull requests.

## License

GNU General Public License v3.0 — see [LICENSE](LICENSE) for details.

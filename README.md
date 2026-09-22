# GoBlog

[![Go Reference](https://pkg.go.dev/badge/github.com/harrydayexe/GoBlog.svg)](https://pkg.go.dev/github.com/harrydayexe/GoBlog/v2)
[![Go Report Card](https://goreportcard.com/badge/github.com/harrydayexe/GoBlog/v2)](https://goreportcard.com/report/github.com/harrydayexe/GoBlog/v2)
[![Test](https://github.com/harrydayexe/GoBlog/actions/workflows/test.yml/badge.svg?event=push)](https://github.com/harrydayexe/GoBlog/actions/workflows/test.yml)
[![License: MIT](https://img.shields.io/badge/License-GPL3-yellow.svg)](LICENSE)
[![DockerHub](https://img.shields.io/docker/v/harrydayexe/goblog?sort=semver)](https://hub.docker.com/repository/docker/harrydayexe/goblog/general)

GoBlog is a blog generation and serving system for creating static blog feeds from Markdown files. It is available as a CLI tool, a Docker image, and an embeddable Go library.

## CLI

The `goblog` binary lives in its own Go module (`cli/`) that is not published to
the module proxy, so it is installed from a package manager or a release archive
rather than with `go install`.

**Homebrew** (macOS and Linux):

```bash
brew install harrydayexe/tap/goblog
```

**Release archive** — grab the archive for your platform from the
[releases page](https://github.com/harrydayexe/GoBlog/releases) and put the
binary on your `PATH`:

```bash
curl -sSL https://github.com/harrydayexe/GoBlog/releases/latest/download/GoBlog_Linux_x86_64.tar.gz | tar -xz goblog
sudo install goblog /usr/local/bin/goblog
```

Archives are published for Linux and macOS on `x86_64` and `arm64`. Windows
archives are built as well but Windows is not a supported install target.

There is also a [Docker image](#docker) if you only need to serve a blog.

```bash
# Generate static files
goblog generate posts/ output/

# Serve locally
goblog serve posts/
```

### Global flags

These flags apply to both `generate` and `serve` and may be passed before or after the subcommand name.

| Flag | Short | Default | Description |
|---|---|---|---|
| `--template-dir` | `-t` | built-in | Path to a custom template directory |
| `--root-path` | `-p` | `/` | Blog root path for subdirectory deployment |
| `--disable-tags` | `-T` | `false` | Disable tag tracking and tag page generation |
| `--disable-reading-time` | | `false` | Disable reading time estimation on posts |
| `--base-url` | | _(none)_ | Scheme + host of the site (e.g. `https://example.com`); required to generate RSS/Atom feeds and canonical/Open Graph URLs. Must not include a path — use `--root-path` for subdirectory deployments |
| `--disable-feeds` | | `false` | Disable RSS and Atom feed generation |
| `--feed-limit` | | `10` | Maximum number of posts to include in each feed (`0` = unlimited) |
| `--assets-dir` | | `<posts>/images` | Directory of images, served at `{root-path}images/`, copied to `<output>/images/`, and read when parsing to measure each image. Ignored if it does not exist |
| `--disable-sitemap` | | `false` | Disable `sitemap.xml` generation |
| `--disable-robots` | | `false` | Disable `robots.txt` generation |
| `--robots-file` | | _(none)_ | Path to a custom `robots.txt` whose contents replace the default rules. The `Sitemap:` line is still appended when a sitemap is generated. Cannot be combined with `--disable-robots` |

`--base-url` also enables `sitemap.xml` and `robots.txt`. With `generate`, both are written to the top of the output directory; when `--root-path` is not `/`, move `robots.txt` to your domain root on deploy, since crawlers only read `/robots.txt`.

### `generate` flags

| Flag | Short | Default | Description |
|---|---|---|---|
| `--raw` | `-r` | `false` | Output raw HTML without template wrapping |

### `serve` flags

When `--base-url` is set, the server also exposes the generated feeds at `{root-path}rss.xml`, `{root-path}atom.xml`, and per-tag feeds at `{root-path}tags/{tag}.rss.xml` / `{root-path}tags/{tag}.atom.xml`. The sitemap is served at `{root-path}sitemap.xml`, and `robots.txt` at `/robots.txt` — the origin root, where crawlers look for it, regardless of `--root-path`.

| Flag | Short | Default | Description |
|---|---|---|---|
| `--port` | `-P` | `8080` | TCP port to listen on |
| `--host` | `-H` | all interfaces | Host address to bind to |
| `--watch` | `-w` | `false` | Watch the posts directory and regenerate on changes |
| `--cache-control` | | `1h` | Max-age TTL for the `Cache-Control` header (`0` disables) |
| `--health-checks` | | `false` | Expose `/healthz/live`, `/healthz/ready`, and `/healthz/startup` endpoints (no auth required); server binds before loading content so probes observe startup state |

### Shell completion

The Homebrew cask installs bash, zsh, and fish completions for you. Release
archives ship the same scripts in a `completions/` directory.

`goblog` can also generate them at runtime — source the appropriate script to
enable tab-completion of subcommands and flags.

**Bash** — add to `~/.bashrc`:

```bash
source <(goblog completion bash)
```

**Zsh** — add to `~/.zshrc` (requires `compinit` to be loaded):

```zsh
autoload -Uz compinit && compinit
source <(goblog completion zsh)
```

**Fish** — write the script to your completions directory:

```fish
goblog completion fish > ~/.config/fish/completions/goblog.fish
```

## Docker

The official image is [`harrydayexe/goblog`](https://hub.docker.com/repository/docker/harrydayexe/goblog/general). It runs `goblog serve --health-checks /posts` by default and exposes port `8080`. Health-check endpoints are enabled in the Docker image. File watching is off by default; pass `--watch` to enable it.

Mount your Markdown posts directory to `/posts`:

```bash
docker run -v ./posts:/posts -p 8080:8080 harrydayexe/goblog
```

The image exposes three health-check endpoints that require no authentication:

| Endpoint | Purpose | Response |
|---|---|---|
| `GET /healthz/live` | Liveness probe | `200 ok` (always) |
| `GET /healthz/ready` | Readiness probe | `200 ok` once posts are loaded; `503` while starting or on error |
| `GET /healthz/startup` | Startup probe | Same semantics as `/healthz/ready` |

To watch for post changes and reload automatically:

```bash
docker run -v ./posts:/posts -p 8080:8080 harrydayexe/goblog /posts --watch
```

Pass any `serve` flags after the image name — re-supply the posts path as the first argument:

```bash
docker run -v ./posts:/posts -p 9000:9000 harrydayexe/goblog /posts --port 9000
docker run -v ./posts:/posts -p 8080:8080 harrydayexe/goblog /posts --root-path /blog/
```

Images in `./posts/images` are served at `/images/` with no extra flags.

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

The CLI is a separate module (`cli/`) that is never published, so none of its
dependencies reach your build.

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

Serving the same posts over HTTP instead. `server.New` takes functional
options; options belonging to the generator or the template renderer are
converted with `AsServerOption`:

```go
srv, err := server.New(os.DirFS("posts/"),
    config.WithPort(8080),
    config.WithSiteTitle("My Blog").AsServerOption(),
    config.WithBaseURL("https://example.com").AsServerOption(),
)
if err != nil {
    panic(err)
}

if err := srv.Run(context.Background()); err != nil {
    panic(err)
}
```

### Logger injection

Every component accepts a structured [`log/slog`](https://pkg.go.dev/log/slog) logger via `config.WithLogger`. When not supplied, each component falls back to `slog.Default()` at construction time.

```go
logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

// Generator and outputter
gen := generator.New(fsys, renderer,
    config.WithLogger(logger).AsGeneratorOption(),
)
writer := outputter.NewDirectoryWriter("output/",
    config.WithLogger(logger).AsGeneratorOption(),
)

// Server
srv, err := server.New(postsFS,
    config.WithPort(8080),
    config.WithLogger(logger).AsServerOption(),
)

// Watcher
w, err := watcher.New("posts/", config.WithLogger(logger).AsWatcherOption())

// Parser
p := parser.New(parser.WithLogger(logger))
```

### SEO metadata

The default templates emit social and search metadata with no template work required:

- **Open Graph** — `og:title`, `og:description`, `og:site_name`, and `og:type` (`article` on posts, `website` elsewhere). Post pages also emit `article:published_time`, `article:modified_time` (from `lastEdited`), `article:author`, and one `article:tag` per tag.
- **Schema.org JSON-LD** — a `BlogPosting` object on post pages, carrying the publish and modified dates, author, keywords, reading time, and publisher; a `WebSite` object elsewhere.
- **Canonical URLs** — `og:url` and `<link rel="canonical">`.
- **X/Twitter** — `twitter:card`, which reads its content from the Open Graph tags above.

Canonical URLs need to know the site's domain, so set `--base-url` (or `config.WithBaseURL`). Without it, every URL-bearing tag is omitted rather than emitted empty; the rest of the metadata is unaffected.

Setting a base URL also populates `GeneratedBlog.Sitemap` and `GeneratedBlog.RobotsTxt`, which `pkg/outputter` writes as `sitemap.xml` / `robots.txt` and `pkg/server` serves. Opt out with `config.WithDisableSitemap()` / `config.WithDisableRobotsTxt()`, or replace the default robots rules with `config.WithRobotsTxt(body)`.

Custom templates can read the same values from the page data: `{{.CanonicalURL}}`, `{{.OGType}}`, and `{{with .Article}}` for the post's publish date, author, and tags. `og:image` is not emitted — posts have no image field.

Full API documentation, including all config options and template data types, is at [pkg.go.dev/github.com/harrydayexe/GoBlog/v2](https://pkg.go.dev/github.com/harrydayexe/GoBlog/v2).

## Front matter

Each post starts with a YAML front matter block:

```yaml
---
title: "Getting Started"
metaTitle: "Getting Started with GoBlog: A Static Blog Generator in Go"
date: 2026-09-17T10:00:00Z
description: "How to turn a directory of Markdown files into a blog."
tags: [go, tutorial]
author: "Jane Doe"
lastEdited: 2026-09-18T09:00:00Z
---
```

| Field | Required | Purpose |
|---|---|---|
| `title` | yes | Display title: the post heading and the text on post cards |
| `date` | yes | Publication date, used for ordering and article metadata |
| `description` | yes | Meta description, also shown on post cards |
| `metaTitle` | no | Title for the `<title>` element and title-based meta tags; falls back to `title` |
| `tags` | no | Tags the post is listed under |
| `author` | no | Post author |
| `lastEdited` | no | Date the post was revised after publication; must not be before `date` |

`metaTitle` exists because the two jobs `title` does have different constraints: a heading can be short and rely on page context, while the `<title>` wants its keyword near the front and has to fit alongside the ` | {{.SiteTitle}}` suffix. Setting it changes the `<title>`, `og:title`, and JSON-LD headline only — the heading, post cards, and feed items keep `title`.

## Heading anchor links

Headings get auto-generated ids (`## Future Work` → `id="future-work"`). Besides standard `[text](#future-work)` links, posts can link to a heading in the same post with wikilink syntax:

```md
See [[#Future Work]] or [[#Future Work|what comes next]].
```

As with standard links, targets are not validated, so a link to a missing heading renders without error. Links to other posts (`[[other-post#heading]]`) are not supported and render as plain text.

## Images

Put images in an `images/` directory inside your posts directory and reference them from a post in either form:

```md
![A diagram of the pipeline](images/pipeline.png)
![A diagram of the pipeline](pipeline.png)
![[pipeline.png|A diagram of the pipeline]]
```

All three render as:

```html
<img src="{root-path}images/pipeline.png" alt="A diagram of the pipeline" width="1200" height="800" loading="lazy" decoding="async" />
```

The `width` and `height` are the image's intrinsic pixel size, read from the asset file while the post is parsed, so the browser can reserve space for the image and the page does not jump as it loads. Files that cannot be measured — missing ones, and formats whose header cannot be decoded such as SVG, AVIF and WebP — simply render without the two attributes. Subdirectories are preserved (`images/screenshots/a.png`). A bare `![[pipeline.png]]` renders with `alt=""`, which tells a screen reader there is nothing to announce; add a `|label` when the image carries meaning. Absolute URLs and paths starting with `/` are left as written and get no dimensions, since there is no local file to measure. As with heading links, missing image files are not reported.

Use `--assets-dir` to keep images elsewhere; if the directory does not exist, image support is simply off. `serve` reads images straight from disk, so adding or replacing one needs no reload (the directory must exist when the server starts); dimensions are measured at parse time, so swapping in a differently sized image needs a reload for `width` and `height` to catch up. `generate` copies the directory into `<output>/images/`.

Library users pass the directory with `config.WithAssetsDir`, ideally via `os.Root` so symlinks cannot escape it:

```go
root, err := os.OpenRoot("posts/images")
if err != nil {
    panic(err)
}
defer root.Close()

gen := generator.New(postsFS, renderer, config.WithAssetsDir(root.FS()).AsGeneratorOption())
writer := outputter.NewDirectoryWriter("output/", config.WithAssetsDir(root.FS()).AsGeneratorOption())
srv, err := server.New(postsFS, config.WithAssetsDir(root.FS()).AsServerOption())
```

The generator forwards the directory to the parser, which is what measures the images. Using the parser on its own, pass it directly with `parser.WithAssetsDir(root.FS())`.

## Contributing

Contributions are welcome. See [CONTRIBUTING.md](CONTRIBUTING.md) for how to set up the project, run tests, and submit pull requests.

## License

GNU General Public License v3.0 — see [LICENSE](LICENSE) for details.

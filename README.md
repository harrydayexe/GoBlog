# GoBlog

[![Go Reference](https://pkg.go.dev/badge/github.com/harrydayexe/GoBlog.svg)](https://pkg.go.dev/github.com/harrydayexe/GoBlog/v2)
[![Go Report Card](https://goreportcard.com/badge/github.com/harrydayexe/GoBlog/v2)](https://goreportcard.com/report/github.com/harrydayexe/GoBlog/v2)
[![Test](https://github.com/harrydayexe/GoBlog/actions/workflows/test.yml/badge.svg?event=push)](https://github.com/harrydayexe/GoBlog/actions/workflows/test.yml)
[![License: MIT](https://img.shields.io/badge/License-GPL3-yellow.svg)](LICENSE)
[![DockerHub](https://img.shields.io/docker/v/harrydayexe/goblog?sort=semver)](https://hub.docker.com/repository/docker/harrydayexe/goblog/general)

GoBlog is a flexible blog generation and serving system for creating static blog feeds from Markdown files. It provides a powerful parser for Markdown content with YAML frontmatter, as well as multiple deployment options including a CLI tool, Docker image, and embeddable Go package.

The project is designed for developers who want a simple, Go-based solution for blog generation with support for modern Markdown features like syntax highlighting, footnotes, and custom templates.

## Installation

### CLI tool

Install the `goblog` binary to `$GOPATH/bin` (ensure that directory is on your `$PATH`):

```bash
go install github.com/harrydayexe/GoBlog/v2/cmd/goblog@latest
```

To install from a local checkout instead, run from the repo root:

```bash
go install ./cmd/goblog
```

### Library

Add GoBlog as a dependency in your Go module:

```bash
go get github.com/harrydayexe/GoBlog/v2
```

## Quick Start

### Using the Parser Package

The parser package reads Markdown files with YAML frontmatter and converts them to structured Post objects:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/harrydayexe/GoBlog/v2/pkg/parser"
)

func main() {
    // Create a new parser (syntax highlighting enabled by default)
    p := parser.New()

    // Parse a single markdown file
    postsFS := os.DirFS("posts/")
    post, err := p.ParseFile(context.Background(), postsFS, "my-post.md")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Title: %s\n", post.Title)
    fmt.Printf("Date: %s\n", post.FormattedDate())
    fmt.Printf("Tags: %v\n", post.Tags)
}
```

### Parsing Multiple Posts

```go
// Parse all markdown files in a directory
posts, err := p.ParseDirectory(context.Background(), postsFS)
if err != nil {
    log.Fatal(err)
}

// Sort by date and filter by tag
posts.SortByDate()
goPosts := posts.FilterByTag("go")

fmt.Printf("Found %d posts about Go\n", len(goPosts))
```

## Documentation

Full API documentation is available at:
- **pkg.go.dev**: https://pkg.go.dev/github.com/harrydayexe/GoBlog/v2

## Markdown Post Format

Posts are written in Markdown with YAML frontmatter:

```markdown
---
title: "My Blog Post Title"
date: 2024-01-15
description: "A brief description of the post"
tags: ["go", "blogging", "markdown"]
---

# Post Content

Your blog post content goes here with **full Markdown support**.

\`\`\`go
func main() {
    fmt.Println("Code highlighting included!")
}
\`\`\`
```

## CLI Usage

```bash
# Generate static blog (alias: goblog g)
goblog generate posts/ output/

# Serve blog locally
goblog serve posts/ --port 8080
```

### `generate` flags

| Flag | Short | Default | Description |
|---|---|---|---|
| `--raw` | `-r` | `false` | Output raw HTML without template wrapping |
| `--disable-tags` | `-T` | `false` | Disable tag tracking and tag page generation |
| `--root-path` | `-p` | `/` | Blog root path for subdirectory deployment |
| `--template-dir` | `-t` | built-in | Path to a custom template directory |

### `serve` flags

| Flag | Short | Default | Description |
|---|---|---|---|
| `--port` | `-P` | `8080` | TCP port to listen on |
| `--host` | `-H` | all interfaces | Host address to bind to |
| `--disable-tags` | `-T` | `false` | Disable tag tracking and tag page generation |
| `--root-path` | `-p` | `/` | Blog root path for subdirectory deployment |
| `--template-dir` | `-t` | built-in | Path to a custom template directory |

## Advanced Features

### Syntax Highlighting CSS

The parser renders code blocks using **CSS classes** (via chroma's `html.WithClasses` option) rather than inline styles. This keeps the generated HTML clean but means you must include a matching stylesheet in your templates.

Generate the stylesheet for any [chroma style](https://github.com/alecthomas/chroma/tree/master/styles) at startup and embed it in a `<style>` tag:

```go
import (
    chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
    "github.com/alecthomas/chroma/v2/styles"
    "strings"
)

formatter := chromahtml.New(chromahtml.WithClasses(true))
style := styles.Get("monokai") // choose any chroma style name
var sb strings.Builder
formatter.WriteCSS(&sb, style)
chromaCSS := sb.String() // embed in a <style> tag in your template
```

There is currently no API to change the highlighter style from within GoBlog — pick whichever style you like when generating the stylesheet above and the CSS class names will match.

The CSS class names follow the Pygments short-name convention (`.k` for keyword, `.s` for string, `.nf` for function name, etc.). A full reference is in [chroma's types.go](https://github.com/alecthomas/chroma/blob/master/types.go).

Footnotes are disabled by default. Enable them with `parser.WithFootnote()`:

```go
p := parser.New(parser.WithFootnote())
```

### Raw Output Mode

GoBlog supports raw HTML output mode for advanced use cases where you need direct access to the generated HTML without template wrappers. This is useful when:

- Integrating GoBlog output into an existing site or CMS
- Building custom templates in your own application
- Embedding blog content in other HTML frameworks
- Processing HTML content programmatically before display

#### Using Raw Output with the CLI

```bash
# Generate raw HTML files without templates
goblog generate posts/ output/ --raw

# Or use the short flag
goblog generate posts/ output/ -r
```

When raw output mode is enabled:
- Individual post files (under `posts/`) contain only the parsed Markdown converted to HTML
- No template wrapping is applied to the content
- The `tags/` directory is not created (tag pages are skipped)
- `index.html` is still written by `DirectoryWriter` but its body is empty

#### Using Raw Output with the Go API

```go
package main

import (
    "context"
    "os"

    "github.com/harrydayexe/GoBlog/v2/pkg/config"
    "github.com/harrydayexe/GoBlog/v2/pkg/generator"
    "github.com/harrydayexe/GoBlog/v2/pkg/outputter"
)

func main() {
    // Create generator with raw output enabled (no renderer needed in raw mode)
    fsys := os.DirFS("posts/")
    gen := generator.New(fsys, nil, config.WithRawOutput())

    // Generate the blog
    blog, err := gen.Generate(context.Background())
    if err != nil {
        panic(err)
    }

    // Access raw HTML bytes directly
    for slug, htmlContent := range blog.Posts {
        // Process or wrap htmlContent as needed
        // htmlContent contains only the parsed Markdown as HTML
    }

    // Write to disk with raw output mode
    writer := outputter.NewDirectoryWriter("output/", config.WithRawOutput())
    writer.HandleGeneratedBlog(context.Background(), blog)
}
```

#### What to Expect from Raw Output

When `RawOutput` is enabled, the `GeneratedBlog` structure contains:

- **`Posts` map**: Keys are post slugs (derived from post titles or filenames), values are raw HTML byte slices containing only the Markdown content converted to HTML
- **`Index` field**: Empty in raw mode — no index page is generated
- **`Tags` map**: Empty — tag pages are not generated in raw output mode
- **`TagsIndex` field**: Empty — tags index is not generated in raw output mode

The HTML content is clean, semantic HTML generated from your Markdown without any surrounding structure like `<html>`, `<head>`, or `<body>` tags. This gives you complete control over how to integrate the content into your site.

### Disable Tags Mode

GoBlog can be configured to skip all tag-related output while still applying full templates to posts and the index page. This is useful when:

- Your content is not taxonomy-driven and you don't need tag pages
- You want a simpler site structure without a `/tags/` section
- You are using a custom navigation scheme in place of GoBlog's built-in tag pages

#### Using Disable Tags with the CLI

```bash
# Generate without tag pages
goblog generate posts/ output/ --disable-tags

# Or use the short flag
goblog generate posts/ output/ -T

# Also works with serve
goblog serve posts/ --disable-tags
```

When disable tags mode is enabled:
- Individual post files are rendered with full templates; with the default templates, per-post tag pills are hidden
- The `tags/` directory is not created — no individual tag pages or tags index
- `index.html` is rendered as normal; with the default templates, the "Tags" nav link in the header is hidden
- The `/tags` and `/tags/{tag}` routes return 404 in serve mode

#### Using Disable Tags with the Go API

```go
package main

import (
    "context"
    "os"

    "github.com/harrydayexe/GoBlog/v2/pkg/config"
    "github.com/harrydayexe/GoBlog/v2/pkg/generator"
    "github.com/harrydayexe/GoBlog/v2/pkg/outputter"
    "github.com/harrydayexe/GoBlog/v2/pkg/templates"
)

func main() {
    fsys := os.DirFS("posts/")
    renderer, _ := generator.NewTemplateRenderer(templates.Default)
    gen := generator.New(fsys, renderer, config.WithDisableTags())

    blog, err := gen.Generate(context.Background())
    if err != nil {
        panic(err)
    }

    // blog.Tags is an empty map; blog.TagsIndex is nil
    // blog.Posts and blog.Index contain fully-rendered HTML without tag UI

    writer := outputter.NewDirectoryWriter("output/", config.WithDisableTags())
    writer.HandleGeneratedBlog(context.Background(), blog)
}
```

#### What to Expect from Disable Tags Output

When `DisableTags` is enabled, the `GeneratedBlog` structure contains:

- **`Posts` map**: Keys are post slugs, values are fully-templated HTML pages; with the default templates, tag pills are not rendered
- **`Index` field**: Fully-templated index page; with the default templates, the "Tags" nav link is not rendered
- **`Tags` map**: Empty — individual tag pages are not generated
- **`TagsIndex` field**: Empty — the tags index page is not generated

#### Custom Templates and TagsEnabled

The `BaseData` struct passed to all templates includes a `TagsEnabled bool` field. Custom templates should use this field to conditionally render tag-related UI:

```html
{{if .TagsEnabled}}
<a href="{{.BlogRoot}}tags">Tags</a>
{{end}}
```

The built-in templates already respect this field. If you provide custom templates that unconditionally render tag links, those links will still appear even when `--disable-tags` is set — update them to gate on `{{.TagsEnabled}}`.

### Blog Root Configuration

When deploying your blog at a subdirectory rather than the root of your domain (e.g., `example.com/blog/` instead of `example.com/`), you need to configure the blog root path. This ensures all generated links in templates use the correct base path.

#### Using Blog Root with the CLI

```bash
# Generate blog for deployment at /blog/ subdirectory
goblog generate posts/ output/ --root-path /blog/

# Or use the short flag
goblog generate posts/ output/ -p /blog/

# Serve blog locally with custom root path
goblog serve posts/ --root-path /blog/ --port 8080
```

When blog root is configured:
- All internal links (navigation, post links, tag links) will use the specified root path
- Links are generated as `/blog/posts/my-post.html` instead of `/posts/my-post.html`
- Home links point to `/blog/` instead of `/`
- Tag links use `/blog/tags/tag-name.html` instead of `/tags/tag-name.html`

#### Using Blog Root with the Go API

```go
package main

import (
    "context"
    "os"

    "github.com/harrydayexe/GoBlog/v2/pkg/config"
    "github.com/harrydayexe/GoBlog/v2/pkg/generator"
    "github.com/harrydayexe/GoBlog/v2/pkg/outputter"
)

func main() {
    // Create generator with blog root for subdirectory deployment
    fsys := os.DirFS("posts/")
    renderer, _ := generator.NewTemplateRenderer(templates.Default)
    gen := generator.New(fsys, renderer, config.WithBlogRoot("/blog/"))

    // Generate the blog
    blog, err := gen.Generate(context.Background())
    if err != nil {
        panic(err)
    }

    // Write to disk - all links will use /blog/ prefix
    writer := outputter.NewDirectoryWriter("output/")
    writer.HandleGeneratedBlog(context.Background(), blog)
}
```

#### Common Use Cases

**Root Deployment (Default)**:
```bash
# Blog at example.com/
goblog generate posts/ output/
# Links: /posts/slug.html, /tags/tag.html, /
```

**Subdirectory Deployment**:
```bash
# Blog at example.com/blog/
goblog generate posts/ output/ --root-path /blog/
# Links: /blog/posts/slug.html, /blog/tags/tag.html, /blog/
```

**Nested Subdirectory**:
```bash
# Blog at example.com/docs/blog/
goblog generate posts/ output/ --root-path /docs/blog/
# Links: /docs/blog/posts/slug.html, /docs/blog/tags/tag.html, /docs/blog/
```

### Custom Templates

GoBlog ships with built-in templates (`templates.Default`) but you can supply
any `fs.FS` — e.g. `os.DirFS("./mytheme")` — to `generator.NewTemplateRenderer`
for a fully custom theme.

#### Required directory layout

Your template root must contain:

```
mytheme/
  pages/
    post.tmpl          receives models.PostPageData
    index.tmpl         receives models.IndexPageData
    tag.tmpl           receives models.TagPageData
    tags-index.tmpl    receives models.TagsIndexPageData
  partials/
    head.tmpl          must {{define "head"}}
    header.tmpl        must {{define "header"}}
    footer.tmpl        must {{define "footer"}}
    post-card.tmpl     must {{define "post-card"}}
  layouts/             optional — loaded but not executed by default
```

Each page template is a complete HTML document. It pulls in partials with:

```html
{{template "head" .}}
{{template "header" .}}
{{template "footer" .}}
```

The `post-card` partial is referenced inside `index.tmpl` and `tag.tmpl` when
ranging over a list of posts.

#### Template data

Every page template receives a struct that embeds
[`models.BaseData`](https://pkg.go.dev/github.com/harrydayexe/GoBlog/v2/pkg/models#BaseData)
(`SiteTitle`, `PageTitle`, `Description`, `Year`, `BlogRoot`, `Environment`), plus
page-specific fields:

| Template | Data struct | Extra fields |
|---|---|---|
| `pages/post.tmpl` | [`PostPageData`](https://pkg.go.dev/github.com/harrydayexe/GoBlog/v2/pkg/models#PostPageData) | `.Post *Post` |
| `pages/index.tmpl` | [`IndexPageData`](https://pkg.go.dev/github.com/harrydayexe/GoBlog/v2/pkg/models#IndexPageData) | `.Posts PostList`, `.TotalPosts int` |
| `pages/tag.tmpl` | [`TagPageData`](https://pkg.go.dev/github.com/harrydayexe/GoBlog/v2/pkg/models#TagPageData) | `.Tag string`, `.Posts []*Post`, `.PostCount int` — not rendered when `--disable-tags` is set |
| `pages/tags-index.tmpl` | [`TagsIndexPageData`](https://pkg.go.dev/github.com/harrydayexe/GoBlog/v2/pkg/models#TagsIndexPageData) | `.Tags []TagInfo`, `.TotalTags int` — not rendered when `--disable-tags` is set |

#### Template helpers (FuncMap)

| Name | Signature | Example output |
|---|---|---|
| `formatDate` | `func(t time.Time) string` | `"January 2, 2006"` |
| `shortDate` | `func(t time.Time) string` | `"Jan 2, 2006"` |
| `year` | `func() int` | `2025` |

#### Go API example

```go
package main

import (
    "context"
    "os"

    "github.com/harrydayexe/GoBlog/v2/pkg/config"
    "github.com/harrydayexe/GoBlog/v2/pkg/generator"
    "github.com/harrydayexe/GoBlog/v2/pkg/outputter"
)

func main() {
    // Use a custom template directory
    renderer, err := generator.NewTemplateRenderer(os.DirFS("./mytheme"))
    if err != nil {
        panic(err)
    }

    fsys := os.DirFS("posts/")
    gen := generator.New(fsys, renderer, config.WithBlogRoot("/blog/"))

    blog, err := gen.Generate(context.Background())
    if err != nil {
        panic(err)
    }

    writer := outputter.NewDirectoryWriter("output/")
    writer.HandleGeneratedBlog(context.Background(), blog)
}
```

To use the built-in templates instead, replace `os.DirFS("./mytheme")` with
`templates.Default` (from `github.com/harrydayexe/GoBlog/v2/pkg/templates`).

## Go Package Reference

| Package | Summary |
|---|---|
| [`pkg/parser`](https://pkg.go.dev/github.com/harrydayexe/GoBlog/v2/pkg/parser) | Parse Markdown + YAML frontmatter into `Post` objects |
| [`pkg/models`](https://pkg.go.dev/github.com/harrydayexe/GoBlog/v2/pkg/models) | Core data types: `Post`, `PostList`, page data structs |
| [`pkg/generator`](https://pkg.go.dev/github.com/harrydayexe/GoBlog/v2/pkg/generator) | Convert a directory of posts into a `GeneratedBlog` in memory |
| [`pkg/outputter`](https://pkg.go.dev/github.com/harrydayexe/GoBlog/v2/pkg/outputter) | Write a `GeneratedBlog` to disk; implement `Outputter` for custom destinations |
| [`pkg/server`](https://pkg.go.dev/github.com/harrydayexe/GoBlog/v2/pkg/server) | HTTP server with atomic live-reload via `Server.UpdatePosts` |
| [`pkg/config`](https://pkg.go.dev/github.com/harrydayexe/GoBlog/v2/pkg/config) | Functional options: `WithRawOutput`, `WithDisableTags`, `WithSiteTitle`, `WithBlogRoot`, `WithEnvironment`, `WithPort`, `WithHost`, `WithMiddleware` |
| [`pkg/templates`](https://pkg.go.dev/github.com/harrydayexe/GoBlog/v2/pkg/templates) | Embedded default templates (`templates.Default`) |

### Site title

```go
gen := generator.New(fsys, renderer, config.WithSiteTitle("My Blog"))
```

### Runtime environment

GoBlog surfaces the environment to all templates as `{{.Environment}}`. Set it via the `ENVIRONMENT` env var (default `"local"`, valid values: `"local"`, `"test"`, `"production"`), or pass it directly:

```go
gen := generator.New(fsys, renderer, config.WithEnvironment("production"))
```

Use it in templates to gate environment-specific markup:

```html
{{if eq .Environment "production"}}<script src="/analytics.js"></script>{{end}}
```

### Serving programmatically

`pkg/server` provides `Server`, which supports atomic handler hot-swapping — new posts can be loaded without restarting the process:

```go
cfg := config.ServerConfig{
    Server: []config.BaseServerOption{
        config.WithPort(8080),
        config.WithMiddleware(logging.New(logger)),
    },
    Gen: []config.GeneratorOption{
        config.WithSiteTitle("My Blog"),
    },
}

srv, err := server.New(logger, os.DirFS("posts/"), cfg)
if err != nil {
    log.Fatal(err)
}

// Swap in new posts atomically while running:
srv.UpdatePosts(os.DirFS("posts/"), context.Background())

// Block until SIGINT/SIGTERM:
srv.Run(context.Background())
```

### Custom output destination

Implement the `outputter.Outputter` interface to write generated content anywhere (database, S3, etc.):

```go
type Outputter interface {
    HandleGeneratedBlog(context.Context, *generator.GeneratedBlog) error
}
```

## Docker Usage

```bash
# Run blog server in container
docker run -v ./posts:/posts -p 8080:8080 goblog/goblog serve /posts
```

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## License

GNU General Public License v3.0 - see [LICENSE](LICENSE) file for details.

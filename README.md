# GoBlog

[![Go Reference](https://pkg.go.dev/badge/github.com/harrydayexe/GoBlog.svg)](https://pkg.go.dev/github.com/harrydayexe/GoBlog)
[![Go Report Card](https://goreportcard.com/badge/github.com/harrydayexe/GoBlog)](https://goreportcard.com/report/github.com/harrydayexe/GoBlog)
[![Test](https://github.com/harrydayexe/GoBlog/actions/workflows/test.yml/badge.svg?event=push)](https://github.com/harrydayexe/GoBlog/actions/workflows/test.yml)
[![License: MIT](https://img.shields.io/badge/License-GPL3-yellow.svg)](LICENSE)

GoBlog is a flexible blog generation and serving system for creating static blog feeds from Markdown files. It provides a powerful parser for Markdown content with YAML frontmatter, as well as multiple deployment options including a CLI tool, Docker image, and embeddable Go package.

The project is designed for developers who want a simple, Go-based solution for blog generation with support for modern Markdown features like syntax highlighting, footnotes, and custom templates.

## Installation

```bash
go get github.com/harrydayexe/GoBlog/v2
```

## Quick Start

### Using the Parser Package

The parser package reads Markdown files with YAML frontmatter and converts them to structured Post objects:

```go
package main

import (
    "fmt"
    "log"

    "github.com/harrydayexe/GoBlog/v2/pkg/parser"
)

func main() {
    // Create a new parser with syntax highlighting enabled
    p := parser.New(
        parser.WithCodeHighlighting(true),
        parser.WithCodeHighlightingStyle("monokai"),
    )

    // Parse a single markdown file
    post, err := p.ParseFile("posts/my-post.md")
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
posts, err := p.ParseDirectory("posts/")
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
# Generate static blog
goblog gen posts/ output/

# Serve blog locally
goblog serve posts/ --port 8080
```

## Advanced Features

### Raw Output Mode

GoBlog supports raw HTML output mode for advanced use cases where you need direct access to the generated HTML without template wrappers. This is useful when:

- Integrating GoBlog output into an existing site or CMS
- Building custom templates in your own application
- Embedding blog content in other HTML frameworks
- Processing HTML content programmatically before display

#### Using Raw Output with the CLI

```bash
# Generate raw HTML files without templates
goblog gen posts/ output/ --raw

# Or use the short flag
goblog gen posts/ output/ -r
```

When raw output mode is enabled:
- Individual post files contain only the parsed Markdown converted to HTML
- No template wrapping is applied to the content
- The `tags/` directory is not created (tag pages are skipped)
- The `index.html` file is still generated but contains raw HTML

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
    // Create generator with raw output enabled
    fsys := os.DirFS("posts/")
    gen := generator.New(fsys, config.WithRawOutput())

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
    writer.HandleGeneratedBlog(blog)
}
```

#### What to Expect from Raw Output

When `RawOutput` is enabled, the `GeneratedBlog` structure contains:

- **`Posts` map**: Keys are post slugs (derived from filenames), values are raw HTML byte slices containing only the Markdown content converted to HTML
- **`Index` field**: Contains raw HTML bytes for the index page (currently empty in raw mode)
- **`Tags` map**: Will be empty - tag pages are not generated in raw output mode

The HTML content is clean, semantic HTML generated from your Markdown without any surrounding structure like `<html>`, `<head>`, or `<body>` tags. This gives you complete control over how to integrate the content into your site.

## Docker Usage 

```bash
# Run blog server in container
docker run -v ./posts:/posts -p 8080:8080 goblog/goblog serve /posts
```

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## License

GNU General Public License v3.0 - see [LICENSE](LICENSE) file for details.

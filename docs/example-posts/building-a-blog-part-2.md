---
title: "Building a Blog, Part 2: Rendering Templates"
date: 2026-03-09T09:00:00Z
description: Wrap parsed posts in HTML templates to produce real pages.
tags:
  - go
  - tutorial
---

# Building a Blog, Part 2: Rendering Templates

Part 1 left us with posts in memory. Now they need to become pages.

## One template per page type

A blog has a handful of page types — the index, a post, a tag listing — and each
one gets its own template. Keep the shared markup (the `<head>`, the header nav,
the footer) in partials so a change lands everywhere at once.

## Pass a data struct, not a map

Give every template an explicit struct rather than a bag of values:

```go
type PostPageData struct {
    BaseData
    Post *Post
}
```

The compiler then tells you when a field a template needs has gone missing, and
the fields themselves document what a custom theme may rely on.

## What is next

Part 3 writes the rendered pages to disk.

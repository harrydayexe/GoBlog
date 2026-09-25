---
title: "Building a Blog, Part 1: Parsing Markdown"
date: 2026-03-02T09:00:00Z
description: Turn a directory of Markdown files into structured post objects.
tags:
  - go
  - tutorial
---

# Building a Blog, Part 1: Parsing Markdown

Every static site generator starts in the same place: a directory of Markdown
files that needs to become structured data. This part builds that step.

## Front matter first

A post is a YAML front matter block followed by Markdown:

```yaml
---
title: "Hello"
date: 2026-03-02T09:00:00Z
description: "My first post"
---
```

Parse the front matter into a struct, render the Markdown body to HTML, and you
have everything a template needs.

## Validate early

Reject a post with no title, date, or description while parsing rather than
letting a half-empty page reach the output directory. The error should name the
file, because the author needs to know which one to fix.

## What is next

With posts in memory, part 2 renders them as HTML pages.

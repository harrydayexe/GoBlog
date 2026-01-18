# GoBlog Project Plan

## Project Overview

GoBlog is a blog generation and serving system for creating static blog feeds from markdown files. It will be available in three forms:
1. CLI tool (with generate and serve modes)
2. Docker image for containerized serving
3. Go package/API for embedding in other applications

## Components

### 1. CLI Tool

#### `goblog gen <input-dir> <output-dir>`
- Takes a directory of `.md` files as input
- Each markdown file contains:
  - Header with metadata
  - Body content (the post)
- Generates a directory of static HTML files ready to be hosted
- Output includes:
  - Individual post pages
  - Tag pages (showing all posts for each tag)
  - `index.html` with most recent posts
- Configurable via flags
- Supports custom templates for page appearance

#### `goblog serve <input-dir>`
- Functions like `gen` but runs a local web server instead of writing files
- Serves the generated content dynamically
- Configurable via flags (port, etc.)

### 2. Docker Image

- Containerized version of `goblog serve`
- Uses volume mounts for markdown post files
- Serves blog content via web server

### 3. GoBlog Package/API

- Go library for embedding blog functionality into other applications
- Enables fully dynamic blog feeds
- Can be integrated into existing web applications
- Architecture and API design TBD

## Development Guidelines

I am using Git Worktrees to run multiple instances of claude code at once. 
You are to use the current working directory (ie the same result as `pwd`) as the parent directory.

Do not search in the GitBlog.git directory as this contains multiple children directories all containing different git states of the repo.

### IMPORTANT: Architecture Decisions

**Claude is NOT TRUSTED to make architectural decisions independently.**

When working on this project:
- ASK QUESTIONS about architecture and design decisions
- DO NOT assume implementation details
- CONFIRM approaches before implementing
- Exact designs are not finalized - clarification is required

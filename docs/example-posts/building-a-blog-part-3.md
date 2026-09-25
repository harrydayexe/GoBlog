---
title: "Building a Blog, Part 3: Writing the Output"
date: 2026-03-16T09:00:00Z
description: Write the generated site to disk, and serve it while you write.
tags:
  - go
  - tutorial
---

# Building a Blog, Part 3: Writing the Output

The last step of the pipeline turns rendered bytes into files a web server can
hand out.

## Generate in memory, write once

Render the whole site into memory first, then write it. A generator that writes
as it goes leaves a half-built directory behind when a single post fails to
parse, which is the worst possible moment to lose the previous output.

## Serve the same bytes

The fastest editing loop is one where the local server holds the same in-memory
site the file writer would have produced, and swaps it wholesale when a file
changes. No reload, no stale page, no separate code path to keep honest.

## Done

Three parts, one pipeline: parse, render, write.

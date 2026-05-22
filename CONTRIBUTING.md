# Contributing to GoBlog

Thank you for your interest in contributing! Bug reports, feature requests, and
pull requests are all welcome. If you have a question, open a
[GitHub Discussion](https://github.com/harrydayexe/GoBlog/discussions) or file
an issue.

## Reporting Bugs & Requesting Features

Open a [GitHub issue](https://github.com/harrydayexe/GoBlog/issues/new). For
bugs, include:

- GoBlog version (`goblog --version`) and Go version (`go version`)
- Steps to reproduce the problem
- What you expected vs. what actually happened

## Security Policy

**Do not open a public issue for security vulnerabilities.**

Please report them privately via
[GitHub Security Advisories](https://github.com/harrydayexe/GoBlog/security/advisories/new).
Include as much detail as possible to help reproduce the issue. You can expect
an acknowledgement within a few days.

## Project Structure

```
cmd/goblog/          CLI entry point (main package)
pkg/                 Public, importable packages
  config/            Configuration parsing
  generator/         Static site generation
  models/            Core data types (Post, etc.)
  outputter/         Output destination abstraction
  parser/            Markdown + YAML frontmatter parser
  server/            HTTP server
  templates/         Default HTML templates
internal/            Private implementation details
  errors/            Sentinel errors
  generator/         Generator internals
  logger/            Logging helpers
  server/            Server internals
  utilities/         Shared helpers
docs/example-posts/  Sample Markdown posts for local runs
.github/workflows/   CI: tests, license-header check, release
```

## Development Setup

**Prerequisites**

- Go 1.26.3+
- [`just`](https://github.com/casey/just) (task runner)
- Docker (optional, for container builds)

**Clone and build**

```bash
git clone https://github.com/harrydayexe/GoBlog.git
cd GoBlog
just build                     # produces dist/goblog
just run-serve                 # serve the example posts at localhost:8080
```

## Tests, Lint & Format

Run these before pushing:

```bash
just test            # run the test suite
just test-race       # run with the race detector
just test-coverage   # generate a coverage report
just fmt             # auto-format all Go code
just lint            # go vet + format check (mirrors CI)
just test-all        # full local CI simulation
```

## License Headers

Every source file must carry an MPL 2.0 header. CI enforces this via the
`check-license` workflow. After creating new source files, run:

```bash
just add-license     # adds missing headers via addlicense
just check-license   # verify all headers are present
```

The underlying tool is [google/addlicense](https://github.com/google/addlicense).

## Submitting a Pull Request

1. Fork the repo and branch from `main`.
2. Keep PRs focused — one logical change per PR.
3. Write a **descriptive PR title**. PRs are squash-merged, so the PR title
   becomes the commit message. Use an imperative phrase, e.g.
   *"Add RSS feed support"* rather than *"added rss"* or *"fixes #42"*.
4. Make sure CI is green. Running `just test-all` and `just check-license`
   locally before pushing covers everything CI checks.
5. **Update documentation** for any public API changes: add or update godoc
   comments on exported symbols, update the relevant `README.md` sections, and
   update any `doc.go` package-level entries that reference available options or
   features.

## License

By contributing, you agree that your contributions will be licensed under the
project's [Mozilla Public License 2.0](LICENSE).

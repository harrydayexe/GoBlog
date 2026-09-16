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

The repository holds three Go modules:

```
go.mod               github.com/harrydayexe/GoBlog/v2        (the public library)
  pkg/               Public, importable packages
cli/go.mod           github.com/harrydayexe/GoBlog/v2/cli    (never published)
  cli/cmd/goblog/    CLI entry point (main package)
  cli/internal/      Private CLI implementation details
integration/go.mod   github.com/harrydayexe/GoBlog/v2/integration
docs/example-posts/  Sample Markdown posts for local runs
.github/workflows/   CI: tests, license-header check, release
```

The CLI lives in its own leaf module so that anything it imports stays out of
the library's dependency graph. Both `cli/` and `integration/` use
`replace github.com/harrydayexe/GoBlog/v2 => ../`, so they always build against
the library at the current commit — no tag or release is needed in between.

Because `go test ./...` stops at a nested `go.mod` boundary, commands run from
the repo root cover the library only. The `just` recipes below iterate over
every module, and CI runs a job per module.

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
just test            # run the unit test suite
just test-race       # run with the race detector
just test-coverage   # generate a coverage report
just fmt             # auto-format all Go code
just lint            # go vet + format check (mirrors CI)
just test-all        # full local CI simulation
```

### Integration tests

The `integration/` directory is a separate Go module containing black-box tests
that boot the CLI and Docker image against a real environment using
[testcontainers-go](https://golang.testcontainers.org/). **Docker is required.**

```bash
just test-integration      # run integration tests with Docker
```

Or run them directly:

```bash
cd integration && go test -v -timeout 10m ./...
```

The unit suite (`just test`) covers the library and CLI modules but
deliberately excludes the integration module, so unit feedback stays fast in
CI even when integration tests are slow.

In CI the integration tests run in a dedicated `integration` job so the two
stages report separately.

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

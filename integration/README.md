# GoBlog Integration Tests

Black-box integration tests for GoBlog using
[testcontainers-go](https://golang.testcontainers.org/). This is a separate Go
module so testcontainers' Docker dependency tree does not enter the published
library's `go.mod`.

## Prerequisites

- Go 1.26.3+
- Docker (running)

## Running

From the repository root:

```bash
just test-integration
```

Or directly:

```bash
cd integration && go test -v -timeout 10m ./...
```

Tests that require Docker are skipped automatically when no Docker daemon is
reachable, so the in-process lifecycle tests (`TestRun_*`) still run in
Docker-less environments.

## What's tested

| Test | Kind | What it covers |
|---|---|---|
| `TestRun_BindError` | in-process | `Server.Run` returns a bind error when the port is already occupied |
| `TestRun_GracefulShutdown` | in-process | Context cancellation causes clean shutdown within the 10 s window |
| `TestServe_Smoke` | container | Docker image starts and serves HTTP 200 (Docker distribution channel) |
| `TestServe_LiveReload` | container | Watcher detects a file change and the running server reflects it |
| `TestServe_BlogRootFlag` | container | `-p` flag correctly prefixes all links with the configured blog root |

## CI

The integration tests run in a dedicated `integration` job in
`.github/workflows/test.yml`, separate from the fast unit `test` job.

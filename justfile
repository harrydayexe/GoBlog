# GoBlog Project Justfile
# Convenient recipes for building, testing, and managing the GoBlog CLI application

# Variables
BINARY_NAME := "goblog"
MODULE := "github.com/harrydayexe/GoBlog/v2"
CLI_DIR := "cli"
MAIN_PATH := "./cmd/goblog"
DIST_DIR := "dist"
COVERAGE_DIR := "coverage"

# Modules covered by the unit suite. The integration module is excluded because
# it needs Docker and is run separately by `test-integration`.
UNIT_MODULES := ". cli"
ALL_MODULES := ". cli integration"

# Build-time version injection using git tags
VERSION := `git describe --tags --always --dirty 2>/dev/null || echo "dev"`
LDFLAGS := '-s -w -X main.version=' + VERSION

# Build binary for current OS/architecture
[default]
[group("build")]
build:
    @echo "Building {{BINARY_NAME}}..."
    @mkdir -p {{DIST_DIR}}
    go -C {{CLI_DIR}} build -ldflags "{{LDFLAGS}}" -o {{justfile_directory()}}/{{DIST_DIR}}/{{BINARY_NAME}} {{MAIN_PATH}}
    @echo "✓ Binary built successfully: {{DIST_DIR}}/{{BINARY_NAME}}"

# Build and install to $GOPATH/bin
install:
    @echo "Installing {{BINARY_NAME}}..."
    go -C {{CLI_DIR}} install -ldflags "{{LDFLAGS}}" {{MAIN_PATH}}
    @echo "✓ Binary installed successfully"

# Remove build artifacts
[group("dev")]
clean:
    @echo "Cleaning build artifacts..."
    rm -rf {{DIST_DIR}}
    rm -rf {{COVERAGE_DIR}}
    go clean -cache
    @echo "✓ Clean complete"

# Run all tests
[group("test")]
test:
    #!/usr/bin/env bash
    set -euo pipefail
    for mod in {{UNIT_MODULES}}; do
        echo "==> $mod"
        go -C "$mod" test ./...
    done

# Run tests with verbose output
[group("test")]
test-verbose:
    #!/usr/bin/env bash
    set -euo pipefail
    for mod in {{UNIT_MODULES}}; do
        echo "==> $mod"
        go -C "$mod" test -v ./...
    done

# Run tests with race detector
[group("test")]
test-race:
    #!/usr/bin/env bash
    set -euo pipefail
    for mod in {{UNIT_MODULES}}; do
        echo "==> $mod"
        go -C "$mod" test -race ./...
    done

# Run integration tests (requires Docker)
[group("test")]
test-integration:
    go -C integration test -v -timeout 10m ./...

# Run tests with coverage profile
[group("test")]
test-coverage:
    # The two modules produce separate profiles: `go tool cover` resolves source
    # files through the module it runs in, so a merged cross-module profile
    # would not be readable.
    @echo "Running tests with coverage..."
    @mkdir -p {{COVERAGE_DIR}}
    go test -coverprofile={{COVERAGE_DIR}}/library.out ./...
    go -C {{CLI_DIR}} test -coverprofile={{justfile_directory()}}/{{COVERAGE_DIR}}/cli.out ./...
    @echo "\nCoverage summary:"
    @printf 'library  ' && go tool cover -func={{COVERAGE_DIR}}/library.out | tail -1
    @printf 'cli      ' && go -C {{CLI_DIR}} tool cover -func={{justfile_directory()}}/{{COVERAGE_DIR}}/cli.out | tail -1

# Generate HTML coverage reports
[group("test")]
coverage-html: test-coverage
    @echo "Generating HTML coverage reports..."
    go tool cover -html={{COVERAGE_DIR}}/library.out -o {{COVERAGE_DIR}}/library.html
    go -C {{CLI_DIR}} tool cover -html={{justfile_directory()}}/{{COVERAGE_DIR}}/cli.out -o {{justfile_directory()}}/{{COVERAGE_DIR}}/cli.html
    @echo "✓ Coverage reports: {{COVERAGE_DIR}}/library.html, {{COVERAGE_DIR}}/cli.html"
    @echo "Opening in browser..."
    @open {{COVERAGE_DIR}}/library.html 2>/dev/null || xdg-open {{COVERAGE_DIR}}/library.html 2>/dev/null || echo "Please open {{COVERAGE_DIR}}/library.html manually"

# Run complete test suite (CI/CD simulation)
[group("test")]
test-all:
    @echo "Running complete test suite (CI/CD workflow)..."
    @echo "\n=== Stage 1: go vet ==="
    @just vet
    @echo "\n=== Stage 2: go test -v ==="
    @just test-verbose
    @echo "\n=== Stage 3: go test -race ==="
    @just test-race
    @echo "\n✓ All tests passed!"

# Run go vet linter
[group("lint")]
vet:
    #!/usr/bin/env bash
    set -euo pipefail
    for mod in {{ALL_MODULES}}; do
        echo "==> $mod"
        go -C "$mod" vet ./...
    done

# Format all Go code
[group("lint")]
fmt:
    #!/usr/bin/env bash
    set -euo pipefail
    for mod in {{ALL_MODULES}}; do
        go -C "$mod" fmt ./...
    done

# Run vulncheck on codebase
[group("lint")]
vulncheck:
    #!/usr/bin/env bash
    set -euo pipefail
    for mod in {{ALL_MODULES}}; do
        echo "==> $mod"
        (cd "$mod" && govulncheck ./...)
    done

# Run go mod tidy
[group("lint")]
mod-tidy:
    #!/usr/bin/env bash
    set -euo pipefail
    for mod in {{ALL_MODULES}}; do
        go -C "$mod" mod tidy
    done

# Check if code is formatted
[group("lint")]
fmt-check:
    @echo "Checking code formatting..."
    @test -z "$(gofmt -l .)" || (echo "Code is not formatted. Run 'just fmt'" && exit 1)
    @echo "✓ Code is properly formatted"

# Run all linting checks
[group("lint")]
lint: mod-tidy vet fmt-check check-license

# Check license headers exist
[group("lint")]
check-license:
    addlicense -check ./

# Add license headers
[group("lint")]
add-license:
    addlicense -l mpl -c "GoBlog Authors" ./

# Run generator command with arguments. Paths are relative to the repo root.
[group('run')]
run-gen *ARGS: build
    {{DIST_DIR}}/{{BINARY_NAME}} generate {{ARGS}}

# Run serve command with optional arguments (defaults to example posts and their series)
[group('run')]
run-serve *ARGS: build
    #!/usr/bin/env bash
    if [ -z "{{ARGS}}" ]; then
        {{DIST_DIR}}/{{BINARY_NAME}} serve docs/example-posts --series-file docs/example-posts/series.yml
    else
        {{DIST_DIR}}/{{BINARY_NAME}} serve {{ARGS}}
    fi

# Build the dockerfile for the current architecture
[group("build")]
docker tag="goblog:latest":
    @echo "Building Docker image..."
    docker build -t {{tag}} .
    @echo "✓ Docker image built successfully"

[group("run")]
run-image tag="goblog:latest": docker
    @echo "Running Docker image..."
    docker run -v ./docs/example-posts/:/posts -p 8080:8080 {{tag}} /posts --series-file /posts/series.yml

# GoBlog Project Justfile
# Convenient recipes for building, testing, and managing the GoBlog CLI application

# Variables
BINARY_NAME := "goblog"
MODULE := "github.com/harrydayexe/GoBlog/v2"
MAIN_PATH := "./cmd/goblog"
DIST_DIR := "dist"
COVERAGE_DIR := "coverage"

# Build-time version injection using git commands
VERSION := `git describe --tags --always --dirty 2>/dev/null || echo "dev"`
COMMIT := `git rev-parse --short HEAD 2>/dev/null || echo "none"`
DATE := `date -u +%Y-%m-%dT%H:%M:%SZ`
LDFLAGS := '-s -w -X main.version=' + VERSION + ' -X main.commit=' + COMMIT + ' -X main.date=' + DATE

# Build binary for current OS/architecture
[default]
[group("dev")]
build:
    @echo "Building {{BINARY_NAME}}..."
    @mkdir -p {{DIST_DIR}}
    go build -ldflags "{{LDFLAGS}}" -o {{DIST_DIR}}/{{BINARY_NAME}} {{MAIN_PATH}}
    @echo "✓ Binary built successfully: {{DIST_DIR}}/{{BINARY_NAME}}"

# Build and install to $GOPATH/bin
install:
    @echo "Installing {{BINARY_NAME}}..."
    go install -ldflags "{{LDFLAGS}}" {{MAIN_PATH}}
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
[group("dev")]
test:
    go test ./...

# Run tests with verbose output
[group("dev")]
test-verbose:
    go test -v ./...

# Run tests with race detector
[group("dev")]
test-race:
    go test -race ./...

# Run tests with coverage profile
[group("dev")]
test-coverage:
    @echo "Running tests with coverage..."
    @mkdir -p {{COVERAGE_DIR}}
    go test -coverprofile={{COVERAGE_DIR}}/coverage.out ./...
    @echo "\nCoverage summary:"
    @go tool cover -func={{COVERAGE_DIR}}/coverage.out | tail -1

# Generate HTML coverage report
[group("dev")]
coverage-html: test-coverage
    @echo "Generating HTML coverage report..."
    go tool cover -html={{COVERAGE_DIR}}/coverage.out -o {{COVERAGE_DIR}}/coverage.html
    @echo "✓ Coverage report: {{COVERAGE_DIR}}/coverage.html"
    @echo "Opening in browser..."
    @open {{COVERAGE_DIR}}/coverage.html 2>/dev/null || xdg-open {{COVERAGE_DIR}}/coverage.html 2>/dev/null || echo "Please open {{COVERAGE_DIR}}/coverage.html manually"

# Run complete test suite (CI/CD simulation)
[group("dev")]
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
    go vet ./...

# Format all Go code
[group("lint")]
fmt:
    go fmt ./...

# Check if code is formatted
[group("lint")]
fmt-check:
    @echo "Checking code formatting..."
    @test -z "$(gofmt -l .)" || (echo "Code is not formatted. Run 'just fmt'" && exit 1)
    @echo "✓ Code is properly formatted"

# Run all linting checks
[group("lint")]
lint: vet fmt-check

# Run generator command with arguments
[group('run')]
run-gen *ARGS:
    go run {{MAIN_PATH}} gen {{ARGS}}

# Run serve command with optional arguments (defaults to example posts)
[group('run')]
run-serve *ARGS:
    #!/usr/bin/env bash
    if [ -z "{{ARGS}}" ]; then
        go run {{MAIN_PATH}} serve docs/example-posts
    else
        go run {{MAIN_PATH}} serve {{ARGS}}
    fi

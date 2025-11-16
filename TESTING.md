# Testing Guide for GoBlog

## Overview

GoBlog follows Test-Driven Development (TDD) principles with comprehensive unit test coverage. This document describes the testing strategy, how to run tests, and best practices for writing new tests.

## Test Coverage

Current test coverage: **88.6%**

### Module Coverage

- `internal/gen/config`: 100%
- `internal/gen/log`: 100%
- `pkg/models`: 100%
- `internal/gen/parser`: 94.3%
- `internal/gen/template`: 83.1%
- `internal/gen/generator`: 78.8%

Coverage reports are automatically generated and uploaded to Codecov on every push and pull request.

## Testing Philosophy

### TDD Principles

1. **Test inputs and outputs, not implementation details**
   - Tests should validate behavior, not internal structures
   - Focus on the public API of each module
   - Avoid testing private functions directly

2. **Isolation**
   - Each test should be independent and isolated
   - Use interfaces for dependency injection
   - Mock external dependencies (filesystem, network, etc.)

3. **Composability**
   - Break down complex logic into smaller, testable units
   - Use interfaces to make components easily mockable
   - Prefer small, focused functions over large monolithic ones

## Running Tests

### Run all tests

```bash
go test ./...
```

### Run tests with coverage

```bash
go test -coverprofile=coverage.out -covermode=atomic ./internal/gen/... ./pkg/models/...
```

### View coverage report

```bash
# Terminal view
go tool cover -func=coverage.out

# HTML view
go tool cover -html=coverage.out
```

### Run tests with race detection

```bash
go test -race ./...
```

### Run specific package tests

```bash
# Test specific package
go test ./internal/gen/config/... -v

# Test specific function
go test ./internal/gen/config/... -run TestConfig_Validate -v
```

## Test Structure

### File Organization

Test files are colocated with the code they test:

```
internal/gen/config/
├── config.go
└── config_test.go
```

### Test Naming Convention

- Test files: `*_test.go`
- Test functions: `Test<FunctionName>` or `Test<Type>_<Method>`
- Examples:
  - `TestParseConfig`
  - `TestConfig_Validate`
  - `TestPost_GenerateSlug`

### Test Structure Pattern

```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name      string
        input     InputType
        expected  OutputType
        expectErr bool
    }{
        {
            name:      "descriptive test case name",
            input:     someInput,
            expected:  expectedOutput,
            expectErr: false,
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := FunctionUnderTest(tt.input)

            if tt.expectErr {
                if err == nil {
                    t.Error("expected error but got nil")
                }
            } else {
                if err != nil {
                    t.Errorf("unexpected error: %v", err)
                }
                if result != tt.expected {
                    t.Errorf("expected %v, got %v", tt.expected, result)
                }
            }
        })
    }
}
```

## Testing Best Practices

### 1. Use Table-Driven Tests

Table-driven tests make it easy to add new test cases and keep tests maintainable:

```go
func TestSlugify(t *testing.T) {
    tests := []struct {
        input    string
        expected string
    }{
        {"Hello World", "hello-world"},
        {"UPPERCASE", "uppercase"},
        {"Special!@#$%Characters", "specialcharacters"},
    }

    for _, tt := range tests {
        t.Run(tt.input, func(t *testing.T) {
            result := slugify(tt.input)
            if result != tt.expected {
                t.Errorf("slugify(%q) = %q, want %q", tt.input, result, tt.expected)
            }
        })
    }
}
```

### 2. Use Test Helpers

Create helper functions to reduce boilerplate:

```go
func createTestPost(t *testing.T, title string) *models.Post {
    t.Helper()
    return &models.Post{
        Title:       title,
        Date:        time.Now(),
        Description: "Test post",
    }
}
```

### 3. Use t.TempDir() for Filesystem Tests

Always use `t.TempDir()` for tests that need temporary directories:

```go
func TestFileOperation(t *testing.T) {
    tmpDir := t.TempDir() // Automatically cleaned up
    // Use tmpDir for file operations
}
```

### 4. Use Subtests for Better Organization

```go
func TestComplexFeature(t *testing.T) {
    t.Run("scenario 1", func(t *testing.T) {
        // Test scenario 1
    })

    t.Run("scenario 2", func(t *testing.T) {
        // Test scenario 2
    })
}
```

### 5. Test Error Cases

Always test both success and error paths:

```go
tests := []struct {
    name      string
    input     string
    expectErr bool
    errText   string
}{
    {
        name:      "valid input",
        input:     "valid",
        expectErr: false,
    },
    {
        name:      "invalid input",
        input:     "invalid",
        expectErr: true,
        errText:   "expected error message",
    },
}
```

## Dependency Injection for Testing

### Logger Interface

The logger has been refactored to use an interface, making it easy to inject test loggers:

```go
func TestWithLogger(t *testing.T) {
    var stdout, stderr bytes.Buffer
    logger := log.NewTestLogger("TEST", false, &stdout, &stderr)

    // Use logger in your test
    parser := parser.New(logger)
    // ...
}
```

### Configuration

Configuration can be created programmatically or from byte slices:

```go
func TestWithConfig(t *testing.T) {
    yaml := `
input_folder: ./posts
output_folder: ./site
`
    cfg, err := config.ParseConfigFromBytes([]byte(yaml))
    if err != nil {
        t.Fatalf("failed to parse config: %v", err)
    }
    // Use cfg in your test
}
```

## Continuous Integration

Tests run automatically on:
- Every push to main/master/develop branches
- Every pull request

The CI pipeline:
1. Runs all tests with race detection
2. Generates coverage reports
3. Uploads coverage to Codecov
4. Runs linting (golangci-lint)
5. Builds the binaries

## Coverage Requirements

- **Project minimum**: 80%
- **Patch minimum**: 80%
- **Pull requests** that decrease coverage by more than 2% will fail CI

## Writing New Tests

When adding new features:

1. Write tests first (TDD approach)
2. Ensure all public functions are tested
3. Test edge cases and error conditions
4. Aim for 100% coverage of new code
5. Use interfaces for external dependencies
6. Keep tests focused and isolated

## Common Testing Patterns

### Testing File Operations

```go
func TestFileOperation(t *testing.T) {
    tmpDir := t.TempDir()
    testFile := filepath.Join(tmpDir, "test.txt")

    if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
        t.Fatalf("failed to create test file: %v", err)
    }

    // Test your file operation
}
```

### Testing with Goroutines

```go
func TestConcurrent(t *testing.T) {
    t.Parallel() // Run in parallel with other tests

    var wg sync.WaitGroup
    // Your concurrent test code
    wg.Wait()
}
```

### Testing Time-Dependent Code

```go
func TestWithTime(t *testing.T) {
    now := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
    post := &models.Post{Date: now}

    if post.FormattedDate() != "March 15, 2024" {
        t.Errorf("unexpected formatted date")
    }
}
```

## Debugging Test Failures

### Verbose Output

```bash
go test -v ./...
```

### Run Specific Test

```bash
go test -run TestSpecificFunction -v
```

### Print Debug Information

```go
func TestDebug(t *testing.T) {
    result := someFunction()
    t.Logf("Debug: result = %+v", result) // Only shown if test fails
}
```

## Test Containers (Future)

For integration tests requiring external services (databases, message queues, etc.), we use testcontainers. This ensures tests are:
- Reproducible
- Isolated
- Don't require manual setup

Example:
```go
// Future use when implementing integration tests
func TestWithDatabase(t *testing.T) {
    ctx := context.Background()
    pgContainer, err := postgres.RunContainer(ctx,
        testcontainers.WithImage("postgres:15"),
        // ... configuration
    )
    if err != nil {
        t.Fatalf("failed to start container: %v", err)
    }
    defer pgContainer.Terminate(ctx)

    // Run tests with database
}
```

## Resources

- [Go Testing Package](https://pkg.go.dev/testing)
- [Testcontainers for Go](https://golang.testcontainers.org/)
- [Codecov Documentation](https://docs.codecov.com/)
- [Go Best Practices for Testing](https://go.dev/doc/tutorial/add-a-test)

## Feedback and Improvements

If you have suggestions for improving our testing strategy, please:
1. Open an issue with the `testing` label
2. Submit a pull request with your proposed changes
3. Include examples and rationale

---

**Remember**: Good tests are an investment in code quality, maintainability, and developer productivity. Write tests that you'd want to debug six months from now!

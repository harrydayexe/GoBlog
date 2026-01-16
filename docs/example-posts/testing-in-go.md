---
title: Comprehensive Guide to Testing in Go
date: 2026-01-05T11:20:00Z
description: Master testing in Go with unit tests, table-driven tests, benchmarks, and mocking strategies
tags:
  - go
  - testing
  - best-practices
  - tutorial
---

# Comprehensive Guide to Testing in Go

Testing is a first-class citizen in Go. The built-in `testing` package provides everything you need to write comprehensive tests.

## Basic Unit Tests

Test files end with `_test.go` and test functions start with `Test`:

```go
// math.go
package math

func Add(a, b int) int {
    return a + b
}

// math_test.go
package math

import "testing"

func TestAdd(t *testing.T) {
    result := Add(2, 3)
    expected := 5

    if result != expected {
        t.Errorf("Add(2, 3) = %d; want %d", result, expected)
    }
}
```

## Table-Driven Tests

The idiomatic Go way to test multiple scenarios:

```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name     string
        a, b     int
        expected int
    }{
        {"positive numbers", 2, 3, 5},
        {"negative numbers", -2, -3, -5},
        {"mixed signs", -2, 3, 1},
        {"zeros", 0, 0, 0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Add(tt.a, tt.b)
            if result != tt.expected {
                t.Errorf("got %d, want %d", result, tt.expected)
            }
        })
    }
}
```

## Test Helpers

Use helper functions to reduce duplication:

```go
func assertEqual(t *testing.T, got, want int) {
    t.Helper() // Marks this as a helper function
    if got != want {
        t.Errorf("got %d, want %d", got, want)
    }
}

func TestSomething(t *testing.T) {
    assertEqual(t, Add(2, 3), 5)
}
```

## Benchmarks

Measure performance with benchmark tests:

```go
func BenchmarkAdd(b *testing.B) {
    for i := 0; i < b.N; i++ {
        Add(2, 3)
    }
}
```

Run with `go test -bench=.`

## Test Coverage

Check your test coverage:

```bash
go test -cover
go test -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Mocking

Go doesn't have built-in mocking, but interfaces make it easy:

```go
type UserRepository interface {
    GetUser(id int) (*User, error)
}

type MockUserRepository struct {
    GetUserFunc func(id int) (*User, error)
}

func (m *MockUserRepository) GetUser(id int) (*User, error) {
    return m.GetUserFunc(id)
}

func TestUserService(t *testing.T) {
    mock := &MockUserRepository{
        GetUserFunc: func(id int) (*User, error) {
            return &User{ID: id, Name: "Test"}, nil
        },
    }

    service := NewUserService(mock)
    // Test using mock...
}
```

## Test Fixtures

Use test fixtures for complex setup:

```go
func TestMain(m *testing.M) {
    // Setup
    setup()

    // Run tests
    code := m.Run()

    // Teardown
    teardown()

    os.Exit(code)
}
```

## Integration Tests

Use build tags to separate integration tests:

```go
//go:build integration
// +build integration

package myapp

func TestDatabaseIntegration(t *testing.T) {
    // Integration test code
}
```

Run with: `go test -tags=integration`

Testing is fundamental to writing reliable Go code. Master these patterns and your code will be robust and maintainable!

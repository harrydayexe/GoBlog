package server_test

import (
	"context"
	"io/fs"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/harrydayexe/GoBlog/v2/pkg/config"
	"github.com/harrydayexe/GoBlog/v2/pkg/server"
)

// TestServerWithoutMiddleware verifies that servers without middleware work correctly
// (backward compatibility).
func TestServerWithoutMiddleware(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	postsFS := createTestFS(t)

	cfg := config.ServerConfig{
		Server: []config.BaseServerOption{
			config.WithPort(8080),
		},
		Gen: []config.GeneratorOption{
			config.WithRawOutput(),
		},
	}

	srv, err := server.New(logger, postsFS, cfg)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	// Make a test request
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

// TestServerWithSingleMiddleware tests that a single middleware is correctly applied.
func TestServerWithSingleMiddleware(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	postsFS := createTestFS(t)

	// Track if middleware was called
	var middlewareCalled bool
	testMiddleware := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			middlewareCalled = true
			w.Header().Set("X-Test-Middleware", "called")
			h.ServeHTTP(w, r)
		})
	}

	cfg := config.ServerConfig{
		Server: []config.BaseServerOption{
			config.WithPort(8080),
			config.WithMiddleware(testMiddleware),
		},
		Gen: []config.GeneratorOption{
			config.WithRawOutput(),
		},
	}

	srv, err := server.New(logger, postsFS, cfg)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	// Make a test request
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if !middlewareCalled {
		t.Error("middleware was not called")
	}

	if w.Header().Get("X-Test-Middleware") != "called" {
		t.Error("middleware did not set expected header")
	}
}

// TestServerWithMultipleMiddleware tests that multiple middleware are chained correctly.
func TestServerWithMultipleMiddleware(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	postsFS := createTestFS(t)

	// Track middleware execution
	var executionOrder []string

	firstMiddleware := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			executionOrder = append(executionOrder, "first-before")
			w.Header().Set("X-First", "true")
			h.ServeHTTP(w, r)
			executionOrder = append(executionOrder, "first-after")
		})
	}

	secondMiddleware := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			executionOrder = append(executionOrder, "second-before")
			w.Header().Set("X-Second", "true")
			h.ServeHTTP(w, r)
			executionOrder = append(executionOrder, "second-after")
		})
	}

	cfg := config.ServerConfig{
		Server: []config.BaseServerOption{
			config.WithPort(8080),
			config.WithMiddleware(firstMiddleware, secondMiddleware),
		},
		Gen: []config.GeneratorOption{
			config.WithRawOutput(),
		},
	}

	srv, err := server.New(logger, postsFS, cfg)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	// Make a test request
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	// Verify both middleware were called
	if w.Header().Get("X-First") != "true" {
		t.Error("first middleware did not set header")
	}
	if w.Header().Get("X-Second") != "true" {
		t.Error("second middleware did not set header")
	}

	// Verify execution order: first is outermost, so it executes first before request
	// and last after request
	expectedOrder := []string{
		"first-before",
		"second-before",
		"second-after",
		"first-after",
	}

	if len(executionOrder) != len(expectedOrder) {
		t.Fatalf("unexpected execution order length: got %d, want %d", len(executionOrder), len(expectedOrder))
	}

	for i, expected := range expectedOrder {
		if executionOrder[i] != expected {
			t.Errorf("execution order[%d]: got %s, want %s", i, executionOrder[i], expected)
		}
	}
}

// TestMiddlewarePersistsAcrossUpdates verifies that middleware continues to work
// after UpdatePosts() is called.
func TestMiddlewarePersistsAcrossUpdates(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	postsFS := createTestFS(t)

	var callCount int
	countingMiddleware := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			h.ServeHTTP(w, r)
		})
	}

	cfg := config.ServerConfig{
		Server: []config.BaseServerOption{
			config.WithPort(8080),
			config.WithMiddleware(countingMiddleware),
		},
		Gen: []config.GeneratorOption{
			config.WithRawOutput(),
		},
	}

	srv, err := server.New(logger, postsFS, cfg)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	// Make first request
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	w1 := httptest.NewRecorder()
	srv.ServeHTTP(w1, req1)

	if callCount != 1 {
		t.Errorf("after first request: expected callCount=1, got %d", callCount)
	}

	// Update posts
	updatedFS := createTestFS(t)
	if err := srv.UpdatePosts(updatedFS, context.Background()); err != nil {
		t.Fatalf("failed to update posts: %v", err)
	}

	// Make second request after update
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	w2 := httptest.NewRecorder()
	srv.ServeHTTP(w2, req2)

	if callCount != 2 {
		t.Errorf("after second request: expected callCount=2, got %d", callCount)
	}
}

// TestMultipleWithMiddlewareCalls tests that multiple WithMiddleware calls
// correctly append to the middleware chain.
func TestMultipleWithMiddlewareCalls(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	postsFS := createTestFS(t)

	firstMiddleware := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-First", "true")
			h.ServeHTTP(w, r)
		})
	}

	secondMiddleware := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Second", "true")
			h.ServeHTTP(w, r)
		})
	}

	cfg := config.ServerConfig{
		Server: []config.BaseServerOption{
			config.WithPort(8080),
			config.WithMiddleware(firstMiddleware),
			config.WithMiddleware(secondMiddleware),
		},
		Gen: []config.GeneratorOption{
			config.WithRawOutput(),
		},
	}

	srv, err := server.New(logger, postsFS, cfg)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	// Make a test request
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	// Both middleware should have been applied
	if w.Header().Get("X-First") != "true" {
		t.Error("first middleware did not set header")
	}
	if w.Header().Get("X-Second") != "true" {
		t.Error("second middleware did not set header")
	}
}

// ExampleServer_withMiddleware demonstrates using middleware with the server.
func ExampleServer_withMiddleware() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	postsFS := fstest.MapFS{
		"post1.md": &fstest.MapFile{
			Data: []byte("---\ntitle: Test Post\ndescription: A test post\ndate: 2024-01-01\n---\nContent"),
		},
	}

	// Custom middleware that adds a header
	customMiddleware := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Powered-By", "GoBlog")
			h.ServeHTTP(w, r)
		})
	}

	cfg := config.ServerConfig{
		Server: []config.BaseServerOption{
			config.WithPort(8080),
			config.WithMiddleware(customMiddleware),
		},
	}

	srv, err := server.New(logger, postsFS, cfg)
	if err != nil {
		logger.Error("failed to create server", "error", err)
		return
	}

	// Server is ready with middleware applied
	_ = srv
}

// createTestFS creates a minimal test filesystem with a single post.
func createTestFS(t *testing.T) fs.FS {
	t.Helper()
	return fstest.MapFS{
		"test-post.md": &fstest.MapFile{
			Data: []byte(strings.TrimSpace(`
---
title: Test Post
description: A test post for middleware testing
date: 2024-01-01
---

# Test Content

This is a test post.
			`)),
		},
	}
}

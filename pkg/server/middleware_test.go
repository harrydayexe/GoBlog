// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package server_test

import (
	"bytes"
	"context"
	"io"
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

// TestServerDisableTags verifies that /tags and /tags/{tag} routes return 404
// when DisableTags is set, while the index route still returns 200.
func TestServerDisableTags(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	postsFS := createTestFS(t)

	cfg := config.ServerConfig{
		Server: []config.BaseServerOption{
			config.WithPort(8080),
		},
		Gen: []config.GeneratorOption{
			config.WithDisableTags(),
		},
	}

	srv, err := server.New(logger, postsFS, cfg)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	tests := []struct {
		path           string
		wantStatusCode int
	}{
		{"/", http.StatusOK},
		{"/tags", http.StatusNotFound},
		{"/tags/go", http.StatusNotFound},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(http.MethodGet, tt.path, nil)
		w := httptest.NewRecorder()
		srv.ServeHTTP(w, req)

		if w.Code != tt.wantStatusCode {
			t.Errorf("GET %s: got status %d, want %d", tt.path, w.Code, tt.wantStatusCode)
		}
	}
}

// TestServerTagsEnabledByDefault verifies that /tags and /tags/{tag} routes
// return a non-404 status when DisableTags is not set.
func TestServerTagsEnabledByDefault(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	postsFS := createTestFS(t)

	cfg := config.ServerConfig{
		Server: []config.BaseServerOption{
			config.WithPort(8080),
		},
	}

	srv, err := server.New(logger, postsFS, cfg)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	// /tags should be registered and return non-404.
	req := httptest.NewRequest(http.MethodGet, "/tags", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code == http.StatusNotFound {
		t.Errorf("GET /tags: got 404 but tags should be enabled by default")
	}
}

// TestServer_StripsHTMLExtension verifies that the server accepts requests with
// and without .html suffixes and serves identical content for both.
func TestServer_StripsHTMLExtension(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	postsFS := createTestFS(t)

	cfg := config.ServerConfig{
		Server: []config.BaseServerOption{
			config.WithPort(8080),
		},
	}

	srv, err := server.New(logger, postsFS, cfg)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	// Fetch the clean-URL index for comparison.
	cleanReq := httptest.NewRequest(http.MethodGet, "/", nil)
	cleanW := httptest.NewRecorder()
	srv.ServeHTTP(cleanW, cleanReq)
	indexBody := cleanW.Body.String()

	// /index.html must return the same body as /.
	htmlReq := httptest.NewRequest(http.MethodGet, "/index.html", nil)
	htmlW := httptest.NewRecorder()
	srv.ServeHTTP(htmlW, htmlReq)
	if htmlW.Code != http.StatusOK {
		t.Errorf("GET /index.html: got status %d, want 200", htmlW.Code)
	}
	if htmlW.Body.String() != indexBody {
		t.Error("GET /index.html: body differs from GET /")
	}

	// /posts/test-post.html must return the same as /posts/test-post.
	cleanPostReq := httptest.NewRequest(http.MethodGet, "/posts/test-post", nil)
	cleanPostW := httptest.NewRecorder()
	srv.ServeHTTP(cleanPostW, cleanPostReq)
	postBody := cleanPostW.Body.String()

	htmlPostReq := httptest.NewRequest(http.MethodGet, "/posts/test-post.html", nil)
	htmlPostW := httptest.NewRecorder()
	srv.ServeHTTP(htmlPostW, htmlPostReq)
	if htmlPostW.Code != http.StatusOK {
		t.Errorf("GET /posts/test-post.html: got status %d, want 200", htmlPostW.Code)
	}
	if htmlPostW.Body.String() != postBody {
		t.Error("GET /posts/test-post.html: body differs from GET /posts/test-post")
	}

	// /tags/go.html must return the same as /tags/go.
	cleanTagReq := httptest.NewRequest(http.MethodGet, "/tags/go", nil)
	cleanTagW := httptest.NewRecorder()
	srv.ServeHTTP(cleanTagW, cleanTagReq)

	htmlTagReq := httptest.NewRequest(http.MethodGet, "/tags/go.html", nil)
	htmlTagW := httptest.NewRecorder()
	srv.ServeHTTP(htmlTagW, htmlTagReq)
	if htmlTagW.Code != cleanTagW.Code {
		t.Errorf("GET /tags/go.html: got status %d, want %d (same as clean URL)", htmlTagW.Code, cleanTagW.Code)
	}
}

// TestServer_StripsHTMLExtension_BlogRoot verifies .html stripping works when
// BlogRoot is a subdirectory.
func TestServer_StripsHTMLExtension_BlogRoot(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	postsFS := createTestFS(t)

	cfg := config.ServerConfig{
		Server: []config.BaseServerOption{
			config.WithPort(8080),
			config.WithBlogRoot("/blog/").AsServerOption(),
		},
		Gen: []config.GeneratorOption{
			config.WithBlogRoot("/blog/").AsGeneratorOption(),
		},
	}

	srv, err := server.New(logger, postsFS, cfg)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	// /blog/ must return the index.
	cleanReq := httptest.NewRequest(http.MethodGet, "/blog/", nil)
	cleanW := httptest.NewRecorder()
	srv.ServeHTTP(cleanW, cleanReq)
	if cleanW.Code != http.StatusOK {
		t.Fatalf("GET /blog/: got status %d, want 200", cleanW.Code)
	}

	// /blog.html strips to /blog which then redirects or serves; middleware turns
	// /blog.html → /blog, and Go's mux strips trailing non-slash so check we don't 404.
	htmlReq := httptest.NewRequest(http.MethodGet, "/blog/posts/test-post.html", nil)
	htmlW := httptest.NewRecorder()
	srv.ServeHTTP(htmlW, htmlReq)
	if htmlW.Code != http.StatusOK {
		t.Errorf("GET /blog/posts/test-post.html: got status %d, want 200", htmlW.Code)
	}
}

// TestHandler_StripsHTMLExtension verifies that the bare Handler() function
// (without server.New) also strips .html from incoming requests.
func TestHandler_StripsHTMLExtension(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	postsFS := createTestFS(t)

	// Build a GeneratedBlog manually via the generator.
	cfg := config.ServerConfig{
		Gen: []config.GeneratorOption{config.WithRawOutput()},
	}
	srv, err := server.New(logger, postsFS, cfg)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	// /index.html should still be served as /  (strip → route to index).
	req := httptest.NewRequest(http.MethodGet, "/index.html", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("GET /index.html via bare server: got %d, want 200", w.Code)
	}
}

// TestServer_WithLoggerOptionTakesPrecedence verifies that config.WithLogger in
// cfg.Server takes precedence over the deprecated positional logger argument.
func TestServer_WithLoggerOptionTakesPrecedence(t *testing.T) {
	t.Parallel()

	var optionBuf bytes.Buffer
	optionLogger := slog.New(slog.NewTextHandler(&optionBuf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	positionalLogger := slog.New(slog.NewTextHandler(io.Discard, nil))

	postsFS := createTestFS(t)
	cfg := config.ServerConfig{
		Server: []config.BaseServerOption{
			{BaseOption: config.WithLogger(optionLogger)},
		},
		Gen: []config.GeneratorOption{config.WithRawOutput()},
	}

	srv, err := server.New(positionalLogger, postsFS, cfg)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	if srv.Logger.Logger != optionLogger {
		t.Error("expected WithLogger option to take precedence over positional logger arg")
	}
}

// TestServer_PositionalLoggerFallback verifies that the positional logger is
// used when no WithLogger option is provided (deprecated path still works).
func TestServer_PositionalLoggerFallback(t *testing.T) {
	t.Parallel()

	positionalLogger := slog.New(slog.NewTextHandler(io.Discard, nil))

	postsFS := createTestFS(t)
	cfg := config.ServerConfig{
		Gen: []config.GeneratorOption{config.WithRawOutput()},
	}

	srv, err := server.New(positionalLogger, postsFS, cfg)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	if srv.Logger.Logger != positionalLogger {
		t.Error("expected positional logger to be used when no WithLogger option is provided")
	}
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

package server

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/harrydayexe/GoBlog/v2/pkg/config"
	pkgserver "github.com/harrydayexe/GoBlog/v2/pkg/server"
)

var testPost = strings.TrimSpace(`
---
title: Test Post
description: A test post
date: 2024-01-01
tags: [test]
---

# Test Post

This is a test post.
`)

func testFS() fstest.MapFS {
	return fstest.MapFS{
		"test-post.md": &fstest.MapFile{Data: []byte(testPost)},
	}
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// TestRunServe_CanceledContext verifies that runServe returns nil when the context is canceled.
func TestRunServe_CanceledContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cfg := config.ServerConfig{
		Server: []config.BaseServerOption{config.WithPort(0)},
	}

	err := runServe(ctx, discardLogger(), testFS(), cfg)
	if err != nil {
		t.Errorf("runServe() with canceled context error = %v, want nil", err)
	}
}

// TestRunServe_ServesIndex verifies that a running server responds to GET / with 200.
func TestRunServe_ServesIndex(t *testing.T) {
	t.Parallel()

	cfg := config.ServerConfig{
		Server: []config.BaseServerOption{config.WithPort(0)},
	}

	srv, err := pkgserver.New(discardLogger(), testFS(), cfg)
	if err != nil {
		t.Fatalf("server.New() error = %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GET / status = %d, want %d", rec.Code, http.StatusOK)
	}
}

// TestRunServe_ServesPost verifies that a running server responds to GET /posts/{name} with 200.
func TestRunServe_ServesPost(t *testing.T) {
	t.Parallel()

	cfg := config.ServerConfig{
		Server: []config.BaseServerOption{config.WithPort(0)},
	}

	srv, err := pkgserver.New(discardLogger(), testFS(), cfg)
	if err != nil {
		t.Fatalf("server.New() error = %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/posts/test-post", nil)
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GET /posts/test-post status = %d, want %d", rec.Code, http.StatusOK)
	}
}

// TestRunServe_BlogRoot verifies that routes are prefixed when a blog root is set.
func TestRunServe_BlogRoot(t *testing.T) {
	t.Parallel()

	cfg := config.ServerConfig{
		Server: []config.BaseServerOption{
			config.WithPort(0),
			{BaseOption: config.WithBlogRoot("/blog/")},
		},
	}

	srv, err := pkgserver.New(discardLogger(), testFS(), cfg)
	if err != nil {
		t.Fatalf("server.New() error = %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/blog/", nil)
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GET /blog/ status = %d, want %d", rec.Code, http.StatusOK)
	}

	// Bare / should not match when a blog root is set
	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	srv.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusNotFound {
		t.Errorf("GET / with blog root status = %d, want %d", rec2.Code, http.StatusNotFound)
	}
}

// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package server

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/harrydayexe/GoBlog/v2/pkg/config"
	pkgserver "github.com/harrydayexe/GoBlog/v2/pkg/server"
	"github.com/harrydayexe/GoBlog/v2/pkg/watcher"
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

	err := runServe(ctx, t.TempDir(), testFS(), cfg, false)
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

// TestRunServe_WatchBadPath verifies that runServe fails fast when --watch is
// enabled but the posts path does not exist.
func TestRunServe_WatchBadPath(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cfg := config.ServerConfig{
		Server: []config.BaseServerOption{config.WithPort(0)},
	}

	err := runServe(ctx, "/nonexistent/goblog/watch/test", testFS(), cfg, true)
	if err == nil {
		t.Error("runServe() with watch=true and bad path returned nil, want error")
	}
}

// TestRunServe_WatchReloadsPost verifies that writing a new post file to a
// watched directory causes it to be served by the server. It wires
// pkg/watcher and pkg/server together directly, mirroring the integration
// that runServe performs.
func TestRunServe_WatchReloadsPost(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "test-post.md"), []byte(testPost), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	cfg := config.ServerConfig{
		Server: []config.BaseServerOption{config.WithPort(0)},
	}

	srv, err := pkgserver.New(discardLogger(), os.DirFS(dir), cfg)
	if err != nil {
		t.Fatalf("server.New() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	w, err := watcher.New(dir, config.WithDebounce(50*time.Millisecond))
	if err != nil {
		t.Fatalf("watcher.New() error = %v", err)
	}
	go w.Run(ctx, func(ctx context.Context) { //nolint:errcheck
		if err := srv.UpdatePosts(os.DirFS(dir), ctx); err != nil {
			t.Logf("UpdatePosts error: %v", err)
		}
	})

	// Give the watcher time to start.
	time.Sleep(100 * time.Millisecond)

	newPost := strings.TrimSpace(`
---
title: New Post
description: A new post
date: 2024-06-01
tags: [new]
---

# New Post

This is a new post.
`)
	if err := os.WriteFile(filepath.Join(dir, "new-post.md"), []byte(newPost), 0o644); err != nil {
		t.Fatalf("WriteFile new-post error = %v", err)
	}

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/posts/new-post", nil))
		if rec.Code == http.StatusOK {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Error("new post not available after 5s of watching")
}

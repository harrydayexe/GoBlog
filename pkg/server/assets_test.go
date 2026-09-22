// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package server_test

import (
	"bytes"
	"image"
	"image/png"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/harrydayexe/GoBlog/v2/pkg/config"
	"github.com/harrydayexe/GoBlog/v2/pkg/generator"
	"github.com/harrydayexe/GoBlog/v2/pkg/server"
)

// pngBytes is the 8-byte PNG signature, enough for content sniffing.
var pngBytes = []byte("\x89PNG\r\n\x1a\n")

func testAssetsFS() fstest.MapFS {
	return fstest.MapFS{
		"pipeline.png":          {Data: pngBytes},
		"screenshots/a.png":     {Data: pngBytes},
		"diagram.html.png":      {Data: pngBytes},
		"screenshots/notes.txt": {Data: []byte("hello")},
	}
}

func newAssetsServer(t *testing.T, opts ...config.ServerOption) *server.Server {
	t.Helper()
	opts = append(opts, config.WithRawOutput().AsServerOption())
	srv, err := server.New(createTestFS(t), opts...)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}
	return srv
}

func get(h http.Handler, target string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, target, nil))
	return w
}

func TestAssets_Served(t *testing.T) {
	t.Parallel()

	srv := newAssetsServer(t, config.WithAssetsDir(testAssetsFS()).AsServerOption())

	tests := []struct {
		name       string
		target     string
		wantStatus int
	}{
		{name: "top-level file", target: "/images/pipeline.png", wantStatus: http.StatusOK},
		{name: "subdirectory file", target: "/images/screenshots/a.png", wantStatus: http.StatusOK},
		{name: "missing file", target: "/images/missing.png", wantStatus: http.StatusNotFound},
		{name: "directory listing suppressed", target: "/images/screenshots/", wantStatus: http.StatusNotFound},
		{name: "root listing suppressed", target: "/images/", wantStatus: http.StatusNotFound},
		{name: "html extension stripping does not interfere", target: "/images/diagram.html.png", wantStatus: http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			w := get(srv, tt.target)
			if w.Code != tt.wantStatus {
				t.Fatalf("GET %s: status %d, want %d", tt.target, w.Code, tt.wantStatus)
			}
		})
	}
}

func TestAssets_Headers(t *testing.T) {
	t.Parallel()

	srv := newAssetsServer(t, config.WithAssetsDir(testAssetsFS()).AsServerOption())

	w := get(srv, "/images/pipeline.png")
	if w.Code != http.StatusOK {
		t.Fatalf("status %d, want 200", w.Code)
	}
	if got := w.Header().Get("Content-Type"); got != "image/png" {
		t.Errorf("Content-Type: got %q, want %q", got, "image/png")
	}
	if got := w.Header().Get("Cache-Control"); got != "public, max-age=3600" {
		t.Errorf("Cache-Control: got %q, want %q", got, "public, max-age=3600")
	}
	if body, _ := io.ReadAll(w.Body); string(body) != string(pngBytes) {
		t.Errorf("unexpected body %q", body)
	}
}

func TestAssets_UnderBlogRoot(t *testing.T) {
	t.Parallel()

	srv := newAssetsServer(t,
		config.WithAssetsDir(testAssetsFS()).AsServerOption(),
		config.WithBlogRoot("/blog/").AsServerOption(),
	)

	if w := get(srv, "/blog/images/pipeline.png"); w.Code != http.StatusOK {
		t.Errorf("GET /blog/images/pipeline.png: status %d, want 200", w.Code)
	}
	if w := get(srv, "/images/pipeline.png"); w.Code != http.StatusNotFound {
		t.Errorf("GET /images/pipeline.png: status %d, want 404", w.Code)
	}
}

func TestAssets_Traversal(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	assets := filepath.Join(dir, "images")
	if err := os.Mkdir(assets, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "secret.txt"), []byte("secret"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(dir, "secret.txt"), filepath.Join(assets, "link.png")); err != nil {
		t.Fatal(err)
	}
	root, err := os.OpenRoot(assets)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { root.Close() })

	h := server.Handler(&generator.GeneratedBlog{}, config.WithAssetsDir(root.FS()))

	for _, target := range []string{
		"/images/../secret.txt",
		"/images/%2e%2e/secret.txt",
		"/images/..%2fsecret.txt",
		"/images/link.png",
	} {
		w := get(h, target)
		if w.Code == http.StatusOK {
			t.Errorf("GET %s: status 200, want non-200; body %q", target, w.Body.String())
		}
	}
}

func TestAssets_NoRouteWhenDisabled(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		opts []config.BaseOption
	}{
		{name: "option not supplied"},
		{name: "nil filesystem", opts: []config.BaseOption{config.WithAssetsDir(nil)}},
		{name: "missing directory", opts: []config.BaseOption{config.WithAssetsDir(os.DirFS(filepath.Join(t.TempDir(), "nope")))}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := server.Handler(&generator.GeneratedBlog{}, tt.opts...)
			if w := get(h, "/images/pipeline.png"); w.Code != http.StatusNotFound {
				t.Errorf("status %d, want 404", w.Code)
			}
		})
	}
}

// TestAssets_DimensionsReachParser verifies that the assets filesystem given
// to the server is also used by the parser to measure images, so served posts
// carry width and height attributes.
func TestAssets_DimensionsReachParser(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 800, 600))); err != nil {
		t.Fatalf("failed to encode png: %v", err)
	}

	posts := fstest.MapFS{
		"image-post.md": {Data: []byte(strings.TrimSpace(`
---
title: Image Post
description: A post with an image
date: 2024-01-01
---

![A diagram](pipeline.png)
		`))},
	}
	srv, err := server.New(posts,
		config.WithAssetsDir(fstest.MapFS{"pipeline.png": {Data: buf.Bytes()}}).AsServerOption(),
		config.WithLogger(slog.New(slog.DiscardHandler)).AsServerOption(),
	)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	w := get(srv, "/posts/image-post")
	if w.Code != http.StatusOK {
		t.Fatalf("GET /posts/image-post status = %d, want %d", w.Code, http.StatusOK)
	}

	want := `<img src="/images/pipeline.png" alt="A diagram" width="800" height="600" loading="lazy" decoding="async" />`
	if body := w.Body.String(); !strings.Contains(body, want) {
		t.Errorf("expected body to contain %q, got:\n%s", want, body)
	}
}

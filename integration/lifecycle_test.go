// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package integration_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/harrydayexe/GoBlog/v2/pkg/config"
	"github.com/harrydayexe/GoBlog/v2/pkg/generator"
	"github.com/harrydayexe/GoBlog/v2/pkg/outputter"
	"github.com/harrydayexe/GoBlog/v2/pkg/server"
	"github.com/harrydayexe/GoBlog/v2/pkg/templates"
)

// TestRun_BindError verifies that Server.Run surfaces a bind error when the
// configured port is already occupied, rather than silently swallowing the
// listenErr (pkg/server/server.go, the ListenAndServe goroutine).
func TestRun_BindError(t *testing.T) {
	// Occupy a port to force the bind conflict.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen: %v", err)
	}
	defer listener.Close()
	port := listener.Addr().(*net.TCPAddr).Port

	dir := t.TempDir()
	writePost(t, dir, "post.md", minimalPost("Hello World"))

	srv, err := server.New(os.DirFS(dir),
		config.WithPort(port),
		config.WithHost("127.0.0.1"),
	)
	if err != nil {
		t.Fatalf("server.New: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Run(ctx); err == nil {
		t.Fatal("expected a bind error from Run, got nil")
	}
}

// TestRun_GracefulShutdown verifies that cancelling the context causes Run to
// return nil and complete well within the 10 s configured shutdown window.
func TestRun_GracefulShutdown(t *testing.T) {
	dir := t.TempDir()
	writePost(t, dir, "post.md", minimalPost("Hello World"))

	// Grab a free port by binding, recording it, then releasing it.
	// NOTE: small TOCTOU window — the port is released here and re-bound by
	// the server below. On a heavily loaded host another process could claim
	// it in between, causing a rare spurious bind failure. Accepted: the
	// server API binds by port number (ListenAndServe) and does not accept a
	// pre-opened listener, so handing the fd over directly is not currently
	// possible.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	srv, err := server.New(os.DirFS(dir),
		config.WithPort(port),
		config.WithHost("127.0.0.1"),
	)
	if err != nil {
		t.Fatalf("server.New: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // guarantees the Run goroutine is unwound if eventually fails

	done := make(chan error, 1)
	go func() { done <- srv.Run(ctx) }()

	// Wait until the server is ready to serve requests.
	addr := fmt.Sprintf("http://127.0.0.1:%d/", port)
	eventually(t, 5*time.Second, 50*time.Millisecond, func() bool {
		//nolint:gosec // test-controlled URL
		resp, err := http.Get(addr)
		if err != nil {
			return false
		}
		resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	})

	// Cancel the context and measure how long shutdown takes.
	start := time.Now()
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run returned unexpected error: %v", err)
		}
		if elapsed := time.Since(start); elapsed > 5*time.Second {
			t.Errorf("shutdown took %s; want < 5 s", elapsed)
		}
	case <-time.After(15 * time.Second):
		t.Fatal("server did not shut down within 15 s")
	}
}

// TestGenerate_SitemapAndRobots exercises the generate path end-to-end: the
// generator, the directory writer, and the base-URL gating that produces
// sitemap.xml and robots.txt.
func TestGenerate_SitemapAndRobots(t *testing.T) {
	postsDir := t.TempDir()
	outputDir := t.TempDir()
	writePost(t, postsDir, "post.md", minimalPost("Hello World"))

	renderer, err := generator.NewTemplateRenderer(templates.Default)
	if err != nil {
		t.Fatalf("NewTemplateRenderer: %v", err)
	}

	opts := []config.GeneratorOption{
		config.WithBaseURL("https://example.com"),
		config.WithHTMLPaths(),
	}
	gen := generator.New(os.DirFS(postsDir), renderer, opts...)
	blog, err := gen.Generate(context.Background())
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	writer := outputter.NewDirectoryWriter(outputDir, opts...)
	if err := writer.HandleGeneratedBlog(context.Background(), blog); err != nil {
		t.Fatalf("HandleGeneratedBlog: %v", err)
	}

	sitemap, err := os.ReadFile(filepath.Join(outputDir, "sitemap.xml"))
	if err != nil {
		t.Fatalf("read sitemap.xml: %v", err)
	}
	for _, want := range []string{
		"http://www.sitemaps.org/schemas/sitemap/0.9",
		"<loc>https://example.com/index.html</loc>",
		"<loc>https://example.com/posts/hello-world.html</loc>",
	} {
		if !strings.Contains(string(sitemap), want) {
			t.Errorf("sitemap.xml does not contain %q:\n%s", want, sitemap)
		}
	}

	robots, err := os.ReadFile(filepath.Join(outputDir, "robots.txt"))
	if err != nil {
		t.Fatalf("read robots.txt: %v", err)
	}
	if want := "User-agent: *\nAllow: /\n\nSitemap: https://example.com/sitemap.xml\n"; string(robots) != want {
		t.Errorf("robots.txt =\n%q\nwant\n%q", robots, want)
	}
}

// TestServe_SitemapAndRobots verifies that both artefacts are reachable over
// HTTP from a running server, with robots.txt at the origin root even though a
// blog root is configured.
func TestServe_SitemapAndRobots(t *testing.T) {
	dir := t.TempDir()
	writePost(t, dir, "post.md", minimalPost("Hello World"))

	srv, err := server.New(os.DirFS(dir),
		config.WithPort(0),
		config.WithBlogRoot("/blog/").AsServerOption(),
		config.WithBaseURL("https://example.com").AsServerOption(),
	)
	if err != nil {
		t.Fatalf("server.New: %v", err)
	}

	tests := []struct {
		path        string
		contentType string
		contains    string
	}{
		{"/blog/sitemap.xml", "application/xml; charset=utf-8", "<loc>https://example.com/blog/posts/hello-world</loc>"},
		// robots.txt is served at the origin root, not under /blog/.
		{"/robots.txt", "text/plain; charset=utf-8", "Sitemap: https://example.com/blog/sitemap.xml"},
	}

	for _, tt := range tests {
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tt.path, nil))

		if rec.Code != http.StatusOK {
			t.Errorf("GET %s status = %d, want %d", tt.path, rec.Code, http.StatusOK)
			continue
		}
		if ct := rec.Header().Get("Content-Type"); ct != tt.contentType {
			t.Errorf("GET %s Content-Type = %q, want %q", tt.path, ct, tt.contentType)
		}
		if !strings.Contains(rec.Body.String(), tt.contains) {
			t.Errorf("GET %s body does not contain %q:\n%s", tt.path, tt.contains, rec.Body.String())
		}
	}

	// The sitemap must not also be reachable outside the blog root.
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil))
	if rec.Code != http.StatusNotFound {
		t.Errorf("GET /sitemap.xml status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

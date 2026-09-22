// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/harrydayexe/GoBlog/v2/pkg/config"
	"github.com/harrydayexe/GoBlog/v2/pkg/generator"
	"github.com/harrydayexe/GoBlog/v2/pkg/server"
)

var (
	fakeSitemap = []byte(`<?xml version="1.0" encoding="UTF-8"?>` +
		`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"></urlset>`)
	fakeRobots = []byte("User-agent: *\nAllow: /\n\nSitemap: https://example.com/sitemap.xml\n")
)

// testBlogWithSitemap returns a minimal GeneratedBlog carrying a sitemap and a
// robots.txt.
func testBlogWithSitemap() *generator.GeneratedBlog {
	blog := generator.NewEmptyGeneratedBlog()
	blog.Sitemap = fakeSitemap
	blog.RobotsTxt = fakeRobots
	return blog
}

// TestHandler_Sitemap verifies that GET /sitemap.xml returns the sitemap with
// the correct status, Content-Type, and body.
func TestHandler_Sitemap(t *testing.T) {
	t.Parallel()

	h := server.Handler(testBlogWithSitemap())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GET /sitemap.xml status = %d, want %d", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/xml; charset=utf-8" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/xml; charset=utf-8")
	}
	if rec.Body.String() != string(fakeSitemap) {
		t.Errorf("body = %q, want %q", rec.Body.String(), fakeSitemap)
	}
}

// TestHandler_RobotsTxt verifies that GET /robots.txt returns the robots file
// with the correct status, Content-Type, and body.
func TestHandler_RobotsTxt(t *testing.T) {
	t.Parallel()

	h := server.Handler(testBlogWithSitemap())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/robots.txt", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GET /robots.txt status = %d, want %d", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/plain; charset=utf-8" {
		t.Errorf("Content-Type = %q, want %q", ct, "text/plain; charset=utf-8")
	}
	if rec.Body.String() != string(fakeRobots) {
		t.Errorf("body = %q, want %q", rec.Body.String(), fakeRobots)
	}
}

// TestHandler_SitemapAndRobots_Empty verifies both routes 404 when the
// generator produced no content for them.
func TestHandler_SitemapAndRobots_Empty(t *testing.T) {
	t.Parallel()

	h := server.Handler(generator.NewEmptyGeneratedBlog())

	for _, path := range []string{"/sitemap.xml", "/robots.txt"} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("GET %s (empty) status = %d, want %d", path, rec.Code, http.StatusNotFound)
		}
	}
}

// TestHandler_SitemapAndRobotsWithBlogRoot verifies that the sitemap moves
// under the blog root while robots.txt stays at the origin root, where
// crawlers look for it.
func TestHandler_SitemapAndRobotsWithBlogRoot(t *testing.T) {
	t.Parallel()

	h := server.Handler(testBlogWithSitemap(), config.WithBlogRoot("/blog/"))

	tests := []struct {
		path string
		want int
	}{
		{"/blog/sitemap.xml", http.StatusOK},
		// robots.txt is served at the origin root regardless of the blog root.
		{"/robots.txt", http.StatusOK},
		// The sitemap is not reachable without the blog root prefix.
		{"/sitemap.xml", http.StatusNotFound},
	}

	for _, tc := range tests {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, tc.path, nil)
		h.ServeHTTP(rec, req)

		if rec.Code != tc.want {
			t.Errorf("GET %s status = %d, want %d", tc.path, rec.Code, tc.want)
		}
	}
}

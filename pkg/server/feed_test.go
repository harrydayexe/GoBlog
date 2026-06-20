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

var fakeRSS = []byte(`<?xml version="1.0" encoding="UTF-8"?><rss version="2.0"></rss>`)
var fakeAtom = []byte(`<?xml version="1.0" encoding="UTF-8"?><feed xmlns="http://www.w3.org/2005/Atom"></feed>`)

// testBlogWithFeeds returns a minimal GeneratedBlog populated with feed content
// for the given tag name.
func testBlogWithFeeds(tag string) *generator.GeneratedBlog {
	blog := generator.NewEmptyGeneratedBlog()
	blog.RSSFeed = fakeRSS
	blog.AtomFeed = fakeAtom
	if tag != "" {
		blog.Tags[tag] = []byte("<html>tag page</html>")
		blog.TagRSSFeeds[tag] = fakeRSS
		blog.TagAtomFeeds[tag] = fakeAtom
	}
	return blog
}

// TestHandler_SiteRSSFeed verifies that GET /rss.xml returns 200 with the
// correct Content-Type when the generator produced a feed.
func TestHandler_SiteRSSFeed(t *testing.T) {
	t.Parallel()

	blog := testBlogWithFeeds("")
	h := server.Handler(blog, nil)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/rss.xml", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GET /rss.xml status = %d, want %d", rec.Code, http.StatusOK)
	}
	ct := rec.Header().Get("Content-Type")
	if ct != "application/rss+xml; charset=utf-8" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/rss+xml; charset=utf-8")
	}
}

// TestHandler_SiteAtomFeed verifies that GET /atom.xml returns 200 with the
// correct Content-Type when the generator produced a feed.
func TestHandler_SiteAtomFeed(t *testing.T) {
	t.Parallel()

	blog := testBlogWithFeeds("")
	h := server.Handler(blog, nil)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/atom.xml", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GET /atom.xml status = %d, want %d", rec.Code, http.StatusOK)
	}
	ct := rec.Header().Get("Content-Type")
	if ct != "application/atom+xml; charset=utf-8" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/atom+xml; charset=utf-8")
	}
}

// TestHandler_SiteRSSFeed_Empty returns 404 when the generator did not produce
// a feed (e.g. base-url not set or feeds disabled).
func TestHandler_SiteRSSFeed_Empty(t *testing.T) {
	t.Parallel()

	blog := generator.NewEmptyGeneratedBlog() // no feed bytes
	h := server.Handler(blog, nil)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/rss.xml", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("GET /rss.xml (empty) status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

// TestHandler_SiteAtomFeed_Empty returns 404 when the generator did not produce
// a feed.
func TestHandler_SiteAtomFeed_Empty(t *testing.T) {
	t.Parallel()

	blog := generator.NewEmptyGeneratedBlog() // no feed bytes
	h := server.Handler(blog, nil)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/atom.xml", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("GET /atom.xml (empty) status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

// TestHandler_TagRSSFeed verifies that GET /tags/{tag}.rss.xml returns 200
// with the correct Content-Type when the generator produced a per-tag feed.
func TestHandler_TagRSSFeed(t *testing.T) {
	t.Parallel()

	blog := testBlogWithFeeds("golang")
	h := server.Handler(blog, nil)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/tags/golang.rss.xml", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GET /tags/golang.rss.xml status = %d, want %d", rec.Code, http.StatusOK)
	}
	ct := rec.Header().Get("Content-Type")
	if ct != "application/rss+xml; charset=utf-8" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/rss+xml; charset=utf-8")
	}
}

// TestHandler_TagAtomFeed verifies that GET /tags/{tag}.atom.xml returns 200
// with the correct Content-Type.
func TestHandler_TagAtomFeed(t *testing.T) {
	t.Parallel()

	blog := testBlogWithFeeds("golang")
	h := server.Handler(blog, nil)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/tags/golang.atom.xml", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GET /tags/golang.atom.xml status = %d, want %d", rec.Code, http.StatusOK)
	}
	ct := rec.Header().Get("Content-Type")
	if ct != "application/atom+xml; charset=utf-8" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/atom+xml; charset=utf-8")
	}
}

// TestHandler_TagRSSFeed_UnknownTag returns 404 for an unknown tag.
func TestHandler_TagRSSFeed_UnknownTag(t *testing.T) {
	t.Parallel()

	blog := testBlogWithFeeds("golang")
	h := server.Handler(blog, nil)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/tags/rust.rss.xml", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("GET /tags/rust.rss.xml (unknown) status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

// TestHandler_TagAtomFeed_UnknownTag returns 404 for an unknown tag.
func TestHandler_TagAtomFeed_UnknownTag(t *testing.T) {
	t.Parallel()

	blog := testBlogWithFeeds("golang")
	h := server.Handler(blog, nil)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/tags/rust.atom.xml", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("GET /tags/rust.atom.xml (unknown) status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

// TestHandler_FeedsWithBlogRoot verifies that feed routes are served under the
// blog root when one is configured.
func TestHandler_FeedsWithBlogRoot(t *testing.T) {
	t.Parallel()

	blog := testBlogWithFeeds("golang")
	h := server.Handler(blog, nil, config.WithBlogRoot("/blog/"))

	tests := []struct {
		path string
		want int
	}{
		{"/blog/rss.xml", http.StatusOK},
		{"/blog/atom.xml", http.StatusOK},
		{"/blog/tags/golang.rss.xml", http.StatusOK},
		{"/blog/tags/golang.atom.xml", http.StatusOK},
		// Without the prefix should not match.
		{"/rss.xml", http.StatusNotFound},
		{"/atom.xml", http.StatusNotFound},
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

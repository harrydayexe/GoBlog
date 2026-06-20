// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package generator

import (
	"context"
	"encoding/xml"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/harrydayexe/GoBlog/v2/pkg/config"
	"github.com/harrydayexe/GoBlog/v2/pkg/models"
	"github.com/harrydayexe/GoBlog/v2/pkg/templates"
)

// ---- helpers ----------------------------------------------------------------

func makePosts(n int) models.PostList {
	posts := make(models.PostList, n)
	for i := range n {
		posts[i] = &models.Post{
			Title:       strings.Repeat("Post ", i+1),
			Date:        time.Date(2024, time.January, n-i, 0, 0, 0, 0, time.UTC),
			Description: "desc",
			Slug:        strings.ToLower(strings.ReplaceAll(strings.TrimSpace(strings.Repeat("post ", i+1)), " ", "-")),
			Content:     []byte("<p>content " + strings.Repeat("post ", i+1) + "</p>"),
		}
	}
	return posts
}

// newFeedsGen constructs a minimal Generator for BuildFeeds tests,
// using the provided base URL, site title, and post limit.
func newFeedsGen(baseURL, siteTitle string, limit int) *Generator {
	opts := []config.GeneratorOption{
		config.WithBaseURL(baseURL),
		config.WithSiteTitle(siteTitle),
	}
	if limit > 0 {
		opts = append(opts, config.WithFeedPostLimit(limit))
	}
	return New(fstest.MapFS{}, nil, opts...)
}

// ---- BuildFeeds tests -------------------------------------------------------

func TestBuildFeeds_EmptyPosts(t *testing.T) {
	t.Parallel()

	gen := newFeedsGen("https://example.com", "Blog", 10)
	rss, atom, err := gen.buildFeeds(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rss != nil || atom != nil {
		t.Errorf("expected nil output for empty posts, got rss=%d bytes atom=%d bytes", len(rss), len(atom))
	}
}

func TestBuildFeeds_ProducesValidXML(t *testing.T) {
	t.Parallel()

	gen := newFeedsGen("https://example.com", "Test Blog", 10)
	posts := makePosts(3)
	rss, atom, err := gen.buildFeeds(posts)
	if err != nil {
		t.Fatalf("BuildFeeds error: %v", err)
	}

	if err := xml.Unmarshal(rss, new(interface{})); err != nil {
		t.Errorf("RSS output is not valid XML: %v", err)
	}
	if err := xml.Unmarshal(atom, new(interface{})); err != nil {
		t.Errorf("Atom output is not valid XML: %v", err)
	}
}

func TestBuildFeeds_AbsoluteURLsInItems(t *testing.T) {
	t.Parallel()

	gen := newFeedsGen("https://myblog.com", "My Blog", 10)
	posts := makePosts(2)
	rss, atom, err := gen.buildFeeds(posts)
	if err != nil {
		t.Fatalf("BuildFeeds error: %v", err)
	}

	for _, feed := range []struct {
		name    string
		content []byte
	}{
		{"RSS", rss},
		{"Atom", atom},
	} {
		for _, post := range posts {
			wantURL := "https://myblog.com/posts/" + post.Slug
			if !strings.Contains(string(feed.content), wantURL) {
				t.Errorf("%s feed missing absolute URL %q", feed.name, wantURL)
			}
		}
	}
}

func TestBuildFeeds_LimitRespected(t *testing.T) {
	t.Parallel()

	gen := newFeedsGen("https://example.com", "Blog", 5)
	posts := makePosts(15)
	rss, _, err := gen.buildFeeds(posts)
	if err != nil {
		t.Fatalf("BuildFeeds error: %v", err)
	}

	// Only the first 5 posts (newest) should appear. Posts 6-15 must not.
	for i, post := range posts {
		url := "https://example.com/posts/" + post.Slug
		if i < 5 {
			if !strings.Contains(string(rss), url) {
				t.Errorf("post %d should be in feed but is missing", i+1)
			}
		} else {
			if strings.Contains(string(rss), url) {
				t.Errorf("post %d should be excluded by limit but is present", i+1)
			}
		}
	}
}

func TestBuildFeeds_FullContentIncluded(t *testing.T) {
	t.Parallel()

	gen := newFeedsGen("https://example.com", "Blog", 10)
	posts := models.PostList{
		{
			Title:   "Hello",
			Slug:    "hello",
			Date:    time.Now(),
			Content: []byte("<p>full HTML content here</p>"),
		},
	}

	rss, atom, err := gen.buildFeeds(posts)
	if err != nil {
		t.Fatalf("BuildFeeds error: %v", err)
	}

	for _, feed := range []struct {
		name    string
		content []byte
	}{
		{"RSS", rss},
		{"Atom", atom},
	} {
		if !strings.Contains(string(feed.content), "full HTML content here") {
			t.Errorf("%s feed does not contain full post content", feed.name)
		}
	}
}

func TestBuildFeeds_PostURLIsItemID(t *testing.T) {
	t.Parallel()

	gen := newFeedsGen("https://example.com", "Blog", 10)
	posts := models.PostList{
		{Title: "A Post", Slug: "a-post", Date: time.Now(), Content: []byte("x")},
	}
	rss, atom, err := gen.buildFeeds(posts)
	if err != nil {
		t.Fatalf("BuildFeeds error: %v", err)
	}

	wantURL := "https://example.com/posts/a-post"
	// RSS uses <guid>; Atom uses <id>. Both must contain the full post URL.
	if !strings.Contains(string(rss), "<guid>"+wantURL+"</guid>") {
		t.Errorf("RSS feed guid does not equal post URL; feed:\n%s", rss)
	}
	if !strings.Contains(string(atom), "<id>"+wantURL+"</id>") {
		t.Errorf("Atom feed id does not equal post URL; feed:\n%s", atom)
	}
}

// ---- AbsURL tests -----------------------------------------------------------

func TestAbsURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		base string
		path string
		want string
	}{
		{"https://example.com", "/posts/hello", "https://example.com/posts/hello"},
		{"https://example.com/", "/posts/hello", "https://example.com/posts/hello"},
		{"https://example.com/blog", "/posts/hello", "https://example.com/blog/posts/hello"},
		{"https://example.com/blog/", "/posts/hello", "https://example.com/blog/posts/hello"},
	}

	for _, tt := range tests {
		t.Run(tt.base+tt.path, func(t *testing.T) {
			t.Parallel()
			got := AbsURL(tt.base, tt.path)
			if got != tt.want {
				t.Errorf("AbsURL(%q, %q) = %q, want %q", tt.base, tt.path, got, tt.want)
			}
		})
	}
}

// ---- generator integration tests -------------------------------------------

func newTestRenderer(t *testing.T) *TemplateRenderer {
	t.Helper()
	r, err := NewTemplateRenderer(templates.Default)
	if err != nil {
		t.Fatalf("NewTemplateRenderer: %v", err)
	}
	return r
}

func newTestPostsFS(t *testing.T, posts map[string]string) fstest.MapFS {
	t.Helper()
	fs := fstest.MapFS{}
	for name, content := range posts {
		fs[name] = &fstest.MapFile{Data: []byte(content)}
	}
	return fs
}

const simplePost = `---
title: "Hello World"
date: 2024-06-01
description: "A simple post"
author: "Alice"
---
# Hello

This is content.
`

const olderPost = `---
title: "Old Post"
date: 2024-01-01
description: "An older post"
---
Old content.
`

func TestGenerator_FeedsSkippedWhenNoBaseURL(t *testing.T) {
	t.Parallel()

	postsFS := newTestPostsFS(t, map[string]string{"hello.md": simplePost})
	gen := New(postsFS, newTestRenderer(t))

	blog, err := gen.Generate(context.Background())
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if blog.RSSFeed != nil || blog.AtomFeed != nil {
		t.Errorf("expected no feeds when base URL is unset, got rss=%d atom=%d", len(blog.RSSFeed), len(blog.AtomFeed))
	}
}

func TestGenerator_FeedsGeneratedWithBaseURL(t *testing.T) {
	t.Parallel()

	postsFS := newTestPostsFS(t, map[string]string{"hello.md": simplePost})
	gen := New(postsFS, newTestRenderer(t), config.WithBaseURL("https://example.com"))

	blog, err := gen.Generate(context.Background())
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if len(blog.RSSFeed) == 0 {
		t.Error("expected RSS feed, got empty bytes")
	}
	if len(blog.AtomFeed) == 0 {
		t.Error("expected Atom feed, got empty bytes")
	}

	// Both should be valid XML.
	if err := xml.Unmarshal(blog.RSSFeed, new(interface{})); err != nil {
		t.Errorf("RSS output is not valid XML: %v", err)
	}
	if err := xml.Unmarshal(blog.AtomFeed, new(interface{})); err != nil {
		t.Errorf("Atom output is not valid XML: %v", err)
	}

	// Post URLs must be absolute.
	if !strings.Contains(string(blog.RSSFeed), "https://example.com/posts/hello-world") {
		t.Errorf("RSS feed missing absolute post URL; feed:\n%s", blog.RSSFeed)
	}
}

func TestGenerator_FeedsDisabled(t *testing.T) {
	t.Parallel()

	postsFS := newTestPostsFS(t, map[string]string{"hello.md": simplePost})
	gen := New(postsFS, newTestRenderer(t),
		config.WithBaseURL("https://example.com"),
		config.WithDisableFeeds(),
	)

	blog, err := gen.Generate(context.Background())
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if blog.RSSFeed != nil || blog.AtomFeed != nil {
		t.Errorf("expected no feeds when WithDisableFeeds is set")
	}
}

func TestGenerator_FeedPostLimit(t *testing.T) {
	t.Parallel()

	// Create 5 posts but limit the feed to 2.
	const postTemplate = `---
title: "Post %d"
date: 2024-0%d-01
description: "Post number %d"
---
Content %d.
`
	postsMap := map[string]string{}
	for i := 1; i <= 5; i++ {
		name := "post" + strings.Repeat("0", 1) + string(rune('0'+i)) + ".md"
		postsMap[name] = strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(
			postTemplate, "%d", "%v"), "%v", strings.Repeat("%v", 1)), "%v", "x"), "x", "1")
	}

	// Simpler: use fstest directly with known slugs.
	fs := fstest.MapFS{
		"post1.md": {Data: []byte("---\ntitle: \"Post 1\"\ndate: 2024-05-01\ndescription: \"p1\"\n---\ncontent 1")},
		"post2.md": {Data: []byte("---\ntitle: \"Post 2\"\ndate: 2024-04-01\ndescription: \"p2\"\n---\ncontent 2")},
		"post3.md": {Data: []byte("---\ntitle: \"Post 3\"\ndate: 2024-03-01\ndescription: \"p3\"\n---\ncontent 3")},
		"post4.md": {Data: []byte("---\ntitle: \"Post 4\"\ndate: 2024-02-01\ndescription: \"p4\"\n---\ncontent 4")},
		"post5.md": {Data: []byte("---\ntitle: \"Post 5\"\ndate: 2024-01-01\ndescription: \"p5\"\n---\ncontent 5")},
	}

	gen := New(fs, newTestRenderer(t),
		config.WithBaseURL("https://example.com"),
		config.WithFeedPostLimit(2),
	)

	blog, err := gen.Generate(context.Background())
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	// Only the 2 newest posts (post1 = May, post2 = April) should appear.
	rss := string(blog.RSSFeed)
	if !strings.Contains(rss, "https://example.com/posts/post-1") {
		t.Error("post 1 (newest) should be in feed")
	}
	if !strings.Contains(rss, "https://example.com/posts/post-2") {
		t.Error("post 2 should be in feed")
	}
	if strings.Contains(rss, "https://example.com/posts/post-3") {
		t.Error("post 3 should be excluded by limit")
	}
}

func TestGenerator_PerTagFeeds(t *testing.T) {
	t.Parallel()

	postsFS := fstest.MapFS{
		"go-post.md": {Data: []byte("---\ntitle: \"Go Post\"\ndate: 2024-06-01\ndescription: \"d\"\ntags:\n  - go\n---\nContent.")},
		"js-post.md": {Data: []byte("---\ntitle: \"JS Post\"\ndate: 2024-05-01\ndescription: \"d\"\ntags:\n  - javascript\n---\nContent.")},
	}

	gen := New(postsFS, newTestRenderer(t), config.WithBaseURL("https://example.com"))

	blog, err := gen.Generate(context.Background())
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if len(blog.TagRSSFeeds) == 0 {
		t.Error("expected per-tag RSS feeds, got none")
	}

	goFeed, ok := blog.TagRSSFeeds["go"]
	if !ok {
		t.Fatal("expected 'go' tag feed")
	}
	if !strings.Contains(string(goFeed), "https://example.com/posts/go-post") {
		t.Errorf("'go' tag feed missing go post; feed:\n%s", goFeed)
	}
	if strings.Contains(string(goFeed), "https://example.com/posts/js-post") {
		t.Errorf("'go' tag feed should not contain js post; feed:\n%s", goFeed)
	}
}

func TestGenerator_FeedsEnabledInBaseData(t *testing.T) {
	t.Parallel()

	// Without base URL: FeedsEnabled must be false in templates.
	postsFS := newTestPostsFS(t, map[string]string{"hello.md": simplePost})
	genNoURL := New(postsFS, newTestRenderer(t))

	blog, err := genNoURL.Generate(context.Background())
	if err != nil {
		t.Fatalf("Generate (no URL): %v", err)
	}
	// The rendered HTML must not contain the RSS link rel.
	if strings.Contains(string(blog.Index), "application/rss+xml") {
		t.Error("index page should not have RSS link when no base URL is set")
	}

	// With base URL: FeedsEnabled must be true in templates.
	genWithURL := New(postsFS, newTestRenderer(t), config.WithBaseURL("https://example.com"))

	blog2, err := genWithURL.Generate(context.Background())
	if err != nil {
		t.Fatalf("Generate (with URL): %v", err)
	}
	if !strings.Contains(string(blog2.Index), "application/rss+xml") {
		t.Error("index page should have RSS discovery link when base URL is set")
	}
}

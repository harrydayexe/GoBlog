// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package generator

import (
	"context"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/harrydayexe/GoBlog/v2/pkg/config"
)

// wantTags asserts that every want appears in the rendered page and no
// notWant does.
func wantTags(t *testing.T, page, rendered string, want, notWant []string) {
	t.Helper()
	for _, w := range want {
		if !strings.Contains(rendered, w) {
			t.Errorf("%s page: missing %s", page, w)
		}
	}
	for _, n := range notWant {
		if strings.Contains(rendered, n) {
			t.Errorf("%s page: unexpectedly contains %s", page, n)
		}
	}
}

// taggedPost is a post with every field the SEO meta tags draw on: an author
// and a tag list as well as the required title, date, and description.
const taggedPost = `---
title: "Hello World"
date: 2024-06-01T09:30:00Z
description: "A simple post"
author: "Alice"
tags: ["go", "testing"]
---
# Hello

This is content.
`

// seoTemplateFS renders only the SEO-relevant BaseData fields so tests can
// assert on them without depending on the default templates' markup.
func seoTemplateFS(body string) fstest.MapFS {
	return fstest.MapFS{
		"pages/index.tmpl":      {Data: []byte(body)},
		"pages/post.tmpl":       {Data: []byte(body)},
		"pages/tag.tmpl":        {Data: []byte(body)},
		"pages/tags-index.tmpl": {Data: []byte(body)},
	}
}

// TestGenerate_OGTypeInTemplateData verifies that post pages are typed as
// Open Graph articles and every other page type as a website.
func TestGenerate_OGTypeInTemplateData(t *testing.T) {
	t.Parallel()

	renderer, err := NewTemplateRenderer(seoTemplateFS(`{{.OGType}}`))
	if err != nil {
		t.Fatalf("NewTemplateRenderer() error = %v", err)
	}

	gen := New(newTestPostsFS(t, map[string]string{"hello.md": taggedPost}), renderer)
	blog, err := gen.Generate(context.Background())
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	for slug, content := range blog.Posts {
		if string(content) != "article" {
			t.Errorf("Post %q OGType = %q, want %q", slug, string(content), "article")
		}
	}
	if got := string(blog.Index); got != "website" {
		t.Errorf("Index OGType = %q, want %q", got, "website")
	}
	if got := string(blog.TagsIndex); got != "website" {
		t.Errorf("TagsIndex OGType = %q, want %q", got, "website")
	}
	if len(blog.Tags) == 0 {
		t.Fatal("expected at least one tag page")
	}
	for tag, content := range blog.Tags {
		if string(content) != "website" {
			t.Errorf("Tag %q OGType = %q, want %q", tag, string(content), "website")
		}
	}
}

// TestGenerate_ArticleMetaInTemplateData verifies that post pages carry an
// ArticleMeta mirroring the post's front matter, and that no other page type
// does.
func TestGenerate_ArticleMetaInTemplateData(t *testing.T) {
	t.Parallel()

	// "-" marks a nil ArticleMeta so an absent struct is distinguishable from
	// one whose fields are all empty.
	renderer, err := NewTemplateRenderer(seoTemplateFS(
		`{{with .Article}}{{.PublishedISO}}|{{.Author}}|{{range .Tags}}{{.}},{{end}}{{else}}-{{end}}`))
	if err != nil {
		t.Fatalf("NewTemplateRenderer() error = %v", err)
	}

	tests := []struct {
		name     string
		opts     []config.GeneratorOption
		wantPost string
	}{
		{
			name:     "post with author and tags",
			wantPost: "2024-06-01T09:30:00Z|Alice|go,testing,",
		},
		{
			name:     "tags disabled clears article tags",
			opts:     []config.GeneratorOption{config.WithDisableTags()},
			wantPost: "2024-06-01T09:30:00Z|Alice|",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gen := New(newTestPostsFS(t, map[string]string{"hello.md": taggedPost}), renderer, tt.opts...)
			blog, err := gen.Generate(context.Background())
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			for slug, content := range blog.Posts {
				if string(content) != tt.wantPost {
					t.Errorf("Post %q ArticleMeta = %q, want %q", slug, string(content), tt.wantPost)
				}
			}
			if got := string(blog.Index); got != "-" {
				t.Errorf("Index ArticleMeta = %q, want nil", got)
			}
			for tag, content := range blog.Tags {
				if string(content) != "-" {
					t.Errorf("Tag %q ArticleMeta = %q, want nil", tag, string(content))
				}
			}
			if blog.TagsIndex != nil {
				if got := string(blog.TagsIndex); got != "-" {
					t.Errorf("TagsIndex ArticleMeta = %q, want nil", got)
				}
			}
		})
	}
}

// TestGenerate_ArticleMetaOmitsMissingAuthor verifies that a post without an
// author in its front matter leaves ArticleMeta.Author empty rather than
// inventing a value.
func TestGenerate_ArticleMetaOmitsMissingAuthor(t *testing.T) {
	t.Parallel()

	renderer, err := NewTemplateRenderer(seoTemplateFS(`{{with .Article}}[{{.Author}}]{{end}}`))
	if err != nil {
		t.Fatalf("NewTemplateRenderer() error = %v", err)
	}

	gen := New(newTestPostsFS(t, map[string]string{"old.md": olderPost}), renderer)
	blog, err := gen.Generate(context.Background())
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	for slug, content := range blog.Posts {
		if string(content) != "[]" {
			t.Errorf("Post %q Author = %q, want empty", slug, string(content))
		}
	}
}

// TestGenerate_CanonicalURLInTemplateData verifies that CanonicalURL is the
// base URL joined with Path for every page type, and empty when no base URL
// is configured.
func TestGenerate_CanonicalURLInTemplateData(t *testing.T) {
	t.Parallel()

	renderer, err := NewTemplateRenderer(seoTemplateFS(`{{.CanonicalURL}}`))
	if err != nil {
		t.Fatalf("NewTemplateRenderer() error = %v", err)
	}

	tests := []struct {
		name          string
		baseURL       string
		blogRoot      string
		htmlPaths     bool
		wantIndex     string
		wantTagsIndex string
		wantPostFmt   func(slug string) string
	}{
		{
			name:          "no base URL leaves canonical empty",
			baseURL:       "",
			blogRoot:      "/",
			wantIndex:     "",
			wantTagsIndex: "",
			wantPostFmt:   func(string) string { return "" },
		},
		{
			name:          "base URL with clean URLs",
			baseURL:       "https://example.com",
			blogRoot:      "/",
			wantIndex:     "https://example.com/",
			wantTagsIndex: "https://example.com/tags",
			wantPostFmt:   func(slug string) string { return "https://example.com/posts/" + slug },
		},
		{
			name:          "trailing slash on base URL is trimmed",
			baseURL:       "https://example.com/",
			blogRoot:      "/",
			wantIndex:     "https://example.com/",
			wantTagsIndex: "https://example.com/tags",
			wantPostFmt:   func(slug string) string { return "https://example.com/posts/" + slug },
		},
		{
			name:          "sub-path blog root",
			baseURL:       "https://example.com",
			blogRoot:      "/blog/",
			wantIndex:     "https://example.com/blog/",
			wantTagsIndex: "https://example.com/blog/tags",
			wantPostFmt:   func(slug string) string { return "https://example.com/blog/posts/" + slug },
		},
		{
			name:          "html paths",
			baseURL:       "https://example.com",
			blogRoot:      "/",
			htmlPaths:     true,
			wantIndex:     "https://example.com/index.html",
			wantTagsIndex: "https://example.com/tags.html",
			wantPostFmt:   func(slug string) string { return "https://example.com/posts/" + slug + ".html" },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			opts := []config.GeneratorOption{
				config.WithBlogRoot(tt.blogRoot).AsGeneratorOption(),
				config.WithDisableFeeds(),
			}
			if tt.baseURL != "" {
				opts = append(opts, config.WithBaseURL(tt.baseURL))
			}
			if tt.htmlPaths {
				opts = append(opts, config.WithHTMLPaths())
			}

			gen := New(newTestPostsFS(t, map[string]string{
				"hello.md": simplePost,
				"old.md":   olderPost,
			}), renderer, opts...)

			blog, err := gen.Generate(context.Background())
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			if got := string(blog.Index); got != tt.wantIndex {
				t.Errorf("Index CanonicalURL = %q, want %q", got, tt.wantIndex)
			}
			if got := string(blog.TagsIndex); got != tt.wantTagsIndex {
				t.Errorf("TagsIndex CanonicalURL = %q, want %q", got, tt.wantTagsIndex)
			}
			for slug, content := range blog.Posts {
				if want := tt.wantPostFmt(slug); string(content) != want {
					t.Errorf("Post %q CanonicalURL = %q, want %q", slug, string(content), want)
				}
			}
		})
	}
}

// TestDefaultTemplates_OpenGraphTags renders the built-in templates and
// asserts on the Open Graph and canonical markup they emit for each page type.
func TestDefaultTemplates_OpenGraphTags(t *testing.T) {
	t.Parallel()

	postsFS := newTestPostsFS(t, map[string]string{
		"hello.md": taggedPost,
		"old.md":   olderPost, // no author, no tags
	})

	gen := New(postsFS, newTestRenderer(t), config.WithBaseURL("https://example.com"), config.WithSiteTitle("My Blog"))
	blog, err := gen.Generate(context.Background())
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	wantTags(t, "post", string(blog.Posts["hello-world"]), []string{
		`<meta property="og:type" content="article">`,
		`<meta property="og:site_name" content="My Blog">`,
		`<meta property="og:url" content="https://example.com/posts/hello-world">`,
		`<link rel="canonical" href="https://example.com/posts/hello-world">`,
		`<meta property="article:published_time" content="2024-06-01T09:30:00Z">`,
		`<meta property="article:author" content="Alice">`,
		`<meta property="article:tag" content="go">`,
		`<meta property="article:tag" content="testing">`,
	}, nil)

	// A post with neither author nor tags emits neither tag.
	wantTags(t, "authorless post", string(blog.Posts["old-post"]), []string{
		`<meta property="og:type" content="article">`,
		`<meta property="article:published_time" content="2024-01-01T00:00:00Z">`,
	}, []string{
		`article:author`,
		`article:tag`,
	})

	wantTags(t, "index", string(blog.Index), []string{
		`<meta property="og:type" content="website">`,
		`<meta property="og:site_name" content="My Blog">`,
		`<meta property="og:url" content="https://example.com/">`,
		`<link rel="canonical" href="https://example.com/">`,
	}, []string{
		`article:`,
	})

	wantTags(t, "tag", string(blog.Tags["go"]), []string{
		`<meta property="og:type" content="website">`,
		`<meta property="og:url" content="https://example.com/tags/go">`,
		`<link rel="canonical" href="https://example.com/tags/go">`,
	}, []string{
		`article:`,
	})

	wantTags(t, "tags index", string(blog.TagsIndex), []string{
		`<meta property="og:type" content="website">`,
		`<meta property="og:url" content="https://example.com/tags">`,
	}, []string{
		`article:`,
	})
}

// TestDefaultTemplates_NoBaseURLOmitsCanonical verifies that without a base
// URL the templates emit no canonical or og:url tag at all, rather than one
// with an empty value.
func TestDefaultTemplates_NoBaseURLOmitsCanonical(t *testing.T) {
	t.Parallel()

	postsFS := newTestPostsFS(t, map[string]string{"hello.md": taggedPost})
	gen := New(postsFS, newTestRenderer(t), config.WithSiteTitle("My Blog"))
	blog, err := gen.Generate(context.Background())
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	for name, page := range map[string][]byte{
		"post":  blog.Posts["hello-world"],
		"index": blog.Index,
	} {
		wantTags(t, name, string(page), []string{
			// og:site_name is emitted regardless of base URL.
			`<meta property="og:site_name" content="My Blog">`,
		}, []string{
			`og:url`,
			`rel="canonical"`,
		})
	}
}

// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package generator

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/harrydayexe/GoBlog/v2/pkg/config"
)

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

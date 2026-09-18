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
	"github.com/harrydayexe/GoBlog/v2/pkg/templates"
)

// newRobotsGen constructs a minimal Generator for buildRobotsTxt tests.
func newRobotsGen(opts ...config.GeneratorOption) *Generator {
	return New(fstest.MapFS{}, nil, append([]config.GeneratorOption{
		config.WithBaseURL("https://example.com"),
	}, opts...)...)
}

func TestBuildRobotsTxt_DefaultBody(t *testing.T) {
	t.Parallel()

	got := string(newRobotsGen().buildRobotsTxt(true))
	want := "User-agent: *\nAllow: /\n\nSitemap: https://example.com/sitemap.xml\n"

	if got != want {
		t.Errorf("robots.txt =\n%q\nwant\n%q", got, want)
	}
}

// TestBuildRobotsTxt_SitemapURLUsesBlogRoot asserts the Sitemap: line points at
// the sitemap's actual location under a sub-path deployment.
func TestBuildRobotsTxt_SitemapURLUsesBlogRoot(t *testing.T) {
	t.Parallel()

	gen := newRobotsGen(config.WithBlogRoot("/blog/").AsGeneratorOption())
	got := string(gen.buildRobotsTxt(true))

	if want := "Sitemap: https://example.com/blog/sitemap.xml\n"; !strings.HasSuffix(got, want) {
		t.Errorf("robots.txt =\n%q\nwant it to end with %q", got, want)
	}
}

func TestBuildRobotsTxt_CustomBody(t *testing.T) {
	t.Parallel()

	gen := newRobotsGen(config.WithRobotsTxt("User-agent: BadBot\nDisallow: /"))
	got := string(gen.buildRobotsTxt(true))
	want := "User-agent: BadBot\nDisallow: /\n\nSitemap: https://example.com/sitemap.xml\n"

	if got != want {
		t.Errorf("robots.txt =\n%q\nwant\n%q", got, want)
	}
	// The custom body replaces the default rules wholesale.
	if strings.Contains(got, "User-agent: *") {
		t.Errorf("custom body did not replace the default rules:\n%s", got)
	}
}

// TestBuildRobotsTxt_CustomBodyTrailingNewlines asserts that a body read from a
// file (which normally ends in a newline) does not gain a blank line before the
// Sitemap: separator.
func TestBuildRobotsTxt_CustomBodyTrailingNewlines(t *testing.T) {
	t.Parallel()

	gen := newRobotsGen(config.WithRobotsTxt("User-agent: *\nDisallow: /drafts/\n\n"))
	got := string(gen.buildRobotsTxt(true))
	want := "User-agent: *\nDisallow: /drafts/\n\nSitemap: https://example.com/sitemap.xml\n"

	if got != want {
		t.Errorf("robots.txt =\n%q\nwant\n%q", got, want)
	}
}

// TestBuildRobotsTxt_NoSitemap asserts no Sitemap: line is advertised when no
// sitemap was produced, whatever the rule body is.
func TestBuildRobotsTxt_NoSitemap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		opts []config.GeneratorOption
		want string
	}{
		{
			name: "default body",
			want: "User-agent: *\nAllow: /\n",
		},
		{
			name: "custom body",
			opts: []config.GeneratorOption{config.WithRobotsTxt("User-agent: *\nDisallow: /")},
			want: "User-agent: *\nDisallow: /\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := string(newRobotsGen(tt.opts...).buildRobotsTxt(false))
			if got != tt.want {
				t.Errorf("robots.txt =\n%q\nwant\n%q", got, tt.want)
			}
			if strings.Contains(got, "Sitemap:") {
				t.Errorf("robots.txt contains a Sitemap: line but no sitemap was produced:\n%s", got)
			}
		})
	}
}

// TestGenerate_RobotsTxtOmitsSitemapWhenNoneProduced asserts that a full
// Generate never advertises a sitemap that was not produced. An empty posts
// directory is reachable in normal operation — `serve --health-checks` starts
// before the posts volume is mounted, and a fresh blog has no posts yet — and
// buildSitemap emits nothing for it, so the Sitemap: line would 404.
func TestGenerate_RobotsTxtOmitsSitemapWhenNoneProduced(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		posts fstest.MapFS
		opts  []config.GeneratorOption
	}{
		{
			name:  "no posts",
			posts: fstest.MapFS{},
		},
		{
			name:  "sitemap disabled",
			posts: sitemapFS(),
			opts:  []config.GeneratorOption{config.WithDisableSitemap()},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			renderer, err := NewTemplateRenderer(templates.Default)
			if err != nil {
				t.Fatalf("NewTemplateRenderer: %v", err)
			}

			opts := append([]config.GeneratorOption{config.WithBaseURL("https://example.com")}, tt.opts...)
			blog, err := New(tt.posts, renderer, opts...).Generate(context.Background())
			if err != nil {
				t.Fatalf("Generate: %v", err)
			}

			if len(blog.Sitemap) != 0 {
				t.Fatalf("expected no sitemap, got:\n%s", blog.Sitemap)
			}
			if got, want := string(blog.RobotsTxt), "User-agent: *\nAllow: /\n"; got != want {
				t.Errorf("robots.txt =\n%q\nwant\n%q", got, want)
			}
		})
	}
}

// TestGenerate_RobotsTxtCustomBody asserts the option survives a full Generate.
func TestGenerate_RobotsTxtCustomBody(t *testing.T) {
	t.Parallel()

	blog := generateSitemapBlog(t,
		config.WithBaseURL("https://example.com"),
		config.WithRobotsTxt("User-agent: *\nDisallow: /private/"),
	)

	want := "User-agent: *\nDisallow: /private/\n\nSitemap: https://example.com/sitemap.xml\n"
	if got := string(blog.RobotsTxt); got != want {
		t.Errorf("robots.txt =\n%q\nwant\n%q", got, want)
	}
}

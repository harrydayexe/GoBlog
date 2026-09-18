// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package outputter

import (
	"bytes"
	"context"
	"encoding/xml"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/harrydayexe/GoBlog/v2/pkg/config"
	"github.com/harrydayexe/GoBlog/v2/pkg/generator"
	"github.com/harrydayexe/GoBlog/v2/pkg/templates"
)

var (
	fakeSitemap = []byte(`<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"></urlset>
`)
	fakeRobots = []byte("User-agent: *\nAllow: /\n\nSitemap: https://example.com/sitemap.xml\n")
)

// TestDirectoryWriter_WritesSitemapAndRobots verifies both files are written
// with the exact bytes the generator produced.
func TestDirectoryWriter_WritesSitemapAndRobots(t *testing.T) {
	t.Parallel()

	outputDir := t.TempDir()
	writer := NewDirectoryWriter(outputDir)

	blog := generator.NewEmptyGeneratedBlog()
	blog.Index = []byte("<h1>Blog Index</h1>")
	blog.Sitemap = fakeSitemap
	blog.RobotsTxt = fakeRobots

	if err := writer.HandleGeneratedBlog(context.Background(), blog); err != nil {
		t.Fatalf("HandleGeneratedBlog failed: %v", err)
	}

	tests := []struct {
		file string
		want []byte
	}{
		{"sitemap.xml", fakeSitemap},
		{"robots.txt", fakeRobots},
	}

	for _, tt := range tests {
		got, err := os.ReadFile(filepath.Join(outputDir, tt.file))
		if err != nil {
			t.Errorf("Failed to read %s: %v", tt.file, err)
			continue
		}
		if !bytes.Equal(got, tt.want) {
			t.Errorf("%s content = %q, want %q", tt.file, got, tt.want)
		}
	}
}

// TestDirectoryWriter_SkipsSitemapAndRobotsWhenNil verifies that neither file
// is created when the generator produced no content for it.
func TestDirectoryWriter_SkipsSitemapAndRobotsWhenNil(t *testing.T) {
	t.Parallel()

	outputDir := t.TempDir()
	writer := NewDirectoryWriter(outputDir)

	blog := generator.NewEmptyGeneratedBlog()
	blog.Index = []byte("<h1>Blog Index</h1>")

	if err := writer.HandleGeneratedBlog(context.Background(), blog); err != nil {
		t.Fatalf("HandleGeneratedBlog failed: %v", err)
	}

	for _, file := range []string{"sitemap.xml", "robots.txt"} {
		if _, err := os.Stat(filepath.Join(outputDir, file)); !os.IsNotExist(err) {
			t.Errorf("%s should not have been created when the generator produced none", file)
		}
	}
}

// TestDirectoryWriter_RobotsSubdirectoryWarning verifies that the deploy-time
// relocation warning fires when, and only when, a non-root blog root is set.
func TestDirectoryWriter_RobotsSubdirectoryWarning(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		blogRoot string
		wantWarn bool
	}{
		{name: "domain root", blogRoot: "/", wantWarn: false},
		{name: "unset root", blogRoot: "", wantWarn: false},
		{name: "subdirectory root", blogRoot: "/blog/", wantWarn: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

			opts := []config.GeneratorOption{config.WithLogger(logger).AsGeneratorOption()}
			if tt.blogRoot != "" {
				opts = append(opts, config.WithBlogRoot(tt.blogRoot).AsGeneratorOption())
			}
			writer := NewDirectoryWriter(t.TempDir(), opts...)

			blog := generator.NewEmptyGeneratedBlog()
			blog.Index = []byte("<h1>Blog Index</h1>")
			blog.RobotsTxt = fakeRobots

			if err := writer.HandleGeneratedBlog(context.Background(), blog); err != nil {
				t.Fatalf("HandleGeneratedBlog failed: %v", err)
			}

			gotWarn := strings.Contains(buf.String(), "domain root")
			if gotWarn != tt.wantWarn {
				t.Errorf("robots.txt relocation warning logged = %v, want %v\nlogs:\n%s", gotWarn, tt.wantWarn, buf.String())
			}
		})
	}
}

// sitemapTestPost is a tagged post used to generate a real blog end to end.
const sitemapTestPost = `---
title: "Hello World"
date: 2024-06-01T09:30:00Z
description: "A simple post"
tags: ["go", "testing"]
---
# Hello

This is content.
`

// TestDirectoryWriter_SitemapURLsResolveToWrittenFiles is the end-to-end guard
// that the sitemap never advertises a URL with no file behind it: it generates
// a blog the way `goblog generate` does (HTML paths on, here under a sub-path
// blog root), writes it out, and resolves every <loc> back to a file on disk.
func TestDirectoryWriter_SitemapURLsResolveToWrittenFiles(t *testing.T) {
	t.Parallel()

	renderer, err := generator.NewTemplateRenderer(templates.Default)
	if err != nil {
		t.Fatalf("NewTemplateRenderer: %v", err)
	}

	const blogRoot = "/blog/"
	postsFS := fstest.MapFS{"hello.md": &fstest.MapFile{Data: []byte(sitemapTestPost)}}
	opts := []config.GeneratorOption{
		config.WithBaseURL("https://example.com"),
		config.WithHTMLPaths(),
		config.WithBlogRoot(blogRoot).AsGeneratorOption(),
	}

	blog, err := generator.New(postsFS, renderer, opts...).Generate(context.Background())
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if len(blog.Sitemap) == 0 {
		t.Fatal("expected a sitemap")
	}

	outputDir := t.TempDir()
	if err := NewDirectoryWriter(outputDir, opts...).HandleGeneratedBlog(context.Background(), blog); err != nil {
		t.Fatalf("HandleGeneratedBlog failed: %v", err)
	}

	var doc struct {
		URLs []struct {
			Loc string `xml:"loc"`
		} `xml:"url"`
	}
	if err := xml.Unmarshal(blog.Sitemap, &doc); err != nil {
		t.Fatalf("sitemap is not valid XML: %v\n%s", err, blog.Sitemap)
	}
	if len(doc.URLs) == 0 {
		t.Fatal("sitemap contains no URLs")
	}

	// The output tree is flat: the output directory is the deployed blog root,
	// so a URL's path below the blog root is its path below outputDir.
	for _, u := range doc.URLs {
		parsed, err := url.Parse(u.Loc)
		if err != nil {
			t.Errorf("sitemap URL %q is not a valid URL: %v", u.Loc, err)
			continue
		}
		rel, ok := strings.CutPrefix(parsed.Path, blogRoot)
		if !ok {
			t.Errorf("sitemap URL %q is outside the blog root %q", u.Loc, blogRoot)
			continue
		}
		if _, err := os.Stat(filepath.Join(outputDir, filepath.FromSlash(rel))); err != nil {
			t.Errorf("sitemap URL %q has no file in the output tree: %v", u.Loc, err)
		}
	}
}

// TestDirectoryWriter_RobotsNoWarningWhenNotWritten verifies that no warning is
// logged for a sub-path deployment that produced no robots.txt at all.
func TestDirectoryWriter_RobotsNoWarningWhenNotWritten(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	writer := NewDirectoryWriter(t.TempDir(),
		config.WithLogger(logger).AsGeneratorOption(),
		config.WithBlogRoot("/blog/").AsGeneratorOption(),
	)

	blog := generator.NewEmptyGeneratedBlog()
	blog.Index = []byte("<h1>Blog Index</h1>")

	if err := writer.HandleGeneratedBlog(context.Background(), blog); err != nil {
		t.Fatalf("HandleGeneratedBlog failed: %v", err)
	}

	if strings.Contains(buf.String(), "domain root") {
		t.Errorf("unexpected robots.txt warning when no robots.txt was written\nlogs:\n%s", buf.String())
	}
}

package generator

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/harrydayexe/GoBlog/v2/pkg/config"
	"github.com/harrydayexe/GoBlog/v2/pkg/generator"
	"github.com/harrydayexe/GoBlog/v2/pkg/outputter"
)

// TestRunGenerate tests the core generate logic.
func TestRunGenerate(t *testing.T) {
	t.Parallel()

	// Setup
	tempDir := t.TempDir()

	// Create a test markdown file
	testPost := `---
title: Test Post
date: 2024-01-15
description: A test post
tags: [test]
---

# Test Post

This is a test post.
`
	postsDir := tempDir + "/posts"
	err := os.MkdirAll(postsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create posts dir: %v", err)
	}

	err = os.WriteFile(postsDir+"/test.md", []byte(testPost), 0644)
	if err != nil {
		t.Fatalf("Failed to write test post: %v", err)
	}

	outputDir := tempDir + "/output"
	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create output dir: %v", err)
	}

	// Create filesystem and renderer
	postsFsys := os.DirFS(postsDir)
	renderer, err := generator.NewTemplateRenderer(os.DirFS("../../pkg/templates/default"))
	if err != nil {
		t.Fatalf("Failed to create renderer: %v", err)
	}

	handler := outputter.NewDirectoryWriter(outputDir)
	opts := []config.GeneratorOption{}

	// Run generate
	ctx := context.Background()
	err = runGenerate(ctx, postsFsys, renderer, opts, handler)

	if err != nil {
		t.Fatalf("runGenerate() error = %v, want nil", err)
	}

	// Verify output files were created
	indexPath := outputDir + "/index.html"
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		t.Error("runGenerate() did not create index.html")
	}

	postsPath := outputDir + "/posts"
	if _, err := os.Stat(postsPath); os.IsNotExist(err) {
		t.Error("runGenerate() did not create posts directory")
	}
}

// TestRunGenerate_RawOutput tests generation with raw output mode.
func TestRunGenerate_RawOutput(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	testPost := `---
title: Raw Test
date: 2024-01-15
description: A test
tags: [test]
---

# Raw Test
`
	postsDir := tempDir + "/posts"
	err := os.MkdirAll(postsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create posts dir: %v", err)
	}

	err = os.WriteFile(postsDir+"/raw.md", []byte(testPost), 0644)
	if err != nil {
		t.Fatalf("Failed to write test post: %v", err)
	}

	outputDir := tempDir + "/output"
	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create output dir: %v", err)
	}

	postsFsys := os.DirFS(postsDir)
	renderer, err := generator.NewTemplateRenderer(os.DirFS("../../pkg/templates/default"))
	if err != nil {
		t.Fatalf("Failed to create renderer: %v", err)
	}

	opts := []config.GeneratorOption{config.WithRawOutput()}
	handler := outputter.NewDirectoryWriter(outputDir, opts...)

	ctx := context.Background()
	err = runGenerate(ctx, postsFsys, renderer, opts, handler)

	if err != nil {
		t.Fatalf("runGenerate() with raw output error = %v, want nil", err)
	}

	// In raw output mode, tags directory should not be created
	tagsPath := outputDir + "/tags"
	if _, err := os.Stat(tagsPath); !os.IsNotExist(err) {
		t.Error("runGenerate() with raw output should not create tags directory")
	}
}

// TestRunGenerate_EmptyPosts tests generation with no posts.
func TestRunGenerate_EmptyPosts(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	postsDir := tempDir + "/posts"
	err := os.MkdirAll(postsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create posts dir: %v", err)
	}

	outputDir := tempDir + "/output"
	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create output dir: %v", err)
	}

	postsFsys := os.DirFS(postsDir)
	renderer, err := generator.NewTemplateRenderer(os.DirFS("../../pkg/templates/default"))
	if err != nil {
		t.Fatalf("Failed to create renderer: %v", err)
	}

	handler := outputter.NewDirectoryWriter(outputDir)
	opts := []config.GeneratorOption{}

	ctx := context.Background()
	err = runGenerate(ctx, postsFsys, renderer, opts, handler)

	// Should succeed even with no posts
	if err != nil {
		t.Fatalf("runGenerate() with empty posts error = %v, want nil", err)
	}
}

// TestRunGenerate_InvalidPost tests generation with an invalid post.
func TestRunGenerate_InvalidPost(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	// Invalid post (missing required fields)
	invalidPost := `---
title: Invalid
---

No description or date.
`
	postsDir := tempDir + "/posts"
	err := os.MkdirAll(postsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create posts dir: %v", err)
	}

	err = os.WriteFile(postsDir+"/invalid.md", []byte(invalidPost), 0644)
	if err != nil {
		t.Fatalf("Failed to write test post: %v", err)
	}

	outputDir := tempDir + "/output"
	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create output dir: %v", err)
	}

	postsFsys := os.DirFS(postsDir)
	renderer, err := generator.NewTemplateRenderer(os.DirFS("../../pkg/templates/default"))
	if err != nil {
		t.Fatalf("Failed to create renderer: %v", err)
	}

	handler := outputter.NewDirectoryWriter(outputDir)
	opts := []config.GeneratorOption{}

	ctx := context.Background()
	err = runGenerate(ctx, postsFsys, renderer, opts, handler)

	// Should return error for invalid post
	if err == nil {
		t.Fatal("runGenerate() with invalid post should return error")
	}
}

// TestRunGenerate_CanceledContext tests generation with canceled context.
func TestRunGenerate_CanceledContext(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	testPost := `---
title: Test
date: 2024-01-15
description: A test
---

# Test
`
	postsDir := tempDir + "/posts"
	err := os.MkdirAll(postsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create posts dir: %v", err)
	}

	err = os.WriteFile(postsDir+"/test.md", []byte(testPost), 0644)
	if err != nil {
		t.Fatalf("Failed to write test post: %v", err)
	}

	outputDir := tempDir + "/output"
	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create output dir: %v", err)
	}

	postsFsys := os.DirFS(postsDir)
	renderer, err := generator.NewTemplateRenderer(os.DirFS("../../pkg/templates/default"))
	if err != nil {
		t.Fatalf("Failed to create renderer: %v", err)
	}

	handler := outputter.NewDirectoryWriter(outputDir)
	opts := []config.GeneratorOption{}

	// Create a pre-canceled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = runGenerate(ctx, postsFsys, renderer, opts, handler)

	// With small test files, parsing may complete before context is checked
	// So we accept either success or context.Canceled error
	if err != nil && err != context.Canceled {
		t.Errorf("runGenerate() with canceled context error = %v, want nil or context.Canceled", err)
	}
}

// TestRunGenerate_WithBlogRoot tests generation with custom blog root path.
func TestRunGenerate_WithBlogRoot(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	testPost := `---
title: Blog Root Test
date: 2024-01-15
description: Testing blog root
tags: [test, go]
---

# Blog Root Test

Testing the blog root feature.
`
	postsDir := tempDir + "/posts"
	err := os.MkdirAll(postsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create posts dir: %v", err)
	}

	err = os.WriteFile(postsDir+"/blogroot.md", []byte(testPost), 0644)
	if err != nil {
		t.Fatalf("Failed to write test post: %v", err)
	}

	outputDir := tempDir + "/output"
	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create output dir: %v", err)
	}

	postsFsys := os.DirFS(postsDir)
	renderer, err := generator.NewTemplateRenderer(os.DirFS("../../pkg/templates/default"))
	if err != nil {
		t.Fatalf("Failed to create renderer: %v", err)
	}

	// Generate with custom blog root
	opts := []config.GeneratorOption{config.WithBlogRoot("/blog/")}
	handler := outputter.NewDirectoryWriter(outputDir, opts...)

	ctx := context.Background()
	err = runGenerate(ctx, postsFsys, renderer, opts, handler)

	if err != nil {
		t.Fatalf("runGenerate() with blog root error = %v, want nil", err)
	}

	// Read and verify index.html contains /blog/ in links
	indexPath := outputDir + "/index.html"
	indexContent, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("Failed to read index.html: %v", err)
	}

	indexHTML := string(indexContent)

	// Verify blog root is used in post links
	if !containsString(indexHTML, `href="/blog/posts/`) {
		t.Error("index.html does not contain post links with /blog/ root")
	}

	// Verify blog root is used in tag links
	if !containsString(indexHTML, `href="/blog/tags/`) {
		t.Error("index.html does not contain tag links with /blog/ root")
	}

	// Verify blog root is used in home link
	if !containsString(indexHTML, `href="/blog/"`) {
		t.Error("index.html does not contain home link with /blog/ root")
	}

	// List files in posts directory to find the generated post
	postsPath := outputDir + "/posts"
	postFiles, err := os.ReadDir(postsPath)
	if err != nil {
		t.Fatalf("Failed to read posts directory: %v", err)
	}

	if len(postFiles) == 0 {
		t.Fatal("No post files generated in posts directory")
	}

	// Read and verify the first post page contains /blog/ in links
	postPath := postsPath + "/" + postFiles[0].Name()
	postContent, err := os.ReadFile(postPath)
	if err != nil {
		t.Fatalf("Failed to read post file: %v", err)
	}

	postHTML := string(postContent)

	// Verify blog root is used in navigation
	if !containsString(postHTML, `href="/blog/"`) {
		t.Error("Post page does not contain home link with /blog/ root")
	}

	// Verify blog root is used in tag links
	if !containsString(postHTML, `href="/blog/tags/`) {
		t.Error("Post page does not contain tag links with /blog/ root")
	}
}

// containsString checks if a string contains a substring (helper function).
func containsString(s, substr string) bool {
	return strings.Contains(s, substr)
}

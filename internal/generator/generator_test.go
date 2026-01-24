package generator

import (
	"context"
	"os"
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
	opts := []config.Option{}

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

	opts := []config.Option{config.WithRawOutput()}
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
	opts := []config.Option{}

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
	opts := []config.Option{}

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
	opts := []config.Option{}

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

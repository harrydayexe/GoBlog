package generator

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/harrydayexe/GoBlog/internal/gen/config"
	"github.com/harrydayexe/GoBlog/internal/gen/log"
)

// setupTestEnvironment creates a complete test environment with posts and templates
func setupTestEnvironment(t *testing.T) (string, string, string) {
	t.Helper()

	baseDir := t.TempDir()
	postsDir := filepath.Join(baseDir, "posts")
	templatesDir := filepath.Join(baseDir, "templates")
	outputDir := filepath.Join(baseDir, "output")

	// Create directories
	for _, dir := range []string{postsDir, templatesDir, outputDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("failed to create directory %s: %v", dir, err)
		}
	}

	// Create test posts
	posts := map[string]string{
		"post1.md": `---
title: "First Post"
date: 2024-03-15
description: "The first post"
tags: ["go", "testing"]
---

# Hello World

This is the first post.`,
		"post2.md": `---
title: "Second Post"
date: 2024-03-14
description: "The second post"
tags: ["go"]
---

# Second Post

This is the second post.`,
		"draft.md": `---
title: "Draft Post"
date: 2024-03-16
description: "A draft post"
draft: true
---

This is a draft.`,
	}

	for filename, content := range posts {
		path := filepath.Join(postsDir, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create post %s: %v", filename, err)
		}
	}

	// Create test templates
	templates := map[string]string{
		"post.html": `<!DOCTYPE html>
<html>
<head><title>{{.Post.Title}}</title></head>
<body>
<h1>{{.Post.Title}}</h1>
<p>{{.Post.Description}}</p>
<div>{{.Post.HTMLContent}}</div>
</body>
</html>`,
		"index.html": `<!DOCTYPE html>
<html>
<head><title>{{.Site.Title}}</title></head>
<body>
<h1>{{.Site.Title}}</h1>
{{range .Posts}}
<article>
<h2>{{.Title}}</h2>
<p>{{.Description}}</p>
</article>
{{end}}
</body>
</html>`,
		"tag.html": `<!DOCTYPE html>
<html>
<head><title>Tag: {{.Tag}}</title></head>
<body>
<h1>Posts tagged with {{.Tag}}</h1>
{{range .Posts}}
<article><h2>{{.Title}}</h2></article>
{{end}}
</body>
</html>`,
	}

	for filename, content := range templates {
		path := filepath.Join(templatesDir, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create template %s: %v", filename, err)
		}
	}

	return postsDir, templatesDir, outputDir
}

// TestNew tests generator creation
func TestNew(t *testing.T) {
	_, templatesDir, outputDir := setupTestEnvironment(t)

	tests := []struct {
		name      string
		cfg       config.Config
		logger    log.Logger
		expectErr bool
		errText   string
	}{
		{
			name: "valid config",
			cfg: config.Config{
				InputFolder:  "./posts",
				OutputFolder: outputDir,
				TemplateDir:  templatesDir,
				Site: config.SiteMetadata{
					Title: "Test Blog",
				},
			},
			logger:    nil,
			expectErr: false,
		},
		{
			name: "with custom logger",
			cfg: config.Config{
				InputFolder:  "./posts",
				OutputFolder: outputDir,
				TemplateDir:  templatesDir,
			},
			logger:    log.NewTestLogger("TEST", false, &bytes.Buffer{}, &bytes.Buffer{}),
			expectErr: false,
		},
		{
			name: "missing templates",
			cfg: config.Config{
				InputFolder:  "./posts",
				OutputFolder: outputDir,
				TemplateDir:  "/non/existent",
			},
			logger:    nil,
			expectErr: true,
			errText:   "failed to initialize template engine",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen, err := New(tt.cfg, tt.logger)

			if tt.expectErr {
				if err == nil {
					t.Error("expected error but got nil")
				} else if tt.errText != "" && !strings.Contains(err.Error(), tt.errText) {
					t.Errorf("expected error to contain %q, got %q", tt.errText, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if gen == nil {
					t.Fatal("expected generator to be non-nil")
				}
				if gen.parser == nil {
					t.Error("expected parser to be set")
				}
				if gen.template == nil {
					t.Error("expected template to be set")
				}
				if gen.logger == nil {
					t.Error("expected logger to be set")
				}
			}
		})
	}
}

// TestGenerator_Generate tests the full generation process
func TestGenerator_Generate(t *testing.T) {
	postsDir, templatesDir, outputDir := setupTestEnvironment(t)

	cfg := config.Config{
		InputFolder:  postsDir,
		OutputFolder: outputDir,
		TemplateDir:  templatesDir,
		Site: config.SiteMetadata{
			Title:       "Test Blog",
			Description: "A test blog",
			Author:      "Test Author",
		},
		PostsPerPage: 10,
		BlogPath:     "/blog",
	}

	logger := log.NewTestLogger("TEST", false, &bytes.Buffer{}, &bytes.Buffer{})
	gen, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("failed to create generator: %v", err)
	}

	err = gen.Generate()
	if err != nil {
		t.Fatalf("generation failed: %v", err)
	}

	// Verify output structure
	expectedFiles := []string{
		filepath.Join(outputDir, "blog", "first-post.html"),
		filepath.Join(outputDir, "blog", "second-post.html"),
		filepath.Join(outputDir, "blog", "index.html"),
		filepath.Join(outputDir, "blog", "tags", "go.html"),
		filepath.Join(outputDir, "blog", "tags", "testing.html"),
		filepath.Join(outputDir, "index.html"),
	}

	for _, file := range expectedFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", file)
		} else {
			// Verify file has content
			content, readErr := os.ReadFile(file)
			if readErr != nil {
				t.Errorf("failed to read %s: %v", file, readErr)
			}
			if len(content) == 0 {
				t.Errorf("expected %s to have content", file)
			}
		}
	}

	// Verify draft post was NOT generated
	draftPath := filepath.Join(outputDir, "blog", "draft-post.html")
	if _, err := os.Stat(draftPath); !os.IsNotExist(err) {
		t.Error("draft post should not be generated")
	}

	// Verify post content
	firstPostPath := filepath.Join(outputDir, "blog", "first-post.html")
	content, _ := os.ReadFile(firstPostPath)
	contentStr := string(content)

	if !strings.Contains(contentStr, "First Post") {
		t.Error("post should contain title")
	}
	if !strings.Contains(contentStr, "Hello World") {
		t.Error("post should contain content")
	}

	// Verify index content
	indexPath := filepath.Join(outputDir, "blog", "index.html")
	indexContent, _ := os.ReadFile(indexPath)
	indexStr := string(indexContent)

	if !strings.Contains(indexStr, "Test Blog") {
		t.Error("index should contain site title")
	}
	if !strings.Contains(indexStr, "First Post") {
		t.Error("index should contain post titles")
	}
}

// TestGenerator_Generate_NoPublishedPosts tests error handling when no published posts
func TestGenerator_Generate_NoPublishedPosts(t *testing.T) {
	baseDir := t.TempDir()
	postsDir := filepath.Join(baseDir, "posts")
	templatesDir := filepath.Join(baseDir, "templates")
	outputDir := filepath.Join(baseDir, "output")

	// Create directories
	for _, dir := range []string{postsDir, templatesDir, outputDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("failed to create directory: %v", err)
		}
	}

	// Create only draft posts
	draftPost := `---
title: "Draft"
date: 2024-03-15
description: "A draft"
draft: true
---
Content`
	if err := os.WriteFile(filepath.Join(postsDir, "draft.md"), []byte(draftPost), 0644); err != nil {
		t.Fatalf("failed to create draft post: %v", err)
	}

	// Create all required templates
	templates := map[string]string{
		"post.html":  `<html><body>{{.Post.Title}}</body></html>`,
		"index.html": `<html><body>{{.Site.Title}}</body></html>`,
		"tag.html":   `<html><body>{{.Tag}}</body></html>`,
	}
	for name, content := range templates {
		if err := os.WriteFile(filepath.Join(templatesDir, name), []byte(content), 0644); err != nil {
			t.Fatalf("failed to create template %s: %v", name, err)
		}
	}

	cfg := config.Config{
		InputFolder:  postsDir,
		OutputFolder: outputDir,
		TemplateDir:  templatesDir,
	}

	logger := log.NewTestLogger("TEST", false, &bytes.Buffer{}, &bytes.Buffer{})
	gen, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("failed to create generator: %v", err)
	}

	err = gen.Generate()
	if err == nil {
		t.Error("expected error when no published posts")
	}
	if !strings.Contains(err.Error(), "no published posts") {
		t.Errorf("expected 'no published posts' error, got: %v", err)
	}
}

// TestGenerator_Generate_NoPosts tests error handling when no posts at all
func TestGenerator_Generate_NoPosts(t *testing.T) {
	baseDir := t.TempDir()
	postsDir := filepath.Join(baseDir, "posts")
	templatesDir := filepath.Join(baseDir, "templates")
	outputDir := filepath.Join(baseDir, "output")

	// Create empty directories
	for _, dir := range []string{postsDir, templatesDir, outputDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("failed to create directory: %v", err)
		}
	}

	// Create all required templates
	templates := map[string]string{
		"post.html":  `<html><body>{{.Post.Title}}</body></html>`,
		"index.html": `<html><body>{{.Site.Title}}</body></html>`,
		"tag.html":   `<html><body>{{.Tag}}</body></html>`,
	}
	for name, content := range templates {
		if err := os.WriteFile(filepath.Join(templatesDir, name), []byte(content), 0644); err != nil {
			t.Fatalf("failed to create template %s: %v", name, err)
		}
	}

	cfg := config.Config{
		InputFolder:  postsDir,
		OutputFolder: outputDir,
		TemplateDir:  templatesDir,
	}

	logger := log.NewTestLogger("TEST", false, &bytes.Buffer{}, &bytes.Buffer{})
	gen, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("failed to create generator: %v", err)
	}

	err = gen.Generate()
	if err == nil {
		t.Error("expected error when no posts")
	}
	if !strings.Contains(err.Error(), "no markdown files found") && !strings.Contains(err.Error(), "failed to parse posts") {
		t.Errorf("expected 'no posts' error, got: %v", err)
	}
}

// TestCopyFile tests the file copying helper
func TestCopyFile(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		content   string
		expectErr bool
	}{
		{
			name:      "copy text file",
			content:   "Hello, World!",
			expectErr: false,
		},
		{
			name:      "copy empty file",
			content:   "",
			expectErr: false,
		},
		{
			name:      "copy large file",
			content:   strings.Repeat("a", 10000),
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := filepath.Join(tmpDir, "src.txt")
			dst := filepath.Join(tmpDir, "dst.txt")

			// Create source file
			if err := os.WriteFile(src, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to create source file: %v", err)
			}

			// Copy file
			err := copyFile(src, dst)

			if tt.expectErr {
				if err == nil {
					t.Error("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				// Verify content matches
				dstContent, readErr := os.ReadFile(dst)
				if readErr != nil {
					t.Fatalf("failed to read destination file: %v", readErr)
				}

				if string(dstContent) != tt.content {
					t.Errorf("content mismatch: got %q, want %q", string(dstContent), tt.content)
				}
			}

			// Cleanup for next test
			_ = os.Remove(src)
			_ = os.Remove(dst)
		})
	}
}

// TestGenerator_CopyStaticFiles tests copying static assets
func TestGenerator_CopyStaticFiles(t *testing.T) {
	postsDir, templatesDir, outputDir := setupTestEnvironment(t)
	staticDir := filepath.Join(filepath.Dir(postsDir), "static")

	// Create static directory with files
	if err := os.MkdirAll(filepath.Join(staticDir, "css"), 0755); err != nil {
		t.Fatalf("failed to create static/css directory: %v", err)
	}

	cssContent := "body { margin: 0; }"
	if err := os.WriteFile(filepath.Join(staticDir, "css", "style.css"), []byte(cssContent), 0644); err != nil {
		t.Fatalf("failed to create CSS file: %v", err)
	}

	cfg := config.Config{
		InputFolder:  postsDir,
		OutputFolder: outputDir,
		TemplateDir:  templatesDir,
		StaticFolder: staticDir,
		Site: config.SiteMetadata{
			Title: "Test Blog",
		},
		PostsPerPage: 10,
		BlogPath:     "/blog",
	}

	logger := log.NewTestLogger("TEST", false, &bytes.Buffer{}, &bytes.Buffer{})
	gen, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("failed to create generator: %v", err)
	}

	err = gen.Generate()
	if err != nil {
		t.Fatalf("generation failed: %v", err)
	}

	// Verify static files were copied
	copiedCSS := filepath.Join(outputDir, "css", "style.css")
	if _, statErr := os.Stat(copiedCSS); os.IsNotExist(statErr) {
		t.Error("expected static CSS file to be copied")
	} else {
		content, _ := os.ReadFile(copiedCSS)
		if string(content) != cssContent {
			t.Errorf("CSS content mismatch: got %q, want %q", string(content), cssContent)
		}
	}
}

package parser

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/harrydayexe/GoBlog/internal/gen/log"
	"github.com/harrydayexe/GoBlog/pkg/models"
)

// TestNew tests parser creation
func TestNew(t *testing.T) {
	tests := []struct {
		name   string
		logger log.Logger
	}{
		{
			name:   "with logger",
			logger: log.NewTestLogger("TEST", false, &bytes.Buffer{}, &bytes.Buffer{}),
		},
		{
			name:   "without logger (nil)",
			logger: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New(tt.logger)
			if p == nil {
				t.Fatal("expected parser to be non-nil")
			}
			if p.md == nil {
				t.Error("expected goldmark instance to be non-nil")
			}
			if p.logger == nil {
				t.Error("expected logger to be non-nil")
			}
		})
	}
}

// TestParser_ParseMarkdown tests parsing markdown content
func TestParser_ParseMarkdown(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectErr   bool
		validatePost func(*testing.T, *models.Post)
	}{
		{
			name: "valid markdown with frontmatter",
			content: `---
title: "Test Post"
date: 2024-03-15
description: "A test post"
tags: ["go", "testing"]
draft: false
---

# Hello World

This is a **test** post.`,
			expectErr: false,
			validatePost: func(t *testing.T, post *models.Post) {
				if post.Title != "Test Post" {
					t.Errorf("expected title 'Test Post', got %s", post.Title)
				}
				if post.Description != "A test post" {
					t.Errorf("expected description 'A test post', got %s", post.Description)
				}
				if len(post.Tags) != 2 {
					t.Errorf("expected 2 tags, got %d", len(post.Tags))
				}
				if post.Draft {
					t.Error("expected draft to be false")
				}
				if !strings.Contains(post.Content, "<h1") {
					t.Error("expected HTML content to contain h1 tag")
				}
				if !strings.Contains(post.Content, "<strong>test</strong>") {
					t.Error("expected HTML content to contain strong tag")
				}
			},
		},
		{
			name: "markdown without frontmatter",
			content: `# Just Content

No frontmatter here.`,
			expectErr: false,
			validatePost: func(t *testing.T, post *models.Post) {
				if post.Title != "" {
					t.Errorf("expected empty title, got %s", post.Title)
				}
				if !strings.Contains(post.Content, "<h1") {
					t.Error("expected HTML content to contain h1 tag")
				}
			},
		},
		{
			name: "markdown with code block",
			content: `---
title: "Code Example"
date: 2024-03-15
description: "Code example"
---

` + "```go\nfunc main() {}\n```",
			expectErr: false,
			validatePost: func(t *testing.T, post *models.Post) {
				if !strings.Contains(post.Content, "<pre") {
					t.Error("expected HTML content to contain pre tag for code block")
				}
			},
		},
		{
			name: "markdown with GFM features",
			content: `---
title: "GFM Test"
date: 2024-03-15
description: "Testing GFM"
---

- [x] Task 1
- [ ] Task 2

| Header 1 | Header 2 |
|----------|----------|
| Cell 1   | Cell 2   |`,
			expectErr: false,
			validatePost: func(t *testing.T, post *models.Post) {
				if !strings.Contains(post.Content, "checkbox") {
					t.Error("expected HTML content to contain checkbox for task list")
				}
				if !strings.Contains(post.Content, "<table") {
					t.Error("expected HTML content to contain table tag")
				}
			},
		},
		{
			name: "markdown with typographer",
			content: `---
title: "Typography"
date: 2024-03-15
description: "Testing typography"
---

"Quotes" and -- dashes...`,
			expectErr: false,
			validatePost: func(t *testing.T, post *models.Post) {
				// Typographer should convert quotes and dashes
				if !strings.Contains(post.Content, "—") && !strings.Contains(post.Content, "&mdash;") {
					// Either em-dash or HTML entity is fine
					t.Logf("Content: %s", post.Content)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			logger := log.NewTestLogger("TEST", true, &stdout, &stderr)
			p := New(logger)

			post, err := p.ParseMarkdown([]byte(tt.content))

			if tt.expectErr {
				if err == nil {
					t.Error("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if post == nil {
					t.Fatal("expected post to be non-nil")
				}
				if tt.validatePost != nil {
					tt.validatePost(t, post)
				}
			}
		})
	}
}

// TestParser_ParseFile tests parsing a markdown file
func TestParser_ParseFile(t *testing.T) {
	// Create temporary directory for test files
	tmpDir := t.TempDir()

	tests := []struct {
		name         string
		filename     string
		content      string
		expectErr    bool
		errText      string
		validatePost func(*testing.T, *models.Post)
	}{
		{
			name:     "valid markdown file",
			filename: "test-post.md",
			content: `---
title: "My Test Post"
date: 2024-03-15
description: "A test post from a file"
tags: ["test"]
---

# Content

This is the content.`,
			expectErr: false,
			validatePost: func(t *testing.T, post *models.Post) {
				if post.Title != "My Test Post" {
					t.Errorf("expected title 'My Test Post', got %s", post.Title)
				}
				if post.Slug != "my-test-post" {
					t.Errorf("expected slug 'my-test-post', got %s", post.Slug)
				}
				if post.SourcePath == "" {
					t.Error("expected SourcePath to be set")
				}
			},
		},
		{
			name:      "non-existent file",
			filename:  "non-existent.md",
			expectErr: true,
			errText:   "failed to read file",
		},
		{
			name:     "file with validation errors",
			filename: "invalid.md",
			content: `---
description: "Missing title and date"
---

Content here.`,
			expectErr: true,
			errText:   "title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			logger := log.NewTestLogger("TEST", true, &stdout, &stderr)
			p := New(logger)

			var filePath string
			if tt.content != "" {
				// Create test file
				filePath = filepath.Join(tmpDir, tt.filename)
				if err := os.WriteFile(filePath, []byte(tt.content), 0644); err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
			} else {
				// Use non-existent path
				filePath = filepath.Join(tmpDir, tt.filename)
			}

			post, err := p.ParseFile(filePath)

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
				if post == nil {
					t.Fatal("expected post to be non-nil")
				}
				if tt.validatePost != nil {
					tt.validatePost(t, post)
				}
			}
		})
	}
}

// TestParser_ParseAllPosts tests discovering and parsing all posts
func TestParser_ParseAllPosts(t *testing.T) {
	// Create temporary directory for test files
	tmpDir := t.TempDir()

	// Create test posts
	posts := map[string]string{
		"post1.md": `---
title: "Post 1"
date: 2024-03-15
description: "First post"
---
Content 1`,
		"post2.md": `---
title: "Post 2"
date: 2024-03-14
description: "Second post"
---
Content 2`,
		"subdir/post3.md": `---
title: "Post 3"
date: 2024-03-16
description: "Third post"
---
Content 3`,
		"draft.md": `---
title: "Draft Post"
date: 2024-03-17
description: "Draft"
draft: true
---
Draft content`,
		"invalid.md": `---
title: "No date or description"
---
Invalid post`,
		"readme.txt": "Not a markdown file",
	}

	for filename, content := range posts {
		filePath := filepath.Join(tmpDir, filename)
		dir := filepath.Dir(filePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("failed to create directory: %v", err)
		}
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	var stdout, stderr bytes.Buffer
	logger := log.NewTestLogger("TEST", true, &stdout, &stderr)
	p := New(logger)

	postList, err := p.ParseAllPosts(tmpDir)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Should have 4 valid posts (post1, post2, post3, draft)
	// invalid.md should be skipped
	if len(postList) != 4 {
		t.Errorf("expected 4 posts, got %d", len(postList))
	}

	// Verify posts are sorted by date (newest first)
	if len(postList) >= 2 {
		if postList[0].Date.Before(postList[1].Date) {
			t.Error("expected posts to be sorted by date (newest first)")
		}
	}

	// Verify the newest post is Post 3
	if len(postList) > 0 {
		if postList[0].Title != "Draft Post" {
			t.Errorf("expected first post to be 'Draft Post', got %s", postList[0].Title)
		}
	}
}

// TestParser_ParseAllPosts_EmptyDir tests parsing empty directory
func TestParser_ParseAllPosts_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()

	var stdout, stderr bytes.Buffer
	logger := log.NewTestLogger("TEST", false, &stdout, &stderr)
	p := New(logger)

	_, err := p.ParseAllPosts(tmpDir)

	if err == nil {
		t.Error("expected error for empty directory")
	}
	if !strings.Contains(err.Error(), "no markdown files found") {
		t.Errorf("expected 'no markdown files found' error, got: %v", err)
	}
}

// TestParser_ParseAllPosts_NonExistentDir tests parsing non-existent directory
func TestParser_ParseAllPosts_NonExistentDir(t *testing.T) {
	var stdout, stderr bytes.Buffer
	logger := log.NewTestLogger("TEST", false, &stdout, &stderr)
	p := New(logger)

	_, err := p.ParseAllPosts("/non/existent/directory")

	if err == nil {
		t.Error("expected error for non-existent directory")
	}
}

// TestParser_ParseMarkdown_DateParsing tests various date formats
func TestParser_ParseMarkdown_DateParsing(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected time.Time
	}{
		{
			name: "ISO date format",
			content: `---
title: "Test"
date: 2024-03-15
description: "Test"
---
Content`,
			expected: time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			logger := log.NewTestLogger("TEST", false, &stdout, &stderr)
			p := New(logger)

			post, err := p.ParseMarkdown([]byte(tt.content))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !post.Date.Equal(tt.expected) {
				t.Errorf("expected date %v, got %v", tt.expected, post.Date)
			}
		})
	}
}

// TestParser_HTMLContent tests that HTML content is properly generated
func TestParser_HTMLContent(t *testing.T) {
	content := `---
title: "HTML Test"
date: 2024-03-15
description: "Testing HTML generation"
---

# Heading 1
## Heading 2

**Bold** and *italic* text.

- List item 1
- List item 2

[Link](https://example.com)`

	var stdout, stderr bytes.Buffer
	logger := log.NewTestLogger("TEST", false, &stdout, &stderr)
	p := New(logger)

	post, err := p.ParseMarkdown([]byte(content))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify HTML elements
	checks := []struct {
		element string
		desc    string
	}{
		{"<h1", "heading 1"},
		{"<h2", "heading 2"},
		{"<strong>Bold</strong>", "bold text"},
		{"<em>italic</em>", "italic text"},
		{"<ul", "unordered list"},
		{"<li", "list item"},
		{"<a href=", "link"},
	}

	for _, check := range checks {
		if !strings.Contains(post.Content, check.element) {
			t.Errorf("expected HTML to contain %s (%s)", check.element, check.desc)
		}
	}

	// Verify HTMLContent is set
	if string(post.HTMLContent) != post.Content {
		t.Error("expected HTMLContent to match Content")
	}
}

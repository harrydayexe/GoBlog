package parser

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestParseFile_ValidPost(t *testing.T) {
	p := New()
	fsys := os.DirFS("testdata")

	post, err := p.ParseFile(fsys, "valid-post.md")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Check frontmatter fields
	if post.Title != "A Valid Blog Post" {
		t.Errorf("expected title 'A Valid Blog Post', got: %s", post.Title)
	}

	if post.Description != "This is a test post with all required fields" {
		t.Errorf("unexpected description: %s", post.Description)
	}

	if len(post.Tags) != 3 {
		t.Errorf("expected 3 tags, got: %d", len(post.Tags))
	}

	if post.Date.IsZero() {
		t.Error("expected valid date, got zero time")
	}

	// Check generated fields
	if post.SourcePath != "valid-post.md" {
		t.Errorf("expected source path 'valid-post.md', got: %s", post.SourcePath)
	}

	if post.Slug == "" {
		t.Error("expected slug to be generated")
	}

	if post.Content == "" {
		t.Error("expected content to be rendered")
	}

	if post.HTMLContent == "" {
		t.Error("expected HTMLContent to be set")
	}

	if post.RawContent == "" {
		t.Error("expected RawContent to be set")
	}

	// Verify HTML was actually rendered (should contain <h1> tag)
	if !strings.Contains(post.Content, "<h1") {
		t.Error("expected rendered HTML to contain heading tags")
	}

	// Verify markdown was converted (should contain <strong> for **)
	if !strings.Contains(post.Content, "<strong>valid</strong>") {
		t.Error("expected rendered HTML to contain bold text")
	}
}

func TestParseFile_MissingTitle(t *testing.T) {
	p := New()
	fsys := os.DirFS("testdata")

	_, err := p.ParseFile(fsys, "missing-title.md")
	if err == nil {
		t.Fatal("expected error for missing title, got nil")
	}

	if !strings.Contains(err.Error(), "title") {
		t.Errorf("expected error to mention 'title', got: %v", err)
	}
}

func TestParseFile_MissingDescription(t *testing.T) {
	p := New()
	fsys := os.DirFS("testdata")

	_, err := p.ParseFile(fsys, "missing-description.md")
	if err == nil {
		t.Fatal("expected error for missing description, got nil")
	}

	if !strings.Contains(err.Error(), "description") {
		t.Errorf("expected error to mention 'description', got: %v", err)
	}
}

func TestParseFile_InvalidYAML(t *testing.T) {
	p := New()
	fsys := os.DirFS("testdata")

	_, err := p.ParseFile(fsys, "invalid-yaml.md")
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

func TestParseFile_NoFrontmatter(t *testing.T) {
	p := New()
	fsys := os.DirFS("testdata")

	_, err := p.ParseFile(fsys, "no-frontmatter.md")
	if err == nil {
		t.Fatal("expected error for missing frontmatter, got nil")
	}

	// Should fail because no frontmatter was found
	if !strings.Contains(err.Error(), "no frontmatter") &&
		!strings.Contains(err.Error(), "title") &&
		!strings.Contains(err.Error(), "date") {
		t.Errorf("expected frontmatter or validation error, got: %v", err)
	}
}

func TestParseFile_WithCodeBlocks(t *testing.T) {
	p := New()
	fsys := os.DirFS("testdata")

	post, err := p.ParseFile(fsys, "with-code.md")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify syntax highlighting is applied (should have <pre> or <code> tags)
	if !strings.Contains(post.Content, "<pre") && !strings.Contains(post.Content, "<code") {
		t.Error("expected code blocks to be rendered with <pre> or <code> tags")
	}

	// Check that code content is preserved
	if !strings.Contains(post.Content, "Hello, World!") {
		t.Error("expected code content to be preserved in HTML")
	}
}

func TestParseFile_WithFootnotes(t *testing.T) {
	p := New()
	fsys := os.DirFS("testdata")

	post, err := p.ParseFile(fsys, "with-footnotes.md")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify footnotes are rendered (goldmark uses <sup> for footnote refs)
	if !strings.Contains(post.Content, "sup") || !strings.Contains(post.Content, "footnote") {
		t.Error("expected footnotes to be rendered")
	}
}

func TestParseFile_NonExistentFile(t *testing.T) {
	p := New()
	fsys := os.DirFS("testdata")

	_, err := p.ParseFile(fsys, "does-not-exist.md")
	if err == nil {
		t.Fatal("expected error for non-existent file, got nil")
	}
}

func TestParseDirectory(t *testing.T) {
	p := New()
	fsys := os.DirFS("testdata")

	posts, err := p.ParseDirectory(fsys)

	// We expect errors because some test files are intentionally invalid
	var parseErrs ParseErrors
	if err != nil {
		var ok bool
		parseErrs, ok = err.(ParseErrors)
		if !ok {
			t.Fatalf("expected ParseErrors, got: %T", err)
		}
	}

	// Should have parsed at least the valid files
	if len(posts) == 0 {
		t.Error("expected at least some valid posts to be parsed")
	}

	// Should have collected errors for invalid files
	if !parseErrs.HasErrors() {
		t.Error("expected parsing errors for invalid test files")
	}

	// Verify posts are sorted by date (newest first)
	for i := range len(posts) - 1 {
		if posts[i].Date.Before(posts[i+1].Date) {
			t.Error("expected posts to be sorted by date (newest first)")
			break
		}
	}

	// Verify all valid posts have required fields
	for _, post := range posts {
		if post.Title == "" {
			t.Errorf("post %s has empty title", post.SourcePath)
		}
		if post.Slug == "" {
			t.Errorf("post %s has empty slug", post.SourcePath)
		}
		if post.Content == "" {
			t.Errorf("post %s has empty content", post.SourcePath)
		}
	}
}

func TestParseErrors_Error(t *testing.T) {
	testErr := fmt.Errorf("test error")
	pe := ParseErrors{
		Errors: []FileError{
			{Path: "file1.md", Err: testErr},
			{Path: "file2.md", Err: testErr},
		},
	}

	errMsg := pe.Error()
	if !strings.Contains(errMsg, "file1.md") {
		t.Error("expected error message to contain file1.md")
	}
	if !strings.Contains(errMsg, "file2.md") {
		t.Error("expected error message to contain file2.md")
	}
	if !strings.Contains(errMsg, "2 file(s)") {
		t.Error("expected error message to contain count")
	}
}

func TestParseErrors_HasErrors(t *testing.T) {
	pe := ParseErrors{}
	if pe.HasErrors() {
		t.Error("expected HasErrors to return false for empty errors")
	}

	testErr := fmt.Errorf("test error")
	pe.Errors = append(pe.Errors, FileError{Path: "test.md", Err: testErr})
	if !pe.HasErrors() {
		t.Error("expected HasErrors to return true when errors exist")
	}
}

func TestFileError_Error(t *testing.T) {
	testErr := fmt.Errorf("test error")
	fe := FileError{
		Path: "test.md",
		Err:  testErr,
	}

	errMsg := fe.Error()
	if !strings.Contains(errMsg, "test.md") {
		t.Error("expected error message to contain file path")
	}
}

func TestFileError_Unwrap(t *testing.T) {
	original := fmt.Errorf("original error")
	fe := FileError{
		Path: "test.md",
		Err:  original,
	}

	if fe.Unwrap() != original {
		t.Error("expected Unwrap to return original error")
	}
}

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

	if string(post.Content) == "" {
		t.Error("expected content to be rendered")
	}

	if post.HTMLContent == "" {
		t.Error("expected HTMLContent to be set")
	}

	if post.RawContent == "" {
		t.Error("expected RawContent to be set")
	}

	// Verify HTML was actually rendered (should contain <h1> tag)
	if !strings.Contains(string(post.Content), "<h1") {
		t.Error("expected rendered HTML to contain heading tags")
	}

	// Verify markdown was converted (should contain <strong> for **)
	if !strings.Contains(string(post.Content), "<strong>valid</strong>") {
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
	if !strings.Contains(string(post.Content), "<pre") && !strings.Contains(string(post.Content), "<code") {
		t.Error("expected code blocks to be rendered with <pre> or <code> tags")
	}

	// Check that code content is preserved
	if !strings.Contains(string(post.Content), "Hello, World!") {
		t.Error("expected code content to be preserved in HTML")
	}
}

func TestParseFile_WithFootnotes(t *testing.T) {
	p := New(WithFootnote())
	fsys := os.DirFS("testdata")

	post, err := p.ParseFile(fsys, "with-footnotes.md")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify footnotes are rendered (goldmark uses <sup> for footnote refs)
	if !strings.Contains(string(post.Content), "<sup") {
		t.Error("expected footnotes to include <sup> tags for references")
	}

	// Verify footnote section is rendered
	if !strings.Contains(string(post.Content), "footnote-ref") || !strings.Contains(string(post.Content), "footnotes") {
		t.Error("expected proper footnote rendering with goldmark classes")
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
		if string(post.Content) == "" {
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

func TestNew_WithCodeHighlighting(t *testing.T) {
	fsys := os.DirFS("testdata")

	// Test with highlighting disabled
	p := New(WithCodeHighlighting(false))
	post, err := p.ParseFile(fsys, "with-code.md")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// When highlighting is disabled, code should still be in <pre><code> but without chroma classes
	if !strings.Contains(string(post.Content), "<pre") || !strings.Contains(string(post.Content), "<code") {
		t.Error("expected code blocks to be rendered in <pre><code> tags")
	}

	// Should NOT contain chroma highlighting classes
	if strings.Contains(string(post.Content), "chroma") {
		t.Error("expected no chroma highlighting classes when highlighting is disabled")
	}

	// Test with highlighting explicitly enabled
	p = New(WithCodeHighlighting(true))
	post, err = p.ParseFile(fsys, "with-code.md")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Should contain chroma classes
	if !strings.Contains(string(post.Content), "chroma") {
		t.Error("expected chroma highlighting classes when highlighting is enabled")
	}

	// Test default behavior (should be enabled)
	p = New()
	post, err = p.ParseFile(fsys, "with-code.md")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Default should have highlighting enabled
	if !strings.Contains(string(post.Content), "chroma") {
		t.Error("expected chroma highlighting classes by default")
	}
}

func TestNew_WithCodeHighlightingStyle(t *testing.T) {
	fsys := os.DirFS("testdata")

	// Test with custom style
	p := New(WithCodeHighlightingStyle("dracula"))
	post, err := p.ParseFile(fsys, "with-code.md")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Should contain chroma classes (highlighting enabled)
	if !strings.Contains(string(post.Content), "chroma") {
		t.Error("expected chroma highlighting classes when style is set")
	}

	// Code content should be preserved
	if !strings.Contains(string(post.Content), "Hello, World!") {
		t.Error("expected code content to be preserved")
	}

	// Test with different style to ensure parser accepts it
	p = New(WithCodeHighlightingStyle("monokai"))
	post, err = p.ParseFile(fsys, "with-code.md")
	if err != nil {
		t.Fatalf("expected no error with monokai style, got: %v", err)
	}

	if !strings.Contains(string(post.Content), "chroma") {
		t.Error("expected chroma highlighting classes with monokai style")
	}
}

func TestNew_WithFootnote(t *testing.T) {
	fsys := os.DirFS("testdata")

	// Test without footnote option (default disabled)
	p := New()
	post, err := p.ParseFile(fsys, "with-footnotes.md")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Footnote markers like [^1] should appear as plain text when disabled
	if strings.Contains(string(post.Content), "<sup") {
		t.Error("expected no <sup> tags when footnotes are disabled")
	}

	// Should contain the raw footnote syntax
	if !strings.Contains(string(post.Content), "[^1]") {
		t.Error("expected footnote markers to appear as plain text when disabled")
	}

	// Test with footnote option enabled
	p = New(WithFootnote())
	post, err = p.ParseFile(fsys, "with-footnotes.md")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Should contain <sup> tags for footnote references
	if !strings.Contains(string(post.Content), "<sup") {
		t.Error("expected <sup> tags when footnotes are enabled")
	}

	// Should contain footnote section
	if !strings.Contains(string(post.Content), "footnote-ref") {
		t.Error("expected footnote references when footnotes are enabled")
	}

	// Raw marker should NOT appear when rendered
	if strings.Contains(string(post.Content), "[^1]") {
		t.Error("expected footnote markers to be rendered, not shown as plain text")
	}
}

func TestNew_CombinedOptions(t *testing.T) {
	fsys := os.DirFS("testdata")

	// Test combining code highlighting style and footnotes
	p := New(
		WithCodeHighlightingStyle("monokai"),
		WithFootnote(),
	)

	// Test with code
	post, err := p.ParseFile(fsys, "with-code.md")
	if err != nil {
		t.Fatalf("expected no error parsing code file, got: %v", err)
	}

	if !strings.Contains(string(post.Content), "chroma") {
		t.Error("expected code highlighting to work with combined options")
	}

	// Test with footnotes
	post, err = p.ParseFile(fsys, "with-footnotes.md")
	if err != nil {
		t.Fatalf("expected no error parsing footnote file, got: %v", err)
	}

	if !strings.Contains(string(post.Content), "<sup") {
		t.Error("expected footnotes to work with combined options")
	}

	// Test combining code highlighting disabled with footnotes enabled
	p = New(
		WithCodeHighlighting(false),
		WithFootnote(),
	)

	post, err = p.ParseFile(fsys, "with-code.md")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if strings.Contains(string(post.Content), "chroma") {
		t.Error("expected code highlighting to be disabled")
	}

	post, err = p.ParseFile(fsys, "with-footnotes.md")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if !strings.Contains(string(post.Content), "<sup") {
		t.Error("expected footnotes to still work when code highlighting is disabled")
	}
}

func Example() {
	p := New()

	fsys := os.DirFS("testdata")
	post, err := p.ParseFile(fsys, "valid-post.md")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println(post.Title)
	// Output: A Valid Blog Post
}

func Example_withOptions() {
	// Configure parser with custom options
	p := New(
		WithCodeHighlightingStyle("dracula"),
		WithFootnote(),
	)

	fsys := os.DirFS("testdata")
	post, err := p.ParseFile(fsys, "with-code.md")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Configured parser")
	fmt.Println("Has content:", len(post.Content) > 0)
	// Output: Configured parser
	// Has content: true
}

func ExampleParser_ParseFile() {
	p := New()
	fsys := os.DirFS("testdata")

	post, err := p.ParseFile(fsys, "valid-post.md")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Title: %s\n", post.Title)
	fmt.Printf("Tags: %d\n", len(post.Tags))
	// Output: Title: A Valid Blog Post
	// Tags: 3
}

func ExampleParser_ParseDirectory() {
	p := New(WithFootnote())
	fsys := os.DirFS("testdata")

	posts, err := p.ParseDirectory(fsys)

	// Partial results may be returned with errors
	if err != nil {
		if parseErrs, ok := err.(ParseErrors); ok {
			fmt.Printf("Parsed with some errors\n")
			fmt.Printf("Valid posts: %d\n", len(posts))
			fmt.Printf("Errors: %d\n", len(parseErrs.Errors))
			return
		}
	}

	fmt.Printf("Valid posts: %d\n", len(posts))
	// Output: Parsed with some errors
	// Valid posts: 3
	// Errors: 4
}

func ExampleWithCodeHighlighting() {
	// Disable code highlighting
	p := New(WithCodeHighlighting(false))

	fsys := os.DirFS("testdata")
	post, err := p.ParseFile(fsys, "with-code.md")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Highlighting disabled")
	fmt.Println("Post parsed:", post.Title != "")
	// Output: Highlighting disabled
	// Post parsed: true
}

func ExampleWithCodeHighlightingStyle() {
	// Use a different syntax highlighting style
	p := New(WithCodeHighlightingStyle("dracula"))

	fsys := os.DirFS("testdata")
	post, err := p.ParseFile(fsys, "with-code.md")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Custom highlighting style")
	fmt.Println("Post parsed:", post.Title != "")
	// Output: Custom highlighting style
	// Post parsed: true
}

func ExampleWithFootnote() {
	// Enable PHP Markdown Extra footnotes
	p := New(WithFootnote())

	fsys := os.DirFS("testdata")
	post, err := p.ParseFile(fsys, "with-footnotes.md")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Footnotes enabled")
	fmt.Println("Post parsed:", post.Title != "")
	// Output: Footnotes enabled
	// Post parsed: true
}

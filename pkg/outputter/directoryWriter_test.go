package outputter

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/harrydayexe/GoBlog/v2/pkg/config"
	"github.com/harrydayexe/GoBlog/v2/pkg/generator"
)

// TestNewDirectoryWriter tests basic construction with and without options.
func TestNewDirectoryWriter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		outputDir   string
		opts        []config.Option
		wantRawOut  bool
	}{
		{
			name:       "without options",
			outputDir:  "/tmp/blog",
			opts:       nil,
			wantRawOut: false,
		},
		{
			name:       "with RawOutput option",
			outputDir:  "/tmp/blog-raw",
			opts:       []config.Option{config.WithRawOutput()},
			wantRawOut: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			writer := NewDirectoryWriter(tt.outputDir, tt.opts...)

			if writer.outputDir != tt.outputDir {
				t.Errorf("outputDir = %q, want %q", writer.outputDir, tt.outputDir)
			}
			if writer.RawOutput.RawOutput != tt.wantRawOut {
				t.Errorf("RawOutput = %v, want %v", writer.RawOutput.RawOutput, tt.wantRawOut)
			}
		})
	}
}

// TestDirectoryWriter_HandleGeneratedBlog_Success tests complete blog write
// and verifies all files are created correctly.
func TestDirectoryWriter_HandleGeneratedBlog_Success(t *testing.T) {
	t.Parallel()

	outputDir := t.TempDir()
	writer := NewDirectoryWriter(outputDir)

	blog := &generator.GeneratedBlog{
		Posts: map[string][]byte{
			"post-1": []byte("<h1>Post 1</h1>"),
			"post-2": []byte("<h1>Post 2</h1>"),
		},
		Index: []byte("<h1>Blog Index</h1>"),
		Tags: map[string][]byte{
			"golang": []byte("<h1>Golang Posts</h1>"),
			"web":    []byte("<h1>Web Posts</h1>"),
		},
	}

	err := writer.HandleGeneratedBlog(context.Background(), blog)
	if err != nil {
		t.Fatalf("HandleGeneratedBlog failed: %v", err)
	}

	// Verify index.html exists
	indexPath := filepath.Join(outputDir, "index.html")
	indexContent, err := os.ReadFile(indexPath)
	if err != nil {
		t.Errorf("Failed to read index.html: %v", err)
	}
	if string(indexContent) != string(blog.Index) {
		t.Errorf("index.html content = %q, want %q", indexContent, blog.Index)
	}

	// Verify post files exist
	for slug, expectedContent := range blog.Posts {
		postPath := filepath.Join(outputDir, "posts", slug+".html")
		content, err := os.ReadFile(postPath)
		if err != nil {
			t.Errorf("Failed to read %s: %v", postPath, err)
			continue
		}
		if string(content) != string(expectedContent) {
			t.Errorf("%s content = %q, want %q", postPath, content, expectedContent)
		}
	}

	// Verify tag files exist in tags subdirectory
	for tag, expectedContent := range blog.Tags {
		tagPath := filepath.Join(outputDir, "tags", tag+".html")
		content, err := os.ReadFile(tagPath)
		if err != nil {
			t.Errorf("Failed to read %s: %v", tagPath, err)
			continue
		}
		if string(content) != string(expectedContent) {
			t.Errorf("%s content = %q, want %q", tagPath, content, expectedContent)
		}
	}
}

// TestDirectoryWriter_HandleGeneratedBlog_RawOutput verifies that no tags
// directory is created when RawOutput is enabled.
func TestDirectoryWriter_HandleGeneratedBlog_RawOutput(t *testing.T) {
	t.Parallel()

	outputDir := t.TempDir()
	writer := NewDirectoryWriter(outputDir, config.WithRawOutput())

	blog := &generator.GeneratedBlog{
		Posts: map[string][]byte{
			"post-1": []byte("<h1>Post 1</h1>"),
		},
		Index: []byte("<h1>Blog Index</h1>"),
		Tags: map[string][]byte{
			"golang": []byte("<h1>Golang Posts</h1>"),
		},
	}

	err := writer.HandleGeneratedBlog(context.Background(), blog)
	if err != nil {
		t.Fatalf("HandleGeneratedBlog failed: %v", err)
	}

	// Verify tags directory does NOT exist
	tagsDir := filepath.Join(outputDir, "tags")
	if _, err := os.Stat(tagsDir); !os.IsNotExist(err) {
		t.Errorf("tags directory should not exist with RawOutput=true, but it does")
	}

	// Verify posts and index still exist
	if _, err := os.Stat(filepath.Join(outputDir, "index.html")); err != nil {
		t.Errorf("index.html should exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outputDir, "posts", "post-1.html")); err != nil {
		t.Errorf("post-1.html should exist: %v", err)
	}
}

// TestDirectoryWriter_HandleGeneratedBlog_EmptyBlog tests handling of empty input.
func TestDirectoryWriter_HandleGeneratedBlog_EmptyBlog(t *testing.T) {
	t.Parallel()

	outputDir := t.TempDir()
	writer := NewDirectoryWriter(outputDir)

	blog := generator.NewEmptyGeneratedBlog()
	blog.Index = []byte("<h1>Empty Blog</h1>")

	err := writer.HandleGeneratedBlog(context.Background(), blog)
	if err != nil {
		t.Fatalf("HandleGeneratedBlog failed with empty blog: %v", err)
	}

	// Verify index.html was created
	indexPath := filepath.Join(outputDir, "index.html")
	if _, err := os.Stat(indexPath); err != nil {
		t.Errorf("index.html should exist: %v", err)
	}

	// Verify output directory was created
	if _, err := os.Stat(outputDir); err != nil {
		t.Errorf("output directory should exist: %v", err)
	}
}

// TestDirectoryWriter_CreatesIndexFile specifically tests that index.html
// is created in the correct location (validates the bug fix on line 47).
func TestDirectoryWriter_CreatesIndexFile(t *testing.T) {
	t.Parallel()

	outputDir := t.TempDir()
	writer := NewDirectoryWriter(outputDir)

	indexContent := []byte("<h1>Test Index</h1>")
	blog := &generator.GeneratedBlog{
		Posts: make(map[string][]byte),
		Index: indexContent,
		Tags:  make(map[string][]byte),
	}

	err := writer.HandleGeneratedBlog(context.Background(), blog)
	if err != nil {
		t.Fatalf("HandleGeneratedBlog failed: %v", err)
	}

	// Verify index.html exists at the correct path
	indexPath := filepath.Join(outputDir, "index.html")
	content, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("index.html not found at expected location %s: %v", indexPath, err)
	}

	if string(content) != string(indexContent) {
		t.Errorf("index.html content = %q, want %q", content, indexContent)
	}

	// Verify we didn't accidentally create a file at the directory path itself
	// (this would have been the bug behavior)
	outputDirInfo, err := os.Stat(outputDir)
	if err != nil {
		t.Fatalf("Failed to stat output directory: %v", err)
	}
	if !outputDirInfo.IsDir() {
		t.Errorf("outputDir %s should be a directory, not a file", outputDir)
	}
}

// TestDirectoryWriter_WritesPostFiles verifies all post files are created
// with the correct .html extension.
func TestDirectoryWriter_WritesPostFiles(t *testing.T) {
	t.Parallel()

	outputDir := t.TempDir()
	writer := NewDirectoryWriter(outputDir)

	posts := map[string][]byte{
		"first-post":       []byte("<h1>First</h1>"),
		"second-post":      []byte("<h1>Second</h1>"),
		"post-with-dashes": []byte("<h1>Dashes</h1>"),
	}

	blog := &generator.GeneratedBlog{
		Posts: posts,
		Index: []byte("<h1>Index</h1>"),
		Tags:  make(map[string][]byte),
	}

	err := writer.HandleGeneratedBlog(context.Background(), blog)
	if err != nil {
		t.Fatalf("HandleGeneratedBlog failed: %v", err)
	}

	// Verify each post file exists with .html extension
	for slug := range posts {
		expectedPath := filepath.Join(outputDir, "posts", slug+".html")
		if _, err := os.Stat(expectedPath); err != nil {
			t.Errorf("Post file %s should exist: %v", expectedPath, err)
		}
	}
}

// TestDirectoryWriter_WritesTagFiles verifies tags subdirectory and files
// are created correctly.
func TestDirectoryWriter_WritesTagFiles(t *testing.T) {
	t.Parallel()

	outputDir := t.TempDir()
	writer := NewDirectoryWriter(outputDir)

	tags := map[string][]byte{
		"golang":     []byte("<h1>Golang</h1>"),
		"javascript": []byte("<h1>JavaScript</h1>"),
		"web-dev":    []byte("<h1>Web Development</h1>"),
	}

	blog := &generator.GeneratedBlog{
		Posts: make(map[string][]byte),
		Index: []byte("<h1>Index</h1>"),
		Tags:  tags,
	}

	err := writer.HandleGeneratedBlog(context.Background(), blog)
	if err != nil {
		t.Fatalf("HandleGeneratedBlog failed: %v", err)
	}

	// Verify tags directory exists
	tagsDir := filepath.Join(outputDir, "tags")
	tagsDirInfo, err := os.Stat(tagsDir)
	if err != nil {
		t.Fatalf("tags directory should exist: %v", err)
	}
	if !tagsDirInfo.IsDir() {
		t.Errorf("tags path should be a directory")
	}

	// Verify each tag file exists
	for tag := range tags {
		expectedPath := filepath.Join(tagsDir, tag+".html")
		if _, err := os.Stat(expectedPath); err != nil {
			t.Errorf("Tag file %s should exist: %v", expectedPath, err)
		}
	}
}

// TestDirectoryWriter_HandleGeneratedBlog_InvalidPath tests behavior with
// invalid output paths.
func TestDirectoryWriter_HandleGeneratedBlog_InvalidPath(t *testing.T) {
	// Note: This test is platform-dependent and may behave differently on
	// different operating systems. On Unix-like systems, most paths are valid.
	// We test with a path that's likely to fail.
	t.Parallel()

	// Use a path that includes a null byte, which is invalid on most systems
	invalidPath := "/tmp/test\x00invalid"
	writer := NewDirectoryWriter(invalidPath)

	blog := &generator.GeneratedBlog{
		Posts: map[string][]byte{"test": []byte("content")},
		Index: []byte("index"),
		Tags:  make(map[string][]byte),
	}

	err := writer.HandleGeneratedBlog(context.Background(), blog)
	if err == nil {
		t.Error("Expected error with invalid path containing null byte, got nil")
	}
}

// TestDirectoryWriter_HandleGeneratedBlog_NoPermissions tests behavior when
// writing to a read-only directory.
func TestDirectoryWriter_HandleGeneratedBlog_NoPermissions(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Skipping permission test when running as root")
	}

	t.Parallel()

	outputDir := t.TempDir()

	// Create a subdirectory and make it read-only
	readOnlyDir := filepath.Join(outputDir, "readonly")
	if err := os.Mkdir(readOnlyDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	if err := os.Chmod(readOnlyDir, 0444); err != nil {
		t.Fatalf("Failed to make directory read-only: %v", err)
	}
	// Restore permissions for cleanup
	t.Cleanup(func() {
		os.Chmod(readOnlyDir, 0755)
	})

	writer := NewDirectoryWriter(readOnlyDir)
	blog := &generator.GeneratedBlog{
		Posts: map[string][]byte{"test": []byte("content")},
		Index: []byte("index"),
		Tags:  make(map[string][]byte),
	}

	err := writer.HandleGeneratedBlog(context.Background(), blog)
	if err == nil {
		t.Error("Expected error when writing to read-only directory, got nil")
	}
}

// TestWriteMapToFiles_ErrorHandling tests that errors from writeMapToFiles
// are properly propagated (validates the bug fix on lines 39 and 43).
func TestWriteMapToFiles_ErrorHandling(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Skipping permission test when running as root")
	}

	t.Parallel()

	// Create a directory structure where we can't write
	tempDir := t.TempDir()
	readOnlyDir := filepath.Join(tempDir, "readonly")
	if err := os.Mkdir(readOnlyDir, 0444); err != nil {
		t.Fatalf("Failed to create read-only directory: %v", err)
	}
	t.Cleanup(func() {
		os.Chmod(readOnlyDir, 0755)
	})

	// Try to write to a subdirectory of the read-only directory
	data := map[string][]byte{
		"test": []byte("content"),
	}

	err := writeMapToFiles(data, filepath.Join(readOnlyDir, "subdir"))
	if err == nil {
		t.Error("Expected error when writing to read-only location, got nil")
	}
}

// TestDirectoryWriter_OverwritesExistingFiles tests idempotency - running
// HandleGeneratedBlog multiple times should work correctly.
func TestDirectoryWriter_OverwritesExistingFiles(t *testing.T) {
	t.Parallel()

	outputDir := t.TempDir()
	writer := NewDirectoryWriter(outputDir)

	// First write
	blog1 := &generator.GeneratedBlog{
		Posts: map[string][]byte{"post": []byte("original content")},
		Index: []byte("original index"),
		Tags:  map[string][]byte{"tag": []byte("original tag")},
	}

	err := writer.HandleGeneratedBlog(context.Background(), blog1)
	if err != nil {
		t.Fatalf("First HandleGeneratedBlog failed: %v", err)
	}

	// Second write with different content
	blog2 := &generator.GeneratedBlog{
		Posts: map[string][]byte{"post": []byte("updated content")},
		Index: []byte("updated index"),
		Tags:  map[string][]byte{"tag": []byte("updated tag")},
	}

	err = writer.HandleGeneratedBlog(context.Background(), blog2)
	if err != nil {
		t.Fatalf("Second HandleGeneratedBlog failed: %v", err)
	}

	// Verify files contain updated content
	indexContent, _ := os.ReadFile(filepath.Join(outputDir, "index.html"))
	if string(indexContent) != "updated index" {
		t.Errorf("index.html not updated correctly")
	}

	postContent, _ := os.ReadFile(filepath.Join(outputDir, "posts", "post.html"))
	if string(postContent) != "updated content" {
		t.Errorf("post.html not updated correctly")
	}

	tagContent, _ := os.ReadFile(filepath.Join(outputDir, "tags", "tag.html"))
	if string(tagContent) != "updated tag" {
		t.Errorf("tag.html not updated correctly")
	}
}

// TestDirectoryWriter_HandlesSpecialCharactersInSlugs tests that slugs with
// special characters are handled correctly.
func TestDirectoryWriter_HandlesSpecialCharactersInSlugs(t *testing.T) {
	t.Parallel()

	outputDir := t.TempDir()
	writer := NewDirectoryWriter(outputDir)

	blog := &generator.GeneratedBlog{
		Posts: map[string][]byte{
			"post-with-dashes":     []byte("content1"),
			"post_with_underscores": []byte("content2"),
			"post.with.dots":       []byte("content3"),
		},
		Index: []byte("index"),
		Tags: map[string][]byte{
			"c++":      []byte("cpp tag"),
			"c#":       []byte("csharp tag"),
			"go-lang":  []byte("golang tag"),
		},
	}

	err := writer.HandleGeneratedBlog(context.Background(), blog)
	if err != nil {
		t.Fatalf("HandleGeneratedBlog failed with special characters: %v", err)
	}

	// Verify all files were created
	for slug := range blog.Posts {
		path := filepath.Join(outputDir, "posts", slug+".html")
		if _, err := os.Stat(path); err != nil {
			t.Errorf("Post file %s should exist: %v", path, err)
		}
	}

	for tag := range blog.Tags {
		path := filepath.Join(outputDir, "tags", tag+".html")
		if _, err := os.Stat(path); err != nil {
			t.Errorf("Tag file %s should exist: %v", path, err)
		}
	}
}

// TestDirectoryWriter_Integration is a full integration test using realistic
// blog content.
func TestDirectoryWriter_Integration(t *testing.T) {
	t.Parallel()

	outputDir := t.TempDir()
	writer := NewDirectoryWriter(outputDir)

	// Create a realistic blog structure
	blog := &generator.GeneratedBlog{
		Posts: map[string][]byte{
			"introduction-to-go": []byte(`
				<!DOCTYPE html>
				<html>
				<head><title>Introduction to Go</title></head>
				<body><h1>Introduction to Go</h1><p>Go is a great language...</p></body>
				</html>
			`),
			"web-development-tips": []byte(`
				<!DOCTYPE html>
				<html>
				<head><title>Web Development Tips</title></head>
				<body><h1>Web Dev Tips</h1><p>Here are some tips...</p></body>
				</html>
			`),
		},
		Index: []byte(`
			<!DOCTYPE html>
			<html>
			<head><title>My Blog</title></head>
			<body><h1>Welcome to My Blog</h1><ul><li>Post 1</li><li>Post 2</li></ul></body>
			</html>
		`),
		Tags: map[string][]byte{
			"golang": []byte(`
				<!DOCTYPE html>
				<html>
				<head><title>Posts tagged: golang</title></head>
				<body><h1>Golang Posts</h1></body>
				</html>
			`),
			"webdev": []byte(`
				<!DOCTYPE html>
				<html>
				<head><title>Posts tagged: webdev</title></head>
				<body><h1>Web Development Posts</h1></body>
				</html>
			`),
		},
	}

	err := writer.HandleGeneratedBlog(context.Background(), blog)
	if err != nil {
		t.Fatalf("Integration test failed: %v", err)
	}

	// Verify the complete structure exists
	expectedFiles := []string{
		"index.html",
		"posts/introduction-to-go.html",
		"posts/web-development-tips.html",
		"tags/golang.html",
		"tags/webdev.html",
	}

	for _, file := range expectedFiles {
		path := filepath.Join(outputDir, file)
		info, err := os.Stat(path)
		if err != nil {
			t.Errorf("Expected file %s not found: %v", file, err)
			continue
		}
		if info.IsDir() {
			t.Errorf("Expected file %s is a directory", file)
		}
		if info.Size() == 0 {
			t.Errorf("File %s is empty", file)
		}
	}

	// Verify directory structure
	tagsDir := filepath.Join(outputDir, "tags")
	if info, err := os.Stat(tagsDir); err != nil || !info.IsDir() {
		t.Errorf("tags directory should exist and be a directory")
	}
}

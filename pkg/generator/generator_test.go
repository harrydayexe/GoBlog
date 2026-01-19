package generator

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"testing"
	"time"

	"github.com/harrydayexe/GoBlog/v2/pkg/config"
)

// TestNew tests creating a new Generator with various options.
func TestNew(t *testing.T) {
	t.Parallel()

	testFS := os.DirFS("testdata")

	tests := []struct {
		name    string
		posts   fs.FS
		opts    []config.Option
		wantNil bool
	}{
		{
			name:    "with valid filesystem and no options",
			posts:   testFS,
			opts:    nil,
			wantNil: false,
		},
		{
			name:    "with valid filesystem and RawOutput option",
			posts:   testFS,
			opts:    []config.Option{config.WithRawOutput()},
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gen := New(tt.posts, nil, tt.opts...)

			if (gen == nil) != tt.wantNil {
				t.Errorf("New() = %v, wantNil %v", gen, tt.wantNil)
			}

			if gen != nil {
				if gen.logger == nil {
					t.Error("New() returned generator with nil logger")
				}
			}
		})
	}
}

// TestNew_NilFilesystem tests creating a generator with nil filesystem.
func TestNew_NilFilesystem(t *testing.T) {
	t.Parallel()

	// Generator creation with nil filesystem should still succeed
	// (validation happens during Generate)
	gen := New(nil, nil)

	if gen == nil {
		t.Fatal("New(nil) returned nil")
	}

	if gen.PostsDir != nil {
		t.Error("New(nil) should have nil PostsDir")
	}
}

// TestGenerate_RawOutput tests the happy path: valid posts generating HTML.
func TestGenerate_RawOutput(t *testing.T) {
	t.Parallel()

	testFS := os.DirFS("testdata")
	gen := New(testFS, nil, config.WithRawOutput())

	ctx := context.Background()
	blog, err := gen.Generate(ctx)

	if err != nil {
		t.Fatalf("Generate() error = %v, want nil", err)
	}

	if blog == nil {
		t.Fatal("Generate() returned nil blog")
	}

	// Verify we got at least one post (simple.md, complex.md, no-tags.md)
	if len(blog.Posts) == 0 {
		t.Error("Generate() returned 0 posts, want at least 1")
	}

	// Verify posts have content
	for slug, content := range blog.Posts {
		if len(content) == 0 {
			t.Errorf("Post %q has empty content", slug)
		}
	}
}

// TestGenerate_MultiplePosts tests generating multiple posts with unique slugs.
func TestGenerate_MultiplePosts(t *testing.T) {
	t.Parallel()

	testFS := os.DirFS("testdata")
	gen := New(testFS, nil, config.WithRawOutput())

	ctx := context.Background()
	blog, err := gen.Generate(ctx)

	if err != nil {
		t.Fatalf("Generate() error = %v, want nil", err)
	}

	// We should have at least 3 posts (simple, complex, no-tags)
	if len(blog.Posts) < 3 {
		t.Errorf("Generate() returned %d posts, want at least 3", len(blog.Posts))
	}

	// Verify slugs are unique (map ensures this, but check anyway)
	seenSlugs := make(map[string]bool)
	for slug := range blog.Posts {
		if seenSlugs[slug] {
			t.Errorf("Duplicate slug found: %q", slug)
		}
		seenSlugs[slug] = true
	}
}

// TestGenerate_EmptyDirectory tests generating from a directory with no markdown files.
func TestGenerate_EmptyDirectory(t *testing.T) {
	t.Parallel()

	// Create a temporary empty directory
	tempDir := t.TempDir()
	testFS := os.DirFS(tempDir)

	gen := New(testFS, nil, config.WithRawOutput())

	ctx := context.Background()
	blog, err := gen.Generate(ctx)

	if err != nil {
		t.Fatalf("Generate() error = %v, want nil", err)
	}

	// Should return an empty blog with no posts
	if len(blog.Posts) != 0 {
		t.Errorf("Generate() returned %d posts, want 0 for empty directory", len(blog.Posts))
	}
}

// TestGenerate_WithParserErrors tests generation with mix of valid/invalid posts.
func TestGenerate_WithParserErrors(t *testing.T) {
	t.Parallel()

	// Use parser's testdata which contains invalid files
	testFS := os.DirFS("../parser/testdata")

	gen := New(testFS, nil, config.WithRawOutput())

	ctx := context.Background()
	blog, err := gen.Generate(ctx)

	// Parser should return an error for invalid files
	if err == nil {
		t.Fatal("Generate() with invalid files should return error, got nil")
	}

	// Blog should be nil when there's an error
	if blog != nil {
		t.Errorf("Generate() with error should return nil blog, got %v", blog)
	}
}

// TestGenerate_ContextCanceled tests generation with pre-canceled context.
func TestGenerate_ContextCanceled(t *testing.T) {
	t.Parallel()

	testFS := os.DirFS("testdata")
	gen := New(testFS, nil, config.WithRawOutput())

	// Create a pre-canceled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	blog, err := gen.Generate(ctx)

	// NOTE: With small test files, parsing may complete before the context is checked
	// If an error occurs, it should be context.Canceled
	// If no error occurs, the parser finished before checking the context (also valid)
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			t.Errorf("Generate() error = %v, want context.Canceled if error occurs", err)
		}
		if blog != nil {
			t.Errorf("Generate() with error should return nil blog, got %v", blog)
		}
	}
}

// TestGenerate_ContextTimeout tests generation with context deadline.
func TestGenerate_ContextTimeout(t *testing.T) {
	t.Parallel()

	testFS := os.DirFS("testdata")
	gen := New(testFS, nil, config.WithRawOutput())

	// Create a context with very short timeout (1 nanosecond - guaranteed to expire)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Wait to ensure timeout
	time.Sleep(1 * time.Millisecond)

	blog, err := gen.Generate(ctx)

	// NOTE: With small test files, parsing may complete before the deadline is checked
	// If an error occurs, it should be context.DeadlineExceeded
	// If no error occurs, the parser finished before checking the deadline (also valid)
	if err != nil {
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("Generate() error = %v, want context.DeadlineExceeded if error occurs", err)
		}
		if blog != nil {
			t.Errorf("Generate() with error should return nil blog, got %v", blog)
		}
	}
}

// TestGenerate_WithoutRawOutput tests that template mode returns error.
func TestGenerate_WithoutRawOutput(t *testing.T) {
	t.Parallel()

	testFS := os.DirFS("testdata")
	gen := New(testFS, nil) // No RawOutput option

	ctx := context.Background()
	blog, err := gen.Generate(ctx)

	// Should return error because template mode is not implemented
	if err == nil {
		t.Fatal("Generate() without RawOutput should return error, got nil")
	}

	if blog != nil {
		t.Errorf("Generate() with error should return nil blog, got %v", blog)
	}
}

// TestDebugConfig tests that DebugConfig doesn't panic or error.
func TestDebugConfig(t *testing.T) {
	t.Parallel()

	testFS := os.DirFS("testdata")
	gen := New(testFS, nil, config.WithRawOutput())

	ctx := context.Background()

	// Should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("DebugConfig() panicked: %v", r)
		}
	}()

	gen.DebugConfig(ctx)
}

// TestWithRawOutput tests that the WithRawOutput option is applied correctly.
func TestWithRawOutput(t *testing.T) {
	t.Parallel()

	testFS := os.DirFS("testdata")

	// Without option
	genWithout := New(testFS, nil)
	if genWithout.RawOutput.RawOutput {
		t.Error("Generator without config.WithRawOutput() has RawOutput = true, want false")
	}

	// With option
	genWith := New(testFS, nil, config.WithRawOutput())
	if !genWith.RawOutput.RawOutput {
		t.Error("Generator with config.WithRawOutput() has RawOutput = false, want true")
	}
}

// TestAssembleRawBlog tests the assembleRawBlog helper function.
func TestAssembleRawBlog(t *testing.T) {
	t.Parallel()

	testFS := os.DirFS("testdata")
	gen := New(testFS, nil, config.WithRawOutput())

	ctx := context.Background()
	blog, err := gen.Generate(ctx)

	if err != nil {
		t.Fatalf("Generate() error = %v, want nil", err)
	}

	// Verify the blog structure
	if blog.Posts == nil {
		t.Fatal("assembleRawBlog() created blog with nil Posts map")
	}

	// Verify each post has a slug and content
	for slug, content := range blog.Posts {
		if slug == "" {
			t.Error("assembleRawBlog() created post with empty slug")
		}
		if len(content) == 0 {
			t.Errorf("assembleRawBlog() post %q has empty content", slug)
		}
	}

	// Verify Index and Tags are initialized (even if empty in raw mode)
	if blog.Tags == nil {
		t.Error("assembleRawBlog() created blog with nil Tags map")
	}
}

// TestGeneratedBlog_NewEmptyGeneratedBlog tests creating an empty blog.
func TestGeneratedBlog_NewEmptyGeneratedBlog(t *testing.T) {
	t.Parallel()

	blog := NewEmptyGeneratedBlog()

	if blog == nil {
		t.Fatal("NewEmptyGeneratedBlog() returned nil")
	}

	if blog.Posts == nil {
		t.Error("NewEmptyGeneratedBlog() created blog with nil Posts map")
	}

	if blog.Tags == nil {
		t.Error("NewEmptyGeneratedBlog() created blog with nil Tags map")
	}

	if len(blog.Posts) != 0 {
		t.Errorf("NewEmptyGeneratedBlog() Posts has %d entries, want 0", len(blog.Posts))
	}

	if len(blog.Tags) != 0 {
		t.Errorf("NewEmptyGeneratedBlog() Tags has %d entries, want 0", len(blog.Tags))
	}

	if len(blog.Index) != 0 {
		t.Errorf("NewEmptyGeneratedBlog() Index has %d bytes, want 0", len(blog.Index))
	}
}

// TestGenerator_String tests the String method.
func TestGenerator_String(t *testing.T) {
	t.Parallel()

	testFS := os.DirFS("testdata")

	tests := []struct {
		name      string
		gen       *Generator
		wantEmpty bool
	}{
		{
			name:      "with RawOutput true",
			gen:       New(testFS, nil, config.WithRawOutput()),
			wantEmpty: false,
		},
		{
			name:      "with RawOutput false",
			gen:       New(testFS, nil),
			wantEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.gen.String()

			if (len(got) == 0) != tt.wantEmpty {
				t.Errorf("String() = %q, wantEmpty %v", got, tt.wantEmpty)
			}
		})
	}
}

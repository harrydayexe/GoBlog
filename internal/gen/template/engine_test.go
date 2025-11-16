package template

import (
	"bytes"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/harrydayexe/GoBlog/internal/gen/config"
	"github.com/harrydayexe/GoBlog/internal/gen/log"
	"github.com/harrydayexe/GoBlog/pkg/models"
)

// createTestTemplates creates minimal valid templates for testing
func createTestTemplates(t *testing.T, dir string) {
	t.Helper()

	postTemplate := `<!DOCTYPE html>
<html>
<head><title>{{.Post.Title}}</title></head>
<body>
<h1>{{.Post.Title}}</h1>
<p>{{.Post.Description}}</p>
<div>{{.Post.HTMLContent}}</div>
</body>
</html>`

	indexTemplate := `<!DOCTYPE html>
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
{{if .HasNext}}<a href="page{{add .Page 1}}.html">Next</a>{{end}}
{{if .HasPrev}}<a href="page{{sub .Page 1}}.html">Prev</a>{{end}}
</body>
</html>`

	tagTemplate := `<!DOCTYPE html>
<html>
<head><title>Tag: {{.Tag}}</title></head>
<body>
<h1>Posts tagged with {{.Tag}}</h1>
{{range .Posts}}
<article>
<h2>{{.Title}}</h2>
</article>
{{end}}
</body>
</html>`

	templates := map[string]string{
		"post.html":  postTemplate,
		"index.html": indexTemplate,
		"tag.html":   tagTemplate,
	}

	for name, content := range templates {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create template %s: %v", name, err)
		}
	}
}

// TestNew tests template engine creation
func TestNew(t *testing.T) {
	tmpDir := t.TempDir()
	createTestTemplates(t, tmpDir)

	tests := []struct {
		name       string
		cfg        config.Config
		logger     log.Logger
		expectErr  bool
		errText    string
	}{
		{
			name: "valid config with templates",
			cfg: config.Config{
				TemplateDir: tmpDir,
				Site: config.SiteMetadata{
					Title: "Test Blog",
				},
			},
			logger:    nil,
			expectErr: false,
		},
		{
			name: "non-existent template directory",
			cfg: config.Config{
				TemplateDir: "/non/existent/dir",
			},
			logger:    nil,
			expectErr: true,
			errText:   "does not exist",
		},
		{
			name: "with custom logger",
			cfg: config.Config{
				TemplateDir: tmpDir,
			},
			logger:    log.NewTestLogger("TEST", false, &bytes.Buffer{}, &bytes.Buffer{}),
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine, err := New(tt.cfg, tt.logger)

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
				if engine == nil {
					t.Fatal("expected engine to be non-nil")
				}
				if engine.postTemplate == nil {
					t.Error("expected postTemplate to be loaded")
				}
				if engine.indexTemplate == nil {
					t.Error("expected indexTemplate to be loaded")
				}
				if engine.tagTemplate == nil {
					t.Error("expected tagTemplate to be loaded")
				}
			}
		})
	}
}

// TestEngine_RenderPost tests rendering a single post
func TestEngine_RenderPost(t *testing.T) {
	tmpDir := t.TempDir()
	createTestTemplates(t, tmpDir)

	cfg := config.Config{
		TemplateDir: tmpDir,
		Site: config.SiteMetadata{
			Title:       "Test Blog",
			Description: "A test blog",
		},
		BlogPath: "/blog",
	}

	logger := log.NewTestLogger("TEST", false, &bytes.Buffer{}, &bytes.Buffer{})
	engine, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	post := &models.Post{
		Title:       "Test Post",
		Description: "A test post",
		Date:        time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC),
		HTMLContent: template.HTML("<p>Hello World</p>"),
		Slug:        "test-post",
	}

	html, err := engine.RenderPost(post)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify rendered HTML contains expected elements
	checks := []string{
		"Test Post",
		"A test post",
		"<p>Hello World</p>",
	}

	for _, check := range checks {
		if !strings.Contains(html, check) {
			t.Errorf("expected rendered HTML to contain %q", check)
		}
	}
}

// TestEngine_RenderIndex tests rendering an index page
func TestEngine_RenderIndex(t *testing.T) {
	tmpDir := t.TempDir()
	createTestTemplates(t, tmpDir)

	cfg := config.Config{
		TemplateDir: tmpDir,
		Site: config.SiteMetadata{
			Title: "Test Blog",
		},
		BlogPath: "/blog",
	}

	logger := log.NewTestLogger("TEST", false, &bytes.Buffer{}, &bytes.Buffer{})
	engine, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	posts := models.PostList{
		{
			Title:       "Post 1",
			Description: "First post",
			Date:        time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			Title:       "Post 2",
			Description: "Second post",
			Date:        time.Date(2024, 3, 14, 0, 0, 0, 0, time.UTC),
		},
	}

	tests := []struct {
		name       string
		page       int
		totalPages int
		checks     []string
	}{
		{
			name:       "first page",
			page:       1,
			totalPages: 2,
			checks: []string{
				"Test Blog",
				"Post 1",
				"Post 2",
				"Next",
			},
		},
		{
			name:       "last page",
			page:       2,
			totalPages: 2,
			checks: []string{
				"Test Blog",
				"Prev",
			},
		},
		{
			name:       "middle page",
			page:       2,
			totalPages: 3,
			checks: []string{
				"Test Blog",
				"Next",
				"Prev",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html, err := engine.RenderIndex(posts, tt.page, tt.totalPages)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for _, check := range tt.checks {
				if !strings.Contains(html, check) {
					t.Errorf("expected rendered HTML to contain %q", check)
				}
			}
		})
	}
}

// TestEngine_RenderTag tests rendering a tag page
func TestEngine_RenderTag(t *testing.T) {
	tmpDir := t.TempDir()
	createTestTemplates(t, tmpDir)

	cfg := config.Config{
		TemplateDir: tmpDir,
		Site: config.SiteMetadata{
			Title: "Test Blog",
		},
		BlogPath: "/blog",
	}

	logger := log.NewTestLogger("TEST", false, &bytes.Buffer{}, &bytes.Buffer{})
	engine, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	posts := models.PostList{
		{
			Title:       "Go Post",
			Description: "About Go",
			Date:        time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC),
			Tags:        []string{"go"},
		},
	}

	allPosts := models.PostList{
		posts[0],
		{
			Title:       "Python Post",
			Description: "About Python",
			Date:        time.Date(2024, 3, 14, 0, 0, 0, 0, time.UTC),
			Tags:        []string{"python"},
		},
	}

	html, err := engine.RenderTag("go", posts, allPosts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checks := []string{
		"Posts tagged with go",
		"Go Post",
	}

	for _, check := range checks {
		if !strings.Contains(html, check) {
			t.Errorf("expected rendered HTML to contain %q", check)
		}
	}

	// Should not contain posts from other tags
	if strings.Contains(html, "Python Post") {
		t.Error("rendered HTML should not contain posts from other tags")
	}
}

// TestWriteHTML tests writing HTML to a file
func TestWriteHTML(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		path      string
		content   string
		expectErr bool
	}{
		{
			name:      "write to new file",
			path:      filepath.Join(tmpDir, "test.html"),
			content:   "<html><body>Test</body></html>",
			expectErr: false,
		},
		{
			name:      "write to nested directory",
			path:      filepath.Join(tmpDir, "nested", "dir", "test.html"),
			content:   "<html><body>Nested</body></html>",
			expectErr: false,
		},
		{
			name:      "overwrite existing file",
			path:      filepath.Join(tmpDir, "existing.html"),
			content:   "<html><body>New Content</body></html>",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For overwrite test, create the file first
			if tt.name == "overwrite existing file" {
				if err := os.WriteFile(tt.path, []byte("Old Content"), 0644); err != nil {
					t.Fatalf("failed to create existing file: %v", err)
				}
			}

			err := WriteHTML(tt.path, tt.content)

			if tt.expectErr {
				if err == nil {
					t.Error("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				// Verify file was written correctly
				content, readErr := os.ReadFile(tt.path)
				if readErr != nil {
					t.Fatalf("failed to read written file: %v", readErr)
				}

				if string(content) != tt.content {
					t.Errorf("file content mismatch: got %q, want %q", string(content), tt.content)
				}
			}
		})
	}
}

// TestEngine_MissingTemplates tests error handling for missing templates
func TestEngine_MissingTemplates(t *testing.T) {
	tmpDir := t.TempDir()

	// Only create post template, not index or tag
	postTemplate := `<html><body>{{.Post.Title}}</body></html>`
	postPath := filepath.Join(tmpDir, "post.html")
	if err := os.WriteFile(postPath, []byte(postTemplate), 0644); err != nil {
		t.Fatalf("failed to create post template: %v", err)
	}

	cfg := config.Config{
		TemplateDir: tmpDir,
	}

	logger := log.NewTestLogger("TEST", false, &bytes.Buffer{}, &bytes.Buffer{})
	_, err := New(cfg, logger)

	if err == nil {
		t.Error("expected error for missing templates")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

// TestTemplateFuncs tests custom template functions
func TestTemplateFuncs(t *testing.T) {
	tests := []struct {
		name     string
		funcName string
		args     []int
		expected int
	}{
		{
			name:     "add function",
			funcName: "add",
			args:     []int{5, 3},
			expected: 8,
		},
		{
			name:     "sub function",
			funcName: "sub",
			args:     []int{10, 3},
			expected: 7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn, ok := templateFuncs[tt.funcName]
			if !ok {
				t.Fatalf("function %s not found in templateFuncs", tt.funcName)
			}

			// Call the function
			result := fn.(func(int, int) int)(tt.args[0], tt.args[1])

			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

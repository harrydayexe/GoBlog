package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestConfig_Validate tests the configuration validation and defaults
func TestConfig_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		input     Config
		expectErr bool
		errText   string
		validate  func(*testing.T, Config)
	}{
		{
			name: "valid config with all fields",
			input: Config{
				Verbose:      true,
				InputFolder:  "./posts",
				OutputFolder: "./site",
				StaticFolder: "./static",
				TemplateDir:  "./templates",
				Site: SiteMetadata{
					Title:       "My Blog",
					Description: "A great blog",
					Author:      "John Doe",
					URL:         "https://example.com",
				},
				PostsPerPage: 10,
				BlogPath:     "/blog",
			},
			expectErr: false,
			validate: func(t *testing.T, cfg Config) {
				if cfg.InputFolder != "./posts" {
					t.Errorf("expected InputFolder ./posts, got %s", cfg.InputFolder)
				}
				if cfg.OutputFolder != "./site" {
					t.Errorf("expected OutputFolder ./site, got %s", cfg.OutputFolder)
				}
			},
		},
		{
			name:      "empty config gets defaults",
			input:     Config{},
			expectErr: false,
			validate: func(t *testing.T, cfg Config) {
				if cfg.InputFolder != "./posts" {
					t.Errorf("expected default InputFolder ./posts, got %s", cfg.InputFolder)
				}
				if cfg.OutputFolder != "./site" {
					t.Errorf("expected default OutputFolder ./site, got %s", cfg.OutputFolder)
				}
				if cfg.StaticFolder != "./static" {
					t.Errorf("expected default StaticFolder ./static, got %s", cfg.StaticFolder)
				}
				if cfg.TemplateDir != "./templates/defaults" {
					t.Errorf("expected default TemplateDir ./templates/defaults, got %s", cfg.TemplateDir)
				}
				if cfg.Site.Title != "My Blog" {
					t.Errorf("expected default Site.Title 'My Blog', got %s", cfg.Site.Title)
				}
				if cfg.Site.Description != "A blog about interesting things" {
					t.Errorf("expected default description, got %s", cfg.Site.Description)
				}
				if cfg.Site.Author != "Anonymous" {
					t.Errorf("expected default Author 'Anonymous', got %s", cfg.Site.Author)
				}
				if cfg.PostsPerPage != 10 {
					t.Errorf("expected default PostsPerPage 10, got %d", cfg.PostsPerPage)
				}
				if cfg.BlogPath != "/blog" {
					t.Errorf("expected default BlogPath '/blog', got %s", cfg.BlogPath)
				}
			},
		},
		{
			name: "same input and output folders",
			input: Config{
				InputFolder:  "./same",
				OutputFolder: "./same",
			},
			expectErr: true,
			errText:   "cannot be the same",
		},
		{
			name: "posts per page capped at 100",
			input: Config{
				InputFolder:  "./posts",
				OutputFolder: "./site",
				PostsPerPage: 500,
			},
			expectErr: false,
			validate: func(t *testing.T, cfg Config) {
				if cfg.PostsPerPage != 100 {
					t.Errorf("expected PostsPerPage capped at 100, got %d", cfg.PostsPerPage)
				}
			},
		},
		{
			name: "zero posts per page defaults to 10",
			input: Config{
				InputFolder:  "./posts",
				OutputFolder: "./site",
				PostsPerPage: 0,
			},
			expectErr: false,
			validate: func(t *testing.T, cfg Config) {
				if cfg.PostsPerPage != 10 {
					t.Errorf("expected PostsPerPage defaulted to 10, got %d", cfg.PostsPerPage)
				}
			},
		},
		{
			name: "negative posts per page defaults to 10",
			input: Config{
				InputFolder:  "./posts",
				OutputFolder: "./site",
				PostsPerPage: -5,
			},
			expectErr: false,
			validate: func(t *testing.T, cfg Config) {
				if cfg.PostsPerPage != 10 {
					t.Errorf("expected PostsPerPage defaulted to 10, got %d", cfg.PostsPerPage)
				}
			},
		},
		{
			name: "partial site metadata gets defaults",
			input: Config{
				InputFolder:  "./posts",
				OutputFolder: "./site",
				Site: SiteMetadata{
					Title: "Custom Title",
				},
			},
			expectErr: false,
			validate: func(t *testing.T, cfg Config) {
				if cfg.Site.Title != "Custom Title" {
					t.Errorf("expected Site.Title 'Custom Title', got %s", cfg.Site.Title)
				}
				if cfg.Site.Description != "A blog about interesting things" {
					t.Errorf("expected default description, got %s", cfg.Site.Description)
				}
				if cfg.Site.Author != "Anonymous" {
					t.Errorf("expected default Author, got %s", cfg.Site.Author)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := tt.input
			err := cfg.Validate()

			if tt.expectErr {
				if err == nil {
					t.Error("expected error but got nil")
				} else if tt.errText != "" && !contains(err.Error(), tt.errText) {
					t.Errorf("expected error to contain %q, got %q", tt.errText, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if tt.validate != nil {
					tt.validate(t, cfg)
				}
			}
		})
	}
}

// TestParseConfigFromBytes tests parsing configuration from bytes
func TestParseConfigFromBytes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		yaml      string
		expectErr bool
		validate  func(*testing.T, Config)
	}{
		{
			name: "valid yaml",
			yaml: `
verbose: true
input_folder: ./my-posts
output_folder: ./my-site
site:
  title: Test Blog
  author: Test Author
`,
			expectErr: false,
			validate: func(t *testing.T, cfg Config) {
				if !cfg.Verbose {
					t.Error("expected Verbose to be true")
				}
				if cfg.InputFolder != "./my-posts" {
					t.Errorf("expected InputFolder ./my-posts, got %s", cfg.InputFolder)
				}
				if cfg.OutputFolder != "./my-site" {
					t.Errorf("expected OutputFolder ./my-site, got %s", cfg.OutputFolder)
				}
				if cfg.Site.Title != "Test Blog" {
					t.Errorf("expected Site.Title 'Test Blog', got %s", cfg.Site.Title)
				}
				if cfg.Site.Author != "Test Author" {
					t.Errorf("expected Site.Author 'Test Author', got %s", cfg.Site.Author)
				}
			},
		},
		{
			name: "minimal valid yaml",
			yaml: `
input_folder: ./posts
output_folder: ./output
`,
			expectErr: false,
			validate: func(t *testing.T, cfg Config) {
				if cfg.InputFolder != "./posts" {
					t.Errorf("expected InputFolder ./posts, got %s", cfg.InputFolder)
				}
				if cfg.OutputFolder != "./output" {
					t.Errorf("expected OutputFolder ./output, got %s", cfg.OutputFolder)
				}
				// Check defaults are applied
				if cfg.StaticFolder != "./static" {
					t.Errorf("expected default StaticFolder, got %s", cfg.StaticFolder)
				}
			},
		},
		{
			name:      "invalid yaml",
			yaml:      `invalid: [yaml: syntax`,
			expectErr: true,
		},
		{
			name: "yaml with validation error",
			yaml: `
input_folder: ./same
output_folder: ./same
`,
			expectErr: true,
		},
		{
			name: "posts_per_page configuration",
			yaml: `
input_folder: ./posts
output_folder: ./site
posts_per_page: 25
`,
			expectErr: false,
			validate: func(t *testing.T, cfg Config) {
				if cfg.PostsPerPage != 25 {
					t.Errorf("expected PostsPerPage 25, got %d", cfg.PostsPerPage)
				}
			},
		},
		{
			name: "blog_path configuration",
			yaml: `
input_folder: ./posts
output_folder: ./site
blog_path: /articles
`,
			expectErr: false,
			validate: func(t *testing.T, cfg Config) {
				if cfg.BlogPath != "/articles" {
					t.Errorf("expected BlogPath /articles, got %s", cfg.BlogPath)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg, err := ParseConfigFromBytes([]byte(tt.yaml))

			if tt.expectErr {
				if err == nil {
					t.Error("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if tt.validate != nil {
					tt.validate(t, cfg)
				}
			}
		})
	}
}

// TestParseConfig tests parsing configuration from a file
func TestParseConfig(t *testing.T) {
	t.Parallel()
	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		filename  string
		content   string
		expectErr bool
		errText   string
		validate  func(*testing.T, Config)
	}{
		{
			name:     "valid config file",
			filename: "valid.yaml",
			content: `
verbose: true
input_folder: ./posts
output_folder: ./site
site:
  title: Test Blog
  description: A test blog
  author: Test Author
  url: https://test.com
posts_per_page: 15
blog_path: /blog
`,
			expectErr: false,
			validate: func(t *testing.T, cfg Config) {
				if !cfg.Verbose {
					t.Error("expected Verbose to be true")
				}
				if cfg.Site.Title != "Test Blog" {
					t.Errorf("expected title 'Test Blog', got %s", cfg.Site.Title)
				}
				if cfg.PostsPerPage != 15 {
					t.Errorf("expected 15 posts per page, got %d", cfg.PostsPerPage)
				}
			},
		},
		{
			name:      "non-existent file",
			filename:  "non-existent.yaml",
			expectErr: true,
			errText:   "failed to read config file",
		},
		{
			name:     "invalid yaml syntax",
			filename: "invalid.yaml",
			content: `
verbose: true
input_folder: [invalid: syntax
`,
			expectErr: true,
			errText:   "failed to unmarshal",
		},
		{
			name:     "validation error",
			filename: "validation-error.yaml",
			content: `
input_folder: ./same
output_folder: ./same
`,
			expectErr: true,
			errText:   "validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
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

			cfg, err := ParseConfig(filePath)

			if tt.expectErr {
				if err == nil {
					t.Error("expected error but got nil")
				} else if tt.errText != "" && !contains(err.Error(), tt.errText) {
					t.Errorf("expected error to contain %q, got %q", tt.errText, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if tt.validate != nil {
					tt.validate(t, cfg)
				}
			}
		})
	}
}

// TestSiteMetadata tests the SiteMetadata struct
func TestSiteMetadata(t *testing.T) {
	t.Parallel()
	metadata := SiteMetadata{
		Title:       "My Blog",
		Description: "A blog about code",
		Author:      "John Doe",
		URL:         "https://example.com",
	}

	if metadata.Title != "My Blog" {
		t.Errorf("expected title 'My Blog', got %s", metadata.Title)
	}
	if metadata.Description != "A blog about code" {
		t.Errorf("expected description 'A blog about code', got %s", metadata.Description)
	}
	if metadata.Author != "John Doe" {
		t.Errorf("expected author 'John Doe', got %s", metadata.Author)
	}
	if metadata.URL != "https://example.com" {
		t.Errorf("expected URL 'https://example.com', got %s", metadata.URL)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

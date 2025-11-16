package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// SiteMetadata contains information about the website
type SiteMetadata struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Author      string `yaml:"author"`
	URL         string `yaml:"url"`
}

// Config represents the configuration structure for the application.
type Config struct {
	// Verbose defines if debug logs should be shown
	Verbose bool `yaml:"verbose"`

	// InputFolder is the directory where input posts are located.
	InputFolder string `yaml:"input_folder"`

	// OutputFolder is the directory where the generated website will be placed.
	OutputFolder string `yaml:"output_folder"`

	// StaticFolder is the directory containing static assets (CSS, JS, images)
	StaticFolder string `yaml:"static_folder"`

	// TemplateDir is the directory containing custom templates (optional)
	// If not specified, default templates will be used from templates/defaults/
	TemplateDir string `yaml:"template_dir"`

	// Site contains metadata about the website
	Site SiteMetadata `yaml:"site"`

	// PostsPerPage defines how many posts to show per page (pagination)
	PostsPerPage int `yaml:"posts_per_page"`

	// BlogPath is the URL path for the blog (e.g., "/blog")
	BlogPath string `yaml:"blog_path"`
}

// ParseConfig reads a YAML configuration file and returns a Config struct.
func ParseConfig(name string) (Config, error) {
	data, err := os.ReadFile(name)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config file %s: %w", name, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal config file %s: %w", name, err)
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// ParseConfigFromBytes parses configuration from byte slice (useful for testing)
func ParseConfigFromBytes(data []byte) (Config, error) {
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// Validate validates and sets defaults for the configuration
func (cfg *Config) Validate() error {
	// Set defaults for folders
	if cfg.InputFolder == "" {
		cfg.InputFolder = "./posts"
	}
	if cfg.OutputFolder == "" {
		cfg.OutputFolder = "./site"
	}
	if cfg.StaticFolder == "" {
		cfg.StaticFolder = "./static"
	}
	if cfg.TemplateDir == "" {
		cfg.TemplateDir = "./templates/defaults"
	}

	// Validate input and output are different
	if cfg.InputFolder == cfg.OutputFolder {
		return fmt.Errorf("input_folder and output_folder cannot be the same")
	}

	// Set defaults for site metadata
	if cfg.Site.Title == "" {
		cfg.Site.Title = "My Blog"
	}
	if cfg.Site.Description == "" {
		cfg.Site.Description = "A blog about interesting things"
	}
	if cfg.Site.Author == "" {
		cfg.Site.Author = "Anonymous"
	}

	// Set defaults for pagination
	if cfg.PostsPerPage <= 0 {
		cfg.PostsPerPage = 10
	} else if cfg.PostsPerPage > 100 {
		cfg.PostsPerPage = 100
	}

	// Set default blog path
	if cfg.BlogPath == "" {
		cfg.BlogPath = "/blog"
	}

	return nil
}

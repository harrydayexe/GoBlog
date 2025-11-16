package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the server configuration
type Config struct {
	Server ServerConfig `yaml:"server"`
	Cache  CacheConfig  `yaml:"cache"`
	Search SearchConfig `yaml:"search"`
	Blog   BlogConfig   `yaml:"blog"`
	Static StaticConfig `yaml:"static"`

	// Content is the path to markdown posts
	ContentFolder string `yaml:"content_folder"`

	// Logging
	Verbose bool `yaml:"verbose"`
}

// ServerConfig contains HTTP server settings
type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// CacheConfig contains cache settings
type CacheConfig struct {
	MaxSizeMB  int64         `yaml:"max_size_mb"`
	TTLMinutes int           `yaml:"ttl_minutes"`
	TTL        time.Duration `yaml:"-"` // Computed from TTLMinutes
}

// SearchConfig contains search index settings
type SearchConfig struct {
	Enabled         bool   `yaml:"enabled"`
	IndexPath       string `yaml:"index_path"`
	RebuildOnStart  bool   `yaml:"rebuild_on_start"`
}

// BlogConfig contains blog-specific settings
type BlogConfig struct {
	Path         string `yaml:"path"`
	PostsPerPage int    `yaml:"posts_per_page"`
}

// StaticConfig contains static file serving settings
type StaticConfig struct {
	Enabled bool   `yaml:"enabled"`
	Folder  string `yaml:"folder"`
}

// Default returns a config with sensible defaults
func Default() *Config {
	return &Config{
		Server: ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
		Cache: CacheConfig{
			MaxSizeMB:  100,
			TTLMinutes: 60,
			TTL:        60 * time.Minute,
		},
		Search: SearchConfig{
			Enabled:        true,
			IndexPath:      "./blog.bleve",
			RebuildOnStart: false,
		},
		Blog: BlogConfig{
			Path:         "/blog",
			PostsPerPage: 10,
		},
		Static: StaticConfig{
			Enabled: false,
			Folder:  "./site",
		},
		ContentFolder: "./posts",
		Verbose:       false,
	}
}

// LoadFromFile loads configuration from a YAML file
func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := Default()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Compute TTL from minutes
	if cfg.Cache.TTLMinutes > 0 {
		cfg.Cache.TTL = time.Duration(cfg.Cache.TTLMinutes) * time.Minute
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate server settings
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid port number: %d (must be 1-65535)", c.Server.Port)
	}

	// Validate content folder exists
	if _, err := os.Stat(c.ContentFolder); os.IsNotExist(err) {
		return fmt.Errorf("content folder does not exist: %s", c.ContentFolder)
	}

	// Validate static folder if enabled
	if c.Static.Enabled {
		if _, err := os.Stat(c.Static.Folder); os.IsNotExist(err) {
			return fmt.Errorf("static folder does not exist: %s", c.Static.Folder)
		}
	}

	// Validate cache settings
	if c.Cache.MaxSizeMB < 1 {
		return fmt.Errorf("cache max size must be at least 1MB")
	}

	if c.Cache.TTLMinutes < 1 {
		return fmt.Errorf("cache TTL must be at least 1 minute")
	}

	// Validate blog settings
	if c.Blog.PostsPerPage < 1 {
		return fmt.Errorf("posts per page must be at least 1")
	}

	return nil
}

// Address returns the full server address
func (c *Config) Address() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

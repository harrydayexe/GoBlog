package config

import (
	"fmt"
	"os"

	"github.com/harrydayexe/GoBlog/internal/gen/log"
	"gopkg.in/yaml.v3"
)

var logger log.CLILogger = *log.NewCLILogger("CONFIG", false)

// SiteMetadata contains information about the website
type SiteMetadata struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Author      string `yaml:"author"`
	URL         string `yaml:"url"`
}

// TemplateConfig contains paths to custom templates
type TemplateConfig struct {
	PostTemplate  string `yaml:"post"`
	IndexTemplate string `yaml:"index"`
	TagTemplate   string `yaml:"tag"`
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

	// Site contains metadata about the website
	Site SiteMetadata `yaml:"site"`

	// Templates contains paths to custom template files
	Templates TemplateConfig `yaml:"templates"`

	// PostsPerPage defines how many posts to show per page (pagination)
	PostsPerPage int `yaml:"posts_per_page"`

	// BlogPath is the URL path for the blog (e.g., "/blog")
	BlogPath string `yaml:"blog_path"`
}

// ParseConfig reads a YAML configuration file and returns a Config struct.
func ParseConfig(name string) (Config, error) {
	data, err := os.ReadFile(name)
	if err != nil {
		werr := fmt.Errorf("failed to read config file %s: %w", name, err)
		logger.Error(werr)
		return Config{}, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		werr := fmt.Errorf("failed to unmarshal config file %s: %w", name, err)
		logger.Error(werr)
		return Config{}, err
	}

	valErr := cfg.validateConfig()
	if valErr != nil {
		werr := fmt.Errorf("config validation failed: %w", valErr)
		logger.Error(werr)
		return Config{}, valErr
	}

	logger.Info("Config file parsed successfully: %+v", name)
	return cfg, nil
}

func (cfg *Config) validateConfig() error {
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
	}

	// Set default blog path
	if cfg.BlogPath == "" {
		cfg.BlogPath = "/blog"
	}

	return nil
}

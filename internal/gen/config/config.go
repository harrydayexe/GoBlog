package config

import (
	"fmt"
	"os"

	"github.com/harrydayexe/GoBlog/internal/gen/log"
	"gopkg.in/yaml.v3"
)

var logger log.CLILogger = *log.NewCLILogger("CONFIG", false)

// Config represents the configuration structure for the application.
type Config struct {
	// Verbose defines if debug logs should be shown
	Verbose bool `yaml:"verbose"`

	// InputFolder is the directory where input posts are located.
	InputFolder string `yaml:"input_folder"`

	// OutputFolder is the directory where the generated website will be placed.
	OutputFolder string `yaml:"output_folder"`
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
	if cfg.InputFolder == "" {
		cfg.InputFolder = "./posts"
	}
	if cfg.OutputFolder == "" {
		cfg.OutputFolder = "./site"
	}
	if cfg.InputFolder == cfg.OutputFolder {
		return fmt.Errorf("input_folder and output_folder cannot be the same")
	}
	return nil
}

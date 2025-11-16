package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/harrydayexe/GoBlog/internal/gen/config"
	"github.com/harrydayexe/GoBlog/internal/gen/generator"
	"github.com/harrydayexe/GoBlog/internal/gen/log"
)

func main() {
	// Parse command line flags
	cfgFileNamePtr := flag.String("config", "config.yaml", "Path to the configuration file")
	flag.Parse()

	// Parse configuration
	cfg, err := config.ParseConfig(*cfgFileNamePtr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse config: %v\n", err)
		os.Exit(1)
	}

	// Set up logger with verbose mode from config
	logger := log.NewCLILogger("MAIN", cfg.Verbose)
	logger.Info("GoBlog Static Site Generator")
	logger.Info("Configuration loaded from: %s", *cfgFileNamePtr)
	logger.Debug("Config: %+v", cfg)

	// Create generator with logger
	gen, err := generator.New(cfg, logger)
	if err != nil {
		logger.Error(fmt.Errorf("failed to create generator: %w", err))
		os.Exit(1)
	}

	// Run generation
	if err := gen.Generate(); err != nil {
		logger.Error(fmt.Errorf("generation failed: %w", err))
		os.Exit(1)
	}

	logger.Info("Done! Your static site is ready.")
}

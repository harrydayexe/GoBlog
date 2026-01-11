package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/harrydayexe/GoBlog/internal/generator"
	"github.com/harrydayexe/GoBlog/internal/logger"
	"github.com/harrydayexe/GoBlog/internal/server"
	"github.com/urfave/cli/v3"
)

// These are replaced at build time by GoReleaser
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	logger := slog.New(logger.NewDefaultCLIHandlerWithVerbosity(os.Stdout, slog.LevelInfo))
	slog.SetDefault(logger)

	cmd := &cli.Command{
		Name:  "GoBlog",
		Usage: "Create a blog feed from posts written in Markdown!",
		Commands: []*cli.Command{
			&generator.GeneratorCommand,
			&server.ServeCommand,
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

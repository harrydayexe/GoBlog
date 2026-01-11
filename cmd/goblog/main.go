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
	var verbosity int

	cmd := &cli.Command{
		Name:                   "GoBlog",
		Usage:                  "Create a blog feed from posts written in Markdown!",
		UseShortOptionHandling: true,
		Commands: []*cli.Command{
			&generator.GeneratorCommand,
			&server.ServeCommand,
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "verbose output (-v, -vv, -vvv for increasing verbosity)",
				Config: cli.BoolConfig{
					Count: &verbosity,
				},
			},
		},
		Before: func(ctx context.Context, c *cli.Command) (context.Context, error) {
			var level slog.Level
			switch verbosity {
			case 0:
				level = slog.LevelWarn
			case 1:
				level = slog.LevelInfo
			case 2:
				level = slog.LevelInfo - 1 // Info logs with params
			default:
				level = slog.LevelDebug // All logs with params
			}

			logger := slog.New(logger.NewDefaultCLIHandlerWithVerbosity(os.Stdout, level))
			slog.SetDefault(logger)

			return ctx, nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

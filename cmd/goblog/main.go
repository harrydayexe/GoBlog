package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/harrydayexe/GoBlog/v2/internal/generator"
	loggermod "github.com/harrydayexe/GoBlog/v2/internal/logger"
	"github.com/harrydayexe/GoBlog/v2/internal/server"
	"github.com/harrydayexe/GoBlog/v2/internal/utilities"
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
	var logger *slog.Logger

	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"V"},
		Usage:   "print only the version",
	}

	cli.VersionPrinter = func(cmd *cli.Command) {
		fmt.Printf("GoBlog version %s\nCommit %s\nDate %s\n", version, commit, date)
	}

	cmd := &cli.Command{
		Name:                   "GoBlog",
		Usage:                  "Create a blog feed from posts written in Markdown!",
		UseShortOptionHandling: true,
		Version:                "unknown",
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

			logger = slog.New(loggermod.NewDefaultCLIHandlerWithVerbosity(os.Stdout, level))
			slog.SetDefault(logger)

			return ctx, nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		utilities.CliErrorHandler(err)
	}
}

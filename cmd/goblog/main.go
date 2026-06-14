// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

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

// version is replaced at build time by GoReleaser
var version = "dev"

// newRootCommand builds and returns the root CLI command. Extracting it from
// main allows the command to be constructed in tests.
func newRootCommand() *cli.Command {
	var verbosity int
	var logger *slog.Logger

	return &cli.Command{
		Name:                   "goblog",
		Usage:                  "Create a blog feed from posts written in Markdown!",
		UseShortOptionHandling: true,
		EnableShellCompletion:  true,
		Version:                buildVersion(),
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
				level = slog.LevelInfo
			case 1:
				level = slog.LevelInfo - 1 // Info logs with params
			default:
				level = slog.LevelDebug // All logs with params
			}

			logger = slog.New(loggermod.NewDefaultCLIHandlerWithVerbosity(os.Stdout, level))
			slog.SetDefault(logger)

			return ctx, nil
		},
	}
}

func main() {
	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"V"},
		Usage:   "print only the version",
	}

	cli.VersionPrinter = func(cmd *cli.Command) {
		fmt.Printf("GoBlog version %s\n", buildVersion())
	}

	if err := newRootCommand().Run(context.Background(), os.Args); err != nil {
		utilities.CliErrorHandler(err)
	}
}

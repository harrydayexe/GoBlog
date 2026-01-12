package generator

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

// GeneratorCommand is the main entry point for the GoBlog generate CLI tool.
// This command is used to set up all the flags, usage info and the action to
// for generate.
var GeneratorCommand cli.Command = cli.Command{
	Name:    "generate",
	Aliases: []string{"g"},
	Usage:   "generate a static blog feed from markdown posts",
	Action: func(ctx context.Context, c *cli.Command) error {
		fmt.Println("generate")
		return nil
	},
	UseShortOptionHandling: true,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "raw",
			Aliases: []string{"r"},
			Usage:   "output raw HTML without template wrapper",
			Value:   false,
		},
	},
}

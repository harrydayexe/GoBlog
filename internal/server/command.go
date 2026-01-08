package server

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

// ServeCommand is the main entry point for the GoBlog serve CLI tool.
// This command is used to set up all the flags, usage info and the action to
// for serve.
var ServeCommand cli.Command = cli.Command{
	Name:    "serve",
	Aliases: []string{"s"},
	Usage:   "serve a static blog feed from markdown posts",
	Action: func(ctx context.Context, c *cli.Command) error {
		fmt.Println("serve")
		return nil
	},
}

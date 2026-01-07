package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v3"
)

// These are replaced at build time by GoReleaser
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd := &cli.Command{
		Name:  "GoBlog",
		Usage: "Create a blog feed from posts written in Markdown!",
		Commands: []*cli.Command{
			{
				Name:    "generate",
				Aliases: []string{"g"},
				Usage:   "generate a static blog feed from markdown posts",
				Action: func(ctx context.Context, c *cli.Command) error {
					fmt.Println("generate")
					return nil
				},
			},
			{
				Name:    "serve",
				Aliases: []string{"s"},
				Usage:   "serve a static blog feed from markdown posts",
				Action: func(ctx context.Context, c *cli.Command) error {
					fmt.Println("serve")
					return nil
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

package server

import (
	"github.com/urfave/cli/v3"
)

// ServeCommand is the main entry point for the GoBlog serve CLI tool.
// This command sets up all the flags, usage info, and action for serve.
var ServeCommand cli.Command = cli.Command{
	Name:                   "serve",
	Aliases:                []string{"s"},
	Usage:                  "serve a blog from markdown posts over HTTP",
	Action:                 NewServeCommand,
	UseShortOptionHandling: true,
	Arguments: []cli.Argument{
		&cli.StringArg{
			Name:      InputPostsDirArgName,
			UsageText: "<input directory>",
		},
	},
	Flags: []cli.Flag{
		&cli.IntFlag{
			Name:    PortFlagName,
			Aliases: []string{"P"},
			Usage:   "port to listen on",
			Value:   8080,
		},
		&cli.StringFlag{
			Name:    HostFlagName,
			Aliases: []string{"H"},
			Usage:   "host address to bind to",
		},
		&cli.StringFlag{
			Name:    TemplateDirFlagName,
			Aliases: []string{"t"},
			Usage:   "directory of templates to use when rendering",
		},
		&cli.StringFlag{
			Name:    BlogRootFlagName,
			Aliases: []string{"p"},
			Usage:   "root path of the blog, defaults to '/'",
		},
	},
}

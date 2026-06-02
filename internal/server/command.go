// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

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
		&cli.BoolFlag{
			Name:    DisableTagsFlagName,
			Aliases: []string{"T"},
			Usage:   "disable tag tracking and tag page generation",
			Value:   false,
		},
		&cli.BoolFlag{
			Name:  DisableReadingTimeFlagName,
			Usage: "disable reading time estimation on posts",
			Value: false,
		},
		&cli.BoolFlag{
			Name:    WatchFlagName,
			Aliases: []string{"w"},
			Usage:   "watch the posts directory and regenerate the blog on changes",
			Value:   false,
		},
	},
}

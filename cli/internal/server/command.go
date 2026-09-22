// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package server

import (
	"time"

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
		&cli.BoolFlag{
			Name:    WatchFlagName,
			Aliases: []string{"w"},
			Usage:   "watch the posts directory and regenerate the blog on changes",
			Value:   false,
		},
		&cli.DurationFlag{
			Name:  CacheControlFlagName,
			Usage: "max-age TTL for the Cache-Control header (0 disables the header)",
			Value: time.Hour,
		},
		&cli.BoolFlag{
			Name:  HealthChecksFlagName,
			Usage: "expose /healthz/live, /healthz/ready, and /healthz/startup endpoints (no auth required); the server binds before loading content so probes observe startup state",
			Value: false,
		},
		&cli.BoolFlag{
			Name:  MetricsFlagName,
			Usage: "record Prometheus metrics and serve them at /metrics on a separate admin listener (never on the blog port)",
			Value: false,
		},
		&cli.IntFlag{
			Name:  MetricsPortFlagName,
			Usage: "port the admin listener serving /metrics binds to",
			Value: defaultMetricsPort,
		},
		&cli.StringFlag{
			Name:  MetricsHostFlagName,
			Usage: "host address the admin listener serving /metrics binds to (defaults to all interfaces; set 127.0.0.1 to keep metrics off the network)",
		},
	},
}

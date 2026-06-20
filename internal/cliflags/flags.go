// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package cliflags defines the CLI flag names and flag definitions that are
// shared between the generate and serve subcommands. Keeping them here avoids
// duplication and import cycles: both the root goblog command (which registers
// the flags on the top-level cli.Command) and the subcommand action packages
// (which read flag values) can import this package.
package cliflags

import "github.com/urfave/cli/v3"

// TemplateDirFlagName is the CLI flag name for setting a custom template directory.
const TemplateDirFlagName = "template-dir"

// BlogRootFlagName is the CLI flag name for setting the blog root path.
const BlogRootFlagName = "root-path"

// DisableTagsFlagName is the CLI flag name for disabling tag page generation.
const DisableTagsFlagName = "disable-tags"

// DisableReadingTimeFlagName is the CLI flag name for disabling reading time estimation.
const DisableReadingTimeFlagName = "disable-reading-time"

// BaseURLFlagName is the CLI flag name for setting the site's absolute base URL
// (required for RSS/Atom feed generation).
const BaseURLFlagName = "base-url"

// DisableFeedsFlagName is the CLI flag name for disabling RSS and Atom feed generation.
const DisableFeedsFlagName = "disable-feeds"

// FeedLimitFlagName is the CLI flag name for setting the maximum number of posts
// in each feed.
const FeedLimitFlagName = "feed-limit"

// Shared returns the CLI flag definitions that apply to both the generate and
// serve subcommands. They are registered on the top-level goblog command so that
// urfave/cli v3's default persistence makes them available in both subcommand
// actions via c.String/Bool/Int.
func Shared() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    TemplateDirFlagName,
			Aliases: []string{"t"},
			Usage:   "directory of templates to use when rendering",
		},
		&cli.StringFlag{
			Name:    BlogRootFlagName,
			Aliases: []string{"p"},
			Usage:   "root path of the blog, defaults to '/'",
			Value:   "/",
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
		&cli.StringFlag{
			Name:  BaseURLFlagName,
			Usage: "absolute base URL of the site (e.g. https://example.com); required for RSS/Atom feed generation",
		},
		&cli.BoolFlag{
			Name:  DisableFeedsFlagName,
			Usage: "disable RSS and Atom feed generation",
			Value: false,
		},
		&cli.IntFlag{
			Name:  FeedLimitFlagName,
			Usage: "maximum number of posts to include in each feed",
			Value: 10,
		},
	}
}

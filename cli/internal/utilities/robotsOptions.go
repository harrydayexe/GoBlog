// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package utilities

import (
	"os"

	"github.com/harrydayexe/GoBlog/v2/cli/internal/cliflags"
	inerrors "github.com/harrydayexe/GoBlog/v2/cli/internal/errors"
	"github.com/harrydayexe/GoBlog/v2/pkg/config"
	"github.com/urfave/cli/v3"
)

// RobotsOptions validates the sitemap and robots.txt flags and converts them
// into generator options. It is shared by the generate and serve subcommands,
// which expose the same flags.
//
// It is called before any generation work so that contradictory flags fail
// fast:
//   - --robots-file together with --disable-robots is an error. One supplies a
//     custom robots.txt and the other suppresses robots.txt entirely; honouring
//     the disable would silently discard the user's file.
//   - --robots-file naming a missing or unreadable path is an error rather
//     than a silent fallback to the default rules.
func RobotsOptions(c *cli.Command) ([]config.GeneratorOption, error) {
	var opts []config.GeneratorOption

	if c.Bool(cliflags.DisableSitemapFlagName) {
		opts = append(opts, config.WithDisableSitemap())
	}

	disableRobots := c.Bool(cliflags.DisableRobotsFlagName)
	robotsFile := c.String(cliflags.RobotsFileFlagName)

	if robotsFile != "" && disableRobots {
		return nil, inerrors.NewConflictingFlagsError(
			cliflags.RobotsFileFlagName,
			cliflags.DisableRobotsFlagName,
			"one supplies a custom robots.txt, the other suppresses robots.txt entirely",
		)
	}

	if disableRobots {
		opts = append(opts, config.WithDisableRobotsTxt())
		return opts, nil
	}

	if robotsFile != "" {
		body, err := os.ReadFile(robotsFile)
		if err != nil {
			return nil, inerrors.NewFlagFileError(cliflags.RobotsFileFlagName, robotsFile, err)
		}
		opts = append(opts, config.WithRobotsTxt(string(body)))
	}

	return opts, nil
}

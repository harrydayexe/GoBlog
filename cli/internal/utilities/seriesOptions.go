// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package utilities

import (
	"os"
	"path/filepath"

	"github.com/harrydayexe/GoBlog/v2/cli/internal/cliflags"
	inerrors "github.com/harrydayexe/GoBlog/v2/cli/internal/errors"
	"github.com/harrydayexe/GoBlog/v2/pkg/config"
	"github.com/urfave/cli/v3"
)

// SeriesOptions validates the --series-file flag and converts it into a
// generator option. It is shared by the generate and serve subcommands, which
// expose the same flag.
//
// It also returns the resolved path of the series file, so serve can watch it
// for changes. Both the option slice and the path are empty when the flag was
// not supplied: series are opt-in and stay off.
//
// A relative path is resolved from the current working directory, like the
// other path flags. The file is read through the filesystem of its parent
// directory so that the generator re-reads it on every regeneration, which is
// what makes editing it during goblog serve --watch take effect.
//
// It is called before any generation work so that a missing or unreadable file
// fails fast rather than silently falling back to "series disabled".
func SeriesOptions(c *cli.Command) ([]config.GeneratorOption, string, error) {
	seriesFile := c.String(cliflags.SeriesFileFlagName)
	if seriesFile == "" {
		return nil, "", nil
	}

	clean := filepath.Clean(seriesFile)
	if _, err := os.ReadFile(clean); err != nil {
		return nil, "", inerrors.NewFlagFileError(cliflags.SeriesFileFlagName, clean, err)
	}

	opts := []config.GeneratorOption{
		config.WithSeriesFile(os.DirFS(filepath.Dir(clean)), filepath.Base(clean)),
	}
	return opts, clean, nil
}

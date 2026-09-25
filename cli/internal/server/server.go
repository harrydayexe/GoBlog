// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package server

import (
	"context"
	"io/fs"
	"log/slog"
	"os"
	"path"
	"strings"

	"github.com/harrydayexe/GoBlog/v2/cli/internal/cliflags"
	"github.com/harrydayexe/GoBlog/v2/cli/internal/utilities"
	"github.com/harrydayexe/GoBlog/v2/pkg/config"
	"github.com/harrydayexe/GoBlog/v2/pkg/server"
	"github.com/harrydayexe/GoBlog/v2/pkg/templates"
	"github.com/harrydayexe/GoBlog/v2/pkg/watcher"
	gwucfg "github.com/harrydayexe/GoWebUtilities/config"
	"github.com/urfave/cli/v3"
)

// NewServeCommand handles the serve command by starting an HTTP server from a directory of markdown posts.
func NewServeCommand(ctx context.Context, c *cli.Command) error {
	// Validate the series file first: a path that cannot be read is an error,
	// not a silent fallback to "series disabled", and it should fail before any
	// directories are touched or posts are parsed.
	seriesOpts, seriesFile, err := utilities.SeriesOptions(c)
	if err != nil {
		return err
	}

	healthChecksEnabled := c.Bool(HealthChecksFlagName)

	inputPostsDir := c.StringArg(InputPostsDirArgName)
	// When health checks are enabled the posts directory may not exist yet
	// (e.g. the volume is not mounted). Allow a non-existent path so the server
	// can start and surface the failure via /healthz/ready rather than exiting.
	inputPostsDir, err = utilities.GetDirectoryFromInput(inputPostsDir, healthChecksEnabled)
	if err != nil {
		return err
	}
	postsFsys := os.DirFS(inputPostsDir)

	envCfg, err := gwucfg.ParseConfig[config.EnvironmentConfig]()
	if err != nil {
		return err
	}

	opts := []config.ServerOption{
		config.WithEnvironment(string(envCfg.Environment)).AsServerOption(),
	}

	if c.Bool(cliflags.DisableTagsFlagName) {
		opts = append(opts, config.WithDisableTags().AsServerOption())
	}

	if c.Bool(cliflags.DisableReadingTimeFlagName) {
		opts = append(opts, config.WithDisableReadingTime().AsServerOption())
	}

	if baseURL := c.String(cliflags.BaseURLFlagName); baseURL != "" {
		opts = append(opts, config.WithBaseURL(baseURL).AsServerOption())
	}

	if c.Bool(cliflags.DisableFeedsFlagName) {
		opts = append(opts, config.WithDisableFeeds().AsServerOption())
	}

	opts = append(opts, config.WithFeedPostLimit(c.Int(cliflags.FeedLimitFlagName)).AsServerOption())

	for _, seriesOpt := range seriesOpts {
		opts = append(opts, seriesOpt.AsServerOption())
	}

	opts = append(opts, config.WithPort(c.Int(PortFlagName)))
	opts = append(opts, config.WithCacheControl(c.Duration(CacheControlFlagName)))

	if host := c.String(HostFlagName); host != "" {
		opts = append(opts, config.WithHost(host))
	}

	templateDirPath := c.String(cliflags.TemplateDirFlagName)
	if templateDirPath == "" {
		slog.Default().DebugContext(ctx, "Using default templates")
		opts = append(opts, config.WithTemplateDir(templates.Default))
	} else {
		slog.Default().DebugContext(ctx, "Using custom templates")
		templateDirPath, err = utilities.GetDirectoryFromInput(templateDirPath, false)
		if err != nil {
			return err
		}
		opts = append(opts, config.WithTemplateDir(os.DirFS(templateDirPath)))
	}

	blogRootString := c.String(cliflags.BlogRootFlagName)
	if blogRootString != "" {
		blogRoot := path.Clean(blogRootString)
		blogRoot = strings.TrimPrefix(blogRoot, ".")
		if !strings.HasPrefix(blogRoot, "/") {
			blogRoot = "/" + blogRoot
		}
		if !strings.HasSuffix(blogRoot, "/") {
			blogRoot += "/"
		}
		// The server forwards its resolved blog root to the generator, so the
		// option only needs applying once.
		opts = append(opts, config.WithBlogRoot(blogRoot).AsServerOption())
	}

	opts = append(opts, config.WithLogger(slog.Default()).AsServerOption())

	assetsRoot, err := utilities.OpenAssetsDir(c.String(cliflags.AssetsDirFlagName), inputPostsDir)
	if err != nil {
		return err
	}
	if assetsRoot != nil {
		defer assetsRoot.Close()
		opts = append(opts, config.WithAssetsDir(assetsRoot.FS()).AsServerOption())
	}

	if healthChecksEnabled {
		opts = append(opts, config.WithHealthChecks())
	}

	return runServe(ctx, inputPostsDir, postsFsys, c.Bool(WatchFlagName), seriesFile, opts...)
}

func runServe(ctx context.Context, postsPath string, posts fs.FS, watch bool, seriesFile string, opts ...config.ServerOption) error {
	srv, err := server.New(posts, opts...)
	if err != nil {
		return err
	}

	if watch {
		watchOpts := []config.WatcherOption{srv.Logger.AsOption().AsWatcherOption()}
		if seriesFile != "" {
			// The series file is neither markdown nor necessarily inside the
			// posts directory, so it has to be watched explicitly.
			watchOpts = append(watchOpts, config.WithWatchFile(seriesFile))
		}

		w, err := watcher.New(postsPath, watchOpts...)
		if err != nil {
			return err
		}
		go func() {
			if err := w.Run(ctx, func(ctx context.Context) {
				if err := srv.UpdatePosts(os.DirFS(postsPath), ctx); err != nil {
					slog.Default().WarnContext(ctx, "watcher: failed to reload posts", slog.Any("error", err))
				}
			}); err != nil {
				slog.Default().WarnContext(ctx, "watcher: stopped with error", slog.Any("error", err))
			}
		}()
	}

	return srv.Run(ctx)
}

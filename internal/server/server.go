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

	"github.com/harrydayexe/GoBlog/v2/internal/cliflags"
	"github.com/harrydayexe/GoBlog/v2/internal/utilities"
	"github.com/harrydayexe/GoBlog/v2/pkg/config"
	"github.com/harrydayexe/GoBlog/v2/pkg/server"
	"github.com/harrydayexe/GoBlog/v2/pkg/templates"
	"github.com/harrydayexe/GoBlog/v2/pkg/watcher"
	gwucfg "github.com/harrydayexe/GoWebUtilities/config"
	"github.com/urfave/cli/v3"
)

// NewServeCommand handles the serve command by starting an HTTP server from a directory of markdown posts.
func NewServeCommand(ctx context.Context, c *cli.Command) error {
	// Validate the sitemap/robots flags first so contradictory or unreadable
	// values fail before any directories are touched or posts are parsed.
	robotsOpts, err := utilities.RobotsOptions(c)
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

	cfg := config.ServerConfig{}

	cfg.Gen = append(cfg.Gen, config.WithEnvironment(string(envCfg.Environment)))

	if c.Bool(cliflags.DisableTagsFlagName) {
		cfg.Gen = append(cfg.Gen, config.WithDisableTags())
	}

	if c.Bool(cliflags.DisableReadingTimeFlagName) {
		cfg.Gen = append(cfg.Gen, config.WithDisableReadingTime())
	}

	if baseURL := c.String(cliflags.BaseURLFlagName); baseURL != "" {
		cfg.Gen = append(cfg.Gen, config.WithBaseURL(baseURL))
	}

	if c.Bool(cliflags.DisableFeedsFlagName) {
		cfg.Gen = append(cfg.Gen, config.WithDisableFeeds())
	}

	cfg.Gen = append(cfg.Gen, config.WithFeedPostLimit(c.Int(cliflags.FeedLimitFlagName)))

	cfg.Gen = append(cfg.Gen, robotsOpts...)

	cfg.Server = append(cfg.Server, config.WithPort(c.Int(PortFlagName)))
	cfg.Server = append(cfg.Server, config.WithCacheControl(c.Duration(CacheControlFlagName)))

	if host := c.String(HostFlagName); host != "" {
		cfg.Server = append(cfg.Server, config.WithHost(host))
	}

	templateDirPath := c.String(cliflags.TemplateDirFlagName)
	if templateDirPath == "" {
		slog.Default().DebugContext(ctx, "Using default templates")
		cfg.TemplateDir = templates.Default
	} else {
		slog.Default().DebugContext(ctx, "Using custom templates")
		templateDirPath, err = utilities.GetDirectoryFromInput(templateDirPath, false)
		if err != nil {
			return err
		}
		cfg.TemplateDir = os.DirFS(templateDirPath)
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
		cfg.Server = append(cfg.Server, config.WithBlogRoot(blogRoot).AsServerOption())
		cfg.Gen = append(cfg.Gen, config.WithBlogRoot(blogRoot).AsGeneratorOption())
	}

	cfg.Server = append(cfg.Server, config.WithLogger(slog.Default()).AsServerOption())

	assetsRoot, err := utilities.OpenAssetsDir(c.String(cliflags.AssetsDirFlagName), inputPostsDir)
	if err != nil {
		return err
	}
	if assetsRoot != nil {
		defer assetsRoot.Close()
		cfg.Server = append(cfg.Server, config.WithAssetsDir(assetsRoot.FS()).AsServerOption())
	}

	if healthChecksEnabled {
		cfg.Server = append(cfg.Server, config.WithHealthChecks())
	}

	return runServe(ctx, inputPostsDir, postsFsys, cfg, c.Bool(WatchFlagName))
}

func runServe(ctx context.Context, postsPath string, posts fs.FS, cfg config.ServerConfig, watch bool) error {
	srv, err := server.New(nil, posts, cfg)
	if err != nil {
		return err
	}

	if watch {
		w, err := watcher.New(postsPath, srv.Logger.AsOption().AsWatcherOption())
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

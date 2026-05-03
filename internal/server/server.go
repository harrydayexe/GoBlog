package server

import (
	"context"
	"io/fs"
	"log/slog"
	"os"
	"path"
	"strings"

	"github.com/harrydayexe/GoBlog/v2/internal/utilities"
	"github.com/harrydayexe/GoBlog/v2/pkg/config"
	"github.com/harrydayexe/GoBlog/v2/pkg/server"
	"github.com/harrydayexe/GoBlog/v2/pkg/templates"
	gwucfg "github.com/harrydayexe/GoWebUtilities/config"
	"github.com/urfave/cli/v3"
)

// NewServeCommand handles the serve command by starting an HTTP server from a directory of markdown posts.
func NewServeCommand(ctx context.Context, c *cli.Command) error {
	inputPostsDir := c.StringArg(InputPostsDirArgName)
	inputPostsDir, err := utilities.GetDirectoryFromInput(inputPostsDir, false)
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

	if c.Bool(DisableTagsFlagName) {
		cfg.Gen = append(cfg.Gen, config.WithDisableTags())
	}
	cfg.Server = append(cfg.Server, config.WithPort(c.Int(PortFlagName)))

	if host := c.String(HostFlagName); host != "" {
		cfg.Server = append(cfg.Server, config.WithHost(host))
	}

	templateDirPath := c.String(TemplateDirFlagName)
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

	blogRootString := c.String(BlogRootFlagName)
	if blogRootString != "" {
		blogRoot := path.Clean(blogRootString)
		blogRoot = strings.TrimPrefix(blogRoot, ".")
		if !strings.HasPrefix(blogRoot, "/") {
			blogRoot = "/" + blogRoot
		}
		if !strings.HasSuffix(blogRoot, "/") {
			blogRoot += "/"
		}
		cfg.Server = append(cfg.Server, config.BaseServerOption{BaseOption: config.WithBlogRoot(blogRoot)})
		cfg.Gen = append(cfg.Gen, config.WithBaseOption(config.WithBlogRoot(blogRoot)))
	}

	return runServe(ctx, slog.Default(), postsFsys, cfg)
}

func runServe(ctx context.Context, logger *slog.Logger, posts fs.FS, cfg config.ServerConfig) error {
	srv, err := server.New(logger, posts, cfg)
	if err != nil {
		return err
	}
	return srv.Run(ctx)
}

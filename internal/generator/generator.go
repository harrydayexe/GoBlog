package generator

import (
	"context"
	"io/fs"

	"github.com/harrydayexe/GoBlog/v2/internal/utilities"
	"github.com/harrydayexe/GoBlog/v2/pkg/config"
	"github.com/harrydayexe/GoBlog/v2/pkg/generator"
	"github.com/harrydayexe/GoBlog/v2/pkg/outputter"
	"github.com/urfave/cli/v3"
)

// NewGeneratorCommand handles the generate command by processing markdown posts into HTML.
func NewGeneratorCommand(ctx context.Context, c *cli.Command) error {
	inputPostsDir := c.StringArg(InputPostsDirArgName)
	outputDir := c.StringArg(OutputDirArgName)
	postsFsys, err := utilities.GetDirectoryFromInput(inputPostsDir)
	if err != nil {
		return err
	}

	opts := []config.CommonOption{}

	rawOutputFlag := c.Bool(RawOutputFlagName)
	if rawOutputFlag {
		opts = append(opts, config.WithRawOutput())
	}

	handler := outputter.NewDirectoryWriter(outputDir, opts...)

	return runGenerate(ctx, postsFsys, opts, handler)
}

func runGenerate(ctx context.Context, postsFsys fs.FS, opts []config.CommonOption, handler outputter.Outputter) error {
	gen := generator.New(postsFsys, opts...)
	gen.DebugConfig(ctx)

	blog, err := gen.Generate(ctx)
	if err != nil {
		return err
	}

	return handler.HandleGeneratedBlog(blog)
}

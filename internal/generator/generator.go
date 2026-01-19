package generator

import (
	"context"
	"io/fs"
	"os"

	"github.com/harrydayexe/GoBlog/v2/internal/utilities"
	"github.com/harrydayexe/GoBlog/v2/pkg/config"
	"github.com/harrydayexe/GoBlog/v2/pkg/generator"
	"github.com/harrydayexe/GoBlog/v2/pkg/outputter"
	"github.com/harrydayexe/GoBlog/v2/pkg/templates"
	"github.com/urfave/cli/v3"
)

// NewGeneratorCommand handles the generate command by processing markdown posts into HTML.
func NewGeneratorCommand(ctx context.Context, c *cli.Command) error {
	inputPostsDir := c.StringArg(InputPostsDirArgName)
	inputPostsDir, err := utilities.GetDirectoryFromInput(inputPostsDir, false)
	if err != nil {
		return err
	}
	postsFsys := os.DirFS(inputPostsDir)

	outputDirString := c.StringArg(OutputDirArgName)
	outputDir, err := utilities.GetDirectoryFromInput(outputDirString, true)
	if err != nil {
		return err
	}

	opts := []config.Option{}

	rawOutputFlag := c.Bool(RawOutputFlagName)
	if rawOutputFlag {
		opts = append(opts, config.WithRawOutput())
	}

	templateDirPath := c.String(TemplateDirFlagName)
	var templateDir fs.FS
	if templateDirPath == "" {
		templateDir = templates.Default
	} else {
		templateDirPath, err = utilities.GetDirectoryFromInput(templateDirPath, false)
		if err != nil {
			return nil
		}

		templateDir = os.DirFS(templateDirPath)
	}

	handler := outputter.NewDirectoryWriter(outputDir, opts...)

	return runGenerate(ctx, postsFsys, templateDir, opts, handler)
}

func runGenerate(ctx context.Context, postsFsys fs.FS, templatesFsys fs.FS, opts []config.Option, handler outputter.Outputter) error {
	gen := generator.New(postsFsys, templatesFsys, opts...)
	gen.DebugConfig(ctx)

	blog, err := gen.Generate(ctx)
	if err != nil {
		return err
	}

	return handler.HandleGeneratedBlog(blog)
}

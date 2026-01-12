package generator

import (
	"context"
	"os"

	"github.com/harrydayexe/GoBlog/v2/pkg/generator"
	"github.com/urfave/cli/v3"
)

func NewGeneratorCommand(ctx context.Context, c *cli.Command) error {
	inputPostsDir := c.String(InputPostsDirFlagName)
	postsFsys := os.DirFS(inputPostsDir)

	options := []generator.Option{}

	rawOutputFlag := c.Bool(RawOutputFlagName)
	if rawOutputFlag {
		options = append(options, generator.WithRawOutput())
	}

	gen, err := generator.New(postsFsys, options...)
	if err != nil {
		return err
	}

	gen.DebugConfig(ctx)

	return nil
}

package generator

import (
	"context"
	"fmt"
	"os"

	"github.com/harrydayexe/GoBlog/v2/pkg/generator"
	"github.com/urfave/cli/v3"
)

func NewGeneratorCommand(ctx context.Context, c *cli.Command) error {
	inputPostsDir := c.StringArg(InputPostsDirArgName)
	postsFsys := os.DirFS(inputPostsDir)

	options := []generator.Option{}

	rawOutputFlag := c.Bool(RawOutputFlagName)
	if rawOutputFlag {
		options = append(options, generator.WithRawOutput())
	}

	gen := generator.New(postsFsys, options...)
	gen.DebugConfig(ctx)

	blog, err := gen.Generate(ctx)
	if err != nil {
		return err
	}

	for slug, post := range blog.Posts {
		fmt.Printf("====== Post: %s ======\n", slug)
		fmt.Printf("%s\n\n", post)
	}

	return nil
}

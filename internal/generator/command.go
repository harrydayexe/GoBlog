package generator

import (
	"github.com/urfave/cli/v3"
)

// GeneratorCommand is the main entry point for the GoBlog generate CLI tool.
// This command is used to set up all the flags, usage info and the action to
// for generate.
var GeneratorCommand cli.Command = cli.Command{
	Name:                   "generate",
	Aliases:                []string{"g"},
	Usage:                  "generate a static blog feed from markdown posts",
	Action:                 NewGeneratorCommand,
	UseShortOptionHandling: true,
	Arguments: []cli.Argument{
		&cli.StringArg{
			Name:      InputPostsDirArgName,
			UsageText: "directory containing markdown posts",
		},
	},
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    RawOutputFlagName,
			Aliases: []string{"r"},
			Usage:   "output raw HTML without template wrapper",
			Value:   false,
		},
	},
}

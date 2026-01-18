package generator

import (
	"fmt"
	"io/fs"

	"github.com/harrydayexe/GoBlog/v2/pkg/config"
	"github.com/harrydayexe/GoBlog/v2/pkg/parser"
)

// GeneratorConfig contains all the configuration to control how a Generator
// operates.
//
// PostsDir specifies the filesystem containing markdown post files. Each file
// should contain front matter metadata and post content.
//
// TemplatesDir specifies the filesystem containing HTML templates for rendering
// the blog. If not specified, default templates will be used.
type GeneratorConfig struct {
	config.CommonConfig
	PostsDir     fs.FS         // The filesystem containing the input posts in markdown
	TemplatesDir fs.FS         // The filesystem containing the templates to use
	ParserConfig parser.Config // The config to use when parsing
}

func (c GeneratorConfig) String() string {
	return fmt.Sprintf("Generator Config\n - RawOutput %t\n", c.RawOutput)
}

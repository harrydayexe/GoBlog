package generator

import (
	"fmt"
	"io/fs"

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
	PostsDir     fs.FS // The filesystem containing the input posts in markdown
	TemplatesDir fs.FS // The filesystem containing the templates to use
	// When true, output should contain only the raw HTML and not be inserted
	// into a template
	RawOutput    bool
	ParserConfig parser.Config // The config to use when parsing
}

func (c GeneratorConfig) String() string {
	return fmt.Sprintf("Generator Config\n - RawOutput %t\n", c.RawOutput)
}

// Option is a function which modifies a GeneratorConfig.
// Options are used to configure optional parameters when creating a new Generator.
type Option func(*GeneratorConfig)

// WithRawOutput sets the config to only generate the raw HTML for each post
// without inserting it into a template.
func WithRawOutput() Option {
	return func(gc *GeneratorConfig) {
		gc.RawOutput = true
	}
}

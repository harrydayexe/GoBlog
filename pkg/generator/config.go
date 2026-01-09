package generator

import "io/fs"

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
}

// Option is a function which modifies a GeneratorConfig.
// Options are used to configure optional parameters when creating a new Generator.
type Option func(*GeneratorConfig)

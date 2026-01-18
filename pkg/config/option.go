package config

import "io/fs"

// Option represents a configuration option that can be applied to
// generator or outputter instances during construction.
//
// Options use the functional options pattern, where each option function
// returns an Option struct containing one or more function pointers that
// modify specific configuration fields.
//
// This type should not be constructed directly by users. Instead, use the
// provided option functions like WithRawOutput() and WithTemplatesDir().
type Option struct {
	WithRawOutputFunc    func(v *RawOutput)
	WithTemplatesDirFunc func(v *TemplatesDir)
}

// RawOutput is a configuration type that controls whether HTML output
// is generated with or without template wrapping.
//
// When RawOutput is true:
//   - The generator produces only Markdown-to-HTML conversion without templates
//   - The outputter skips creating the tags directory
//   - Individual post files contain raw HTML fragments
//
// This type is typically embedded in generator and outputter configuration
// structs and should be set using the WithRawOutput() option function.
type RawOutput struct{ RawOutput bool }

// WithRawOutput returns an Option that enables raw HTML output mode.
//
// When this option is applied to a generator, it will produce HTML content
// without template wrapping - only the Markdown-to-HTML conversion is performed.
// When applied to an outputter, it will skip creating the tags directory.
//
// This is useful for scenarios where you want to integrate GoBlog's HTML
// output into your own templates or existing site structure.
//
// Example usage:
//
//	gen := generator.New(fsys, config.WithRawOutput())
//	writer := outputter.NewDirectoryWriter("output/", config.WithRawOutput())
func WithRawOutput() Option {
	return Option{
		WithRawOutputFunc: func(v *RawOutput) {
			v.RawOutput = true
		},
	}
}

// TemplatesDir is a configuration type that specifies the filesystem
// containing template files for rendering blog pages.
//
// The filesystem should contain template files used by the generator
// to render blog posts, tag pages, and the index page. If not specified,
// the generator will use default templates.
//
// This type is typically embedded in generator configuration structs and
// should be set using the WithTemplatesDir() option function.
type TemplatesDir struct{ TemplatesDir fs.FS }

// WithTemplatesDir returns an Option that sets a custom template filesystem.
//
// The provided filesystem should contain the template files that the generator
// will use to render blog pages. This allows you to customize the appearance
// of your blog by providing your own templates instead of using the defaults.
//
// The templatesDir parameter must be a valid fs.FS implementation containing
// the necessary template files.
//
// Example usage:
//
//	templateFS := os.DirFS("custom-templates/")
//	gen := generator.New(postsFS, config.WithTemplatesDir(templateFS))
func WithTemplatesDir(templatesDir fs.FS) Option {
	return Option{
		WithTemplatesDirFunc: func(v *TemplatesDir) {
			v.TemplatesDir = templatesDir
		},
	}
}

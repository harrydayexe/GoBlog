package config

type CommonConfig struct {
	// When true, output should contain only the raw HTML and not be inserted
	// into a template
	RawOutput bool
}

// CommonOption is a function which modifies a CommonConfig.
// Options are used to configure optional parameters when creating a new config.
type CommonOption func(*CommonConfig)

// WithRawOutput sets the config to only generate the raw HTML for each post
// without inserting it into a template.
//
// This option affects both Generator and Outputter behavior:
//   - Generator: Returns HTML content without template wrappers
//   - DirectoryWriter: Skips creating the tags/ directory
//
// Use this when you need to:
//   - Integrate GoBlog content into an existing site or CMS
//   - Apply your own custom templates programmatically
//   - Process HTML content before final rendering
//   - Embed blog posts in other web frameworks
//
// Example:
//
//     gen := generator.New(fsys, config.WithRawOutput())
//     blog, _ := gen.Generate(ctx)
//     // blog.Posts contains raw HTML fragments
//
// When raw output is enabled, the GeneratedBlog.Tags map will be empty
// and GeneratedBlog.Posts will contain only the parsed Markdown as HTML.
func WithRawOutput() CommonOption {
	return func(gc *CommonConfig) {
		gc.RawOutput = true
	}
}

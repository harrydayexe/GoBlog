// Package config provides common configuration options used across
// the GoBlog generator and outputter packages.
//
// The package implements the functional options pattern for configuration,
// allowing packages to accept both required positional parameters and optional
// configuration via option functions.
//
// # Functional Options Pattern
//
// Configuration is applied through Option values that can be passed to
// constructors. Each option function modifies specific configuration fields:
//
//	gen := generator.New(postsFS, config.WithRawOutput())
//	writer := outputter.NewDirectoryWriter("output/", config.WithRawOutput())
//
// Options are implemented as structs containing functions that modify
// embedded configuration types. This pattern allows for:
//   - Backward compatibility when adding new options
//   - Clear, self-documenting API calls
//   - Optional parameters without function overloading
//
// # Available Options
//
// WithRawOutput() enables raw HTML output mode without template wrapping.
// When enabled in the generator, markdown is converted to HTML without
// inserting it into page templates. When enabled in the outputter, the
// tags directory is not created.
//
// WithTemplatesDir(fs.FS) specifies a custom filesystem containing templates
// to use for rendering blog pages. This allows you to provide your own
// template files instead of using the defaults.
//
// # Usage Examples
//
// Basic usage with a single option:
//
//	fsys := os.DirFS("posts/")
//	gen := generator.New(fsys, config.WithRawOutput())
//
// Multiple options can be combined:
//
//	templateFS := os.DirFS("templates/")
//	gen := generator.New(fsys,
//	    config.WithRawOutput(),
//	    config.WithTemplatesDir(templateFS),
//	)
//
// # Concurrency
//
// Option values are safe to create and use concurrently. Configuration
// structs that embed RawOutput and TemplatesDir are safe to read
// concurrently once created, but should not be modified after construction.
package config

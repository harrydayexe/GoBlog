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
//	gen := generator.New(postsFS, nil, config.WithRawOutput())
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
// WithSiteTitle(title string) sets the site title used in generated HTML
// pages and templates.
//
// WithEnvironment(env string) sets the runtime environment ("local", "test",
// or "production") surfaced to templates via models.BaseData.Environment.
// Use config.EnvironmentConfig to read the value from the ENVIRONMENT env var.
//
// WithBlogRoot(path string) sets the root path for the blog when deploying
// at a subdirectory rather than domain root. For example, use "/blog/" when
// deploying at example.com/blog/. This ensures all generated links in templates
// use the correct base path. Default is "/" for root deployment.
//
// # Usage Examples
//
// Basic usage with a single option:
//
//	fsys := os.DirFS("posts/")
//	gen := generator.New(fsys, nil, config.WithRawOutput())
//
// Multiple options can be combined:
//
//	renderer, _ := generator.NewTemplateRenderer(templates.Default)
//	gen := generator.New(fsys, renderer,
//	    config.WithRawOutput(),
//	    config.WithSiteTitle("My Blog"),
//	)
//
// Configuring blog root for subdirectory deployment:
//
//	renderer, _ := generator.NewTemplateRenderer(templates.Default)
//	gen := generator.New(fsys, renderer,
//	    config.WithBlogRoot("/blog/"),
//	)
//
// # Concurrency
//
// Option values are safe to create and use concurrently. Configuration
// structs are safe to read concurrently once created, but should not be
// modified after construction.
package config

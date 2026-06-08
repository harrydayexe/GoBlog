// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

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
// WithDisableTags() disables all tag-related output. When enabled, the
// generator skips rendering tag pages and the tags index, post tag slices
// are cleared, the outputter skips the tags directory, and the HTTP server
// does not register /tags routes. Unlike WithRawOutput(), posts and the index
// page are still rendered with full templates.
//
// WithDisableReadingTime() disables estimated reading time on posts. When
// enabled, Post.ReadingTimeMinutes is left at zero and the default templates
// suppress the "· N min read" annotation next to each post date. Reading time
// is enabled by default (220 WPM, rounded up, 1-minute minimum).
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
// WithLogger(l *slog.Logger) is a BaseOption that sets the structured logger
// used by the receiving component. It flows into every constructor that
// accepts BaseOption values (generator.New via GeneratorOption, server.New via
// BaseServerOption, outputter.NewDirectoryWriter via GeneratorOption) and into
// watcher.New via WatcherOption (which embeds BaseOption). When not supplied,
// each constructor falls back to slog.Default() at construction time. Passing
// nil has the same effect as omitting the option.
//
// WithFuncs(funcs template.FuncMap) is a RendererOption that registers
// additional template functions for use in all templates. Functions are merged
// into the built-in FuncMap (formatDate, shortDate, year). A function whose
// name matches a built-in silently replaces it. Pass RendererOption values to
// generator.NewTemplateRenderer, or to ServerConfig.RendererOpts for the HTTP
// server path.
//
// WithHTMLPaths() is a GeneratorOption that switches BaseData.Path values to
// use .html file extensions instead of clean URLs. When enabled, the index
// path becomes /index.html (or /<root>.html for a non-root BlogRoot), and all
// other pages get .html appended. This is automatically applied by the goblog
// generate CLI so canonical URLs in static output match the actual files on
// disk. Library users serving via pkg/server should leave this off; the server
// accepts both clean URLs and .html URLs automatically.
//
// WithCustomData(data map[string]any) is a GeneratorOption that merges
// arbitrary key-value data into models.BaseData.Custom, making it accessible
// in all templates as {{.Custom.key}}. Multiple calls merge their maps, with
// later values overwriting earlier ones for duplicate keys. The field is nil
// when no WithCustomData option is supplied.
//
// # Option types
//
// GeneratorOption carries options for generator.New and outputter.NewDirectoryWriter,
// including WithRawOutput, WithDisableTags, WithDisableReadingTime, WithSiteTitle,
// WithEnvironment, WithCustomData, WithHTMLPaths, and (via the embedded BaseOption)
// WithLogger and WithBlogRoot.
// BaseServerOption carries options for the HTTP server (port, host, middleware,
// and via the embedded BaseOption: WithLogger, WithBlogRoot).
// WatcherOption carries options for watcher.New (debounce, and via the embedded
// BaseOption: WithLogger, WithBlogRoot).
// RendererOption carries options for generator.NewTemplateRenderer (custom funcs).
// ServerConfig groups all three option types plus a TemplateDir filesystem for
// the server constructor (server.New).
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
// Registering a custom template function:
//
//	renderer, _ := generator.NewTemplateRenderer(
//	    templates.Default,
//	    config.WithFuncs(template.FuncMap{"upper": strings.ToUpper}),
//	)
//
// Injecting custom data into templates:
//
//	gen := generator.New(fsys, renderer,
//	    config.WithCustomData(map[string]any{
//	        "author": "Jane Smith",
//	    }),
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

// Package config provides common configuration types and options used across
// the GoBlog generator and outputter packages.
//
// The package implements the functional options pattern for configuration,
// allowing packages to accept both required positional parameters and optional
// configuration via option functions.
//
// # CommonConfig
//
// CommonConfig contains configuration options that are shared between the
// generator and outputter packages. Components embed CommonConfig in their
// own configuration structs to inherit these shared options.
//
// # Usage with Generator
//
// The generator package uses CommonConfig for optional parameters:
//
//	postsFS := os.DirFS("posts/")
//	gen := generator.New(postsFS, config.WithRawOutput())
//
// # Usage with Outputter
//
// The outputter package also uses CommonConfig:
//
//	writer := outputter.NewDirectoryWriter("output/",
//	    config.WithRawOutput(),
//	)
//
// # Functional Options Pattern
//
// Options are implemented as functions that modify a CommonConfig:
//
//	type CommonOption func(*CommonConfig)
//
// This pattern allows for:
//   - Backward compatibility when adding new options
//   - Clear, self-documenting API calls
//   - Optional parameters without function overloading
//
// # Available Options
//
// WithRawOutput() - Generates raw HTML content without template wrapping.
// When enabled in the generator, markdown is converted to HTML without
// inserting it into page templates. When enabled in the outputter, the
// tags directory is not created.
//
// # Concurrency
//
// CommonConfig structs are safe to read concurrently once created. Option
// functions should not be called concurrently with reads.
package config

// Package generator provides functionality for generating static HTML blog sites
// from markdown files.
//
// The generator uses the functional options pattern for configuration. Required
// parameters are passed as positional arguments, while optional parameters are
// configured via option functions.
//
// TODO: Add usage examples
//
// The Generator returns all generated content in memory via GeneratedBlog.
// Callers are responsible for I/O operations such as writing
// files to disk or serving content via HTTP.
//
// The Generator is safe for concurrent use once created, though Generate
// operations should not be run concurrently on the same Generator instance.
package generator

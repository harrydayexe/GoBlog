// Package generator provides functionality for generating static HTML blog sites
// from markdown files.
//
// The generator uses the functional options pattern for configuration. Required
// parameters are passed as positional arguments, while optional parameters are
// configured via option functions.
//
// # Basic Usage
//
// Create a generator and generate a blog:
//
//	import (
//		"context"
//		"os"
//		"github.com/harrydayexe/GoBlog/v2/pkg/generator"
//		"github.com/harrydayexe/GoBlog/v2/pkg/config"
//	)
//
//	// Create generator with posts from a directory
//	fsys := os.DirFS("posts/")
//	gen := generator.New(fsys, nil, config.WithRawOutput())
//
//	// Generate the blog
//	blog, err := gen.Generate(context.Background())
//	if err != nil {
//		log.Fatal(err)
//	}
//
// # Output
//
// The Generator returns all generated content in memory via GeneratedBlog.
// Callers are responsible for I/O operations such as writing
// files to disk or serving content via HTTP.
//
// See the examples_test.go file for more usage examples.
//
// # Concurrency
//
// The Generator is safe for concurrent use once created, though Generate
// operations should not be run concurrently on the same Generator instance.
package generator

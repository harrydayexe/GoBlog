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
//	// Create a generator using the built-in templates:
//
//	import "github.com/harrydayexe/GoBlog/v2/pkg/templates"
//
//	fsys := os.DirFS("posts/")
//	renderer, err := generator.NewTemplateRenderer(templates.Default)
//	if err != nil {
//		log.Fatal(err)
//	}
//	gen := generator.New(fsys, renderer)
//
//	// Generate the blog
//	blog, err := gen.Generate(context.Background())
//	if err != nil {
//		log.Fatal(err)
//	}
//
// In raw output mode no templates are applied, so the renderer may be nil:
//
//	gen := generator.New(fsys, nil, config.WithRawOutput())
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

// Package outputter provides interfaces and implementations for handling
// generated blog content from the generator package.
//
// The Outputter interface defines how generated blog content should be
// processed. Implementations can write to disk, serve via HTTP, store in
// databases, or perform any other output operation.
//
// The package includes DirectoryWriter, a filesystem-based implementation
// that writes blog content as static HTML files to a specified directory.
//
// # Basic Usage
//
// Create a DirectoryWriter and use it to write generated blog content:
//
//	import "github.com/harrydayexe/GoBlog/v2/pkg/templates"
//
//	postsFS := os.DirFS("posts/")
//	renderer, err := generator.NewTemplateRenderer(templates.Default)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	gen := generator.New(postsFS, renderer)
//	blog, err := gen.Generate(context.Background())
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	writer := outputter.NewDirectoryWriter("output/")
//	if err := writer.HandleGeneratedBlog(context.Background(), blog); err != nil {
//	    log.Fatal(err)
//	}
//
// # Directory Structure
//
// DirectoryWriter creates the following structure:
//
//	output/
//	├── index.html           # Blog index page
//	├── post-slug-1.html     # Individual post pages
//	├── post-slug-2.html
//	└── tags/                # Tag pages (unless RawOutput is enabled)
//	    ├── tag-1.html
//	    └── tag-2.html
//
// # Configuration
//
// DirectoryWriter supports the functional options pattern for optional
// configuration:
//
//	writer := outputter.NewDirectoryWriter("output/",
//	    config.WithRawOutput(),
//	)
//
// When RawOutput is enabled, the tags directory is not created.
//
// # Concurrency
//
// Outputter implementations should be safe for concurrent use. DirectoryWriter
// is safe for concurrent calls to HandleGeneratedBlog, though concurrent writes
// to the same output directory may result in race conditions at the filesystem
// level.
package outputter

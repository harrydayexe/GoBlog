package outputter

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/harrydayexe/GoBlog/v2/pkg/config"
	"github.com/harrydayexe/GoBlog/v2/pkg/generator"
)

// DirectoryWriter is an Outputter implementation that writes blog content
// as static HTML files to a filesystem directory.
//
// It creates an index.html file, individual post HTML files, and (unless
// RawOutput is enabled) a tags subdirectory with tag pages and a tags index.
//
// DirectoryWriter is safe for concurrent use, though concurrent writes to
// the same output directory may result in filesystem race conditions.
type DirectoryWriter struct {
	config.RawOutput
	outputDir string
	logger    *slog.Logger
}

// NewDirectoryWriter creates a new DirectoryWriter with the specified output
// directory and optional configuration.
//
// The outputDir parameter specifies where blog files will be written. The
// directory will be created if it does not exist.
//
// Optional configuration can be provided via functional options from the
// config package:
//
//	writer := NewDirectoryWriter("/var/www/blog",
//	    config.WithRawOutput(),
//	)
//
// This is the recommended constructor for most use cases. Use
// NewDirectoryWriterWithConfig when you need to provide an explicit
// DirectoryWriterConfig struct.
func NewDirectoryWriter(outputDir string, opts ...config.Option) DirectoryWriter {
	logger := slog.Default()

	dw := DirectoryWriter{
		outputDir: outputDir,
		logger:    logger,
	}

	for _, opt := range opts {
		if opt.WithRawOutputFunc != nil {
			opt.WithRawOutputFunc(&dw.RawOutput)
		}
	}

	logger.Debug("Directory Writer created", slog.String("output directory", outputDir))

	return dw
}

// HandleGeneratedBlog writes the generated blog content to the filesystem
// as static HTML files.
//
// The method creates the following structure in the output directory:
//   - index.html: the main blog index page
//   - posts/{slug}.html: individual post files, one per post
//   - tags/{tag}.html: tag pages (only if RawOutput is false)
//   - tags/index.html: tags index page (only if RawOutput is false)
//
// When RawOutput mode is enabled (via config.WithRawOutput()), the tags/
// directory is not created and individual post files contain only raw HTML
// fragments without template wrappers. This is useful when you plan to
// wrap the content with your own templates or integrate it into an existing
// site structure.
//
// All necessary directories are created automatically with permissions 0755.
// Files are written with permissions 0644.
//
// Returns an error if:
//   - The output directory cannot be created
//   - Any file write operation fails (e.g., insufficient permissions)
//   - The filesystem is full or read-only
//
// If an error occurs partway through writing, some files may have been
// created successfully while others were not. The output directory may
// be in a partial state.
func (dw DirectoryWriter) HandleGeneratedBlog(ctx context.Context, blog *generator.GeneratedBlog) error {
	dw.logger.InfoContext(ctx, "Writing blog to directory")
	if err := writeMapToFiles(blog.Posts, filepath.Join(dw.outputDir, "posts")); err != nil {
		return err
	}

	// Always write index.html
	if err := os.WriteFile(filepath.Join(dw.outputDir, "index.html"), blog.Index, 0644); err != nil {
		return err
	}

	// Only write tags and tags index if NOT in RawOutput mode
	if !dw.RawOutput.RawOutput {
		if err := writeMapToFiles(blog.Tags, filepath.Join(dw.outputDir, "tags")); err != nil {
			return err
		}
		// Write tags index page if it has content
		if len(blog.TagsIndex) > 0 {
			if err := os.WriteFile(filepath.Join(dw.outputDir, "tags", "index.html"), blog.TagsIndex, 0644); err != nil {
				return err
			}
		}
	}

	dw.logger.InfoContext(ctx, "Finished writing to output directory")
	return nil
}

// writeMapToFiles writes a map of filename->content pairs to disk as HTML files.
//
// Each key in the data map becomes a filename with ".html" appended, and the
// corresponding byte slice is written as the file content.
//
// The outputDir is created if it doesn't exist, with permissions 0755.
// Files are written with permissions 0644.
//
// Returns an error if directory creation or any file write fails.
func writeMapToFiles(data map[string][]byte, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	for filename, content := range data {
		htmlFile := filename + ".html"
		path := filepath.Join(outputDir, htmlFile)
		if err := os.WriteFile(path, content, 0644); err != nil {
			return err
		}
	}
	return nil
}

package outputter

import (
	"os"
	"path/filepath"

	"github.com/harrydayexe/GoBlog/v2/pkg/config"
	"github.com/harrydayexe/GoBlog/v2/pkg/generator"
)

// DirectoryWriterConfig contains configuration for DirectoryWriter.
//
// It embeds config.CommonConfig to inherit shared options like RawOutput.
// OutputPath specifies the directory where HTML files will be written.
type DirectoryWriterConfig struct {
	config.CommonConfig
	// OutputPath is the directory where the blog files will be written.
	// The directory will be created if it does not exist.
	OutputPath string
}

// DirectoryWriter is an Outputter implementation that writes blog content
// as static HTML files to a filesystem directory.
//
// It creates an index.html file, individual post HTML files, and (unless
// RawOutput is enabled) a tags subdirectory with tag pages.
//
// DirectoryWriter is safe for concurrent use, though concurrent writes to
// the same output directory may result in filesystem race conditions.
type DirectoryWriter struct {
	config DirectoryWriterConfig
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
func NewDirectoryWriter(outputDir string, opts ...config.CommonOption) DirectoryWriter {
	config := DirectoryWriterConfig{
		OutputPath: outputDir,
	}

	for _, opt := range opts {
		opt(&config.CommonConfig)
	}

	return NewDirectoryWriterWithConfig(config)
}

// NewDirectoryWriterWithConfig creates a new DirectoryWriter with an explicit
// configuration struct.
//
// This constructor is useful when you need to programmatically build a
// DirectoryWriterConfig or when working with configuration systems that
// populate config structs.
//
// For most use cases, prefer NewDirectoryWriter which uses the functional
// options pattern.
func NewDirectoryWriterWithConfig(config DirectoryWriterConfig) DirectoryWriter {
	return DirectoryWriter{
		config: config,
	}
}

// HandleGeneratedBlog writes the generated blog content to the filesystem
// as static HTML files.
//
// The method creates the following structure in the output directory:
//   - index.html: the main blog index page
//   - {slug}.html: individual post files, one per post
//   - tags/{tag}.html: tag pages (only if RawOutput is false)
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
func (dw DirectoryWriter) HandleGeneratedBlog(blog *generator.GeneratedBlog) error {
	if err := writeMapToFiles(blog.Posts, dw.config.OutputPath); err != nil {
		return err
	}
	if !dw.config.RawOutput {
		if err := writeMapToFiles(blog.Tags, filepath.Join(dw.config.OutputPath, "tags")); err != nil {
			return err
		}
	}
	if err := os.WriteFile(filepath.Join(dw.config.OutputPath, "index.html"), blog.Index, 0644); err != nil {
		return err
	}

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

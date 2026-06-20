// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

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
// RawOutput or DisableTags is enabled) a tags subdirectory with tag pages
// and a tags index.
//
// DirectoryWriter is safe for concurrent use, though concurrent writes to
// the same output directory may result in filesystem race conditions.
type DirectoryWriter struct {
	config.RawOutput
	config.DisableTags
	config.Logger
	outputDir string
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
//	writer := NewDirectoryWriter("/var/www/blog",
//	    config.WithDisableTags(),
//	)
//
// This is the recommended constructor for most use cases.
func NewDirectoryWriter(outputDir string, opts ...config.GeneratorOption) DirectoryWriter {
	dw := DirectoryWriter{
		outputDir: outputDir,
	}

	for _, opt := range opts {
		if opt.WithRawOutputFunc != nil {
			opt.WithRawOutputFunc(&dw.RawOutput)
		} else if opt.WithDisableTagsFunc != nil {
			opt.WithDisableTagsFunc(&dw.DisableTags)
		} else if opt.WithLoggerFunc != nil {
			opt.WithLoggerFunc(&dw.Logger)
		}
	}

	if dw.Logger.Logger == nil {
		dw.Logger.Logger = slog.Default()
	}

	dw.Logger.Logger.Debug("Directory Writer created", slog.String("output directory", outputDir))

	return dw
}

// HandleGeneratedBlog writes the generated blog content to the filesystem
// as static HTML files.
//
// The method creates the following structure in the output directory:
//   - index.html: the main blog index page
//   - posts/{slug}.html: individual post files, one per post
//   - tags/{tag}.html: tag pages (only if RawOutput and DisableTags are false)
//   - tags/index.html: tags index page (only if RawOutput and DisableTags are false)
//   - rss.xml: site-wide RSS 2.0 feed (only when the generator produced one)
//   - atom.xml: site-wide Atom feed (only when the generator produced one)
//   - tags/{tag}.rss.xml: per-tag RSS 2.0 feed (only when the generator produced them)
//   - tags/{tag}.atom.xml: per-tag Atom feed (only when the generator produced them)
//
// When RawOutput mode is enabled (via config.WithRawOutput()), the tags/
// directory is not created and individual post files contain only raw HTML
// fragments without template wrappers. This is useful when you plan to
// wrap the content with your own templates or integrate it into an existing
// site structure.
//
// When DisableTags mode is enabled (via config.WithDisableTags()), the tags/
// directory is not created. Posts and the index page are still written with
// full templates.
//
// Feed files are written only when the generator has populated them (i.e. when
// config.WithBaseURL was set and config.WithDisableFeeds was not applied). The
// outputter does not need to know about feed configuration — it simply writes
// whatever the generator produced.
//
// All necessary directories are created automatically with permissions 0755.
// Files are written with permissions 0644.
//
// Returns an error if:
//   - The output directory cannot be created
//   - Any file write operation fails (e.g., insufficient permissions)
//   - The filesystem is full or read-only
//
// # Partial Write Behavior
//
// If an error occurs partway through writing, some files may have been
// created successfully while others were not. The output directory will be
// left in a partial state and may require manual cleanup. To ensure atomic
// writes, consider writing to a temporary directory first and renaming it
// on success.
func (dw DirectoryWriter) HandleGeneratedBlog(ctx context.Context, blog *generator.GeneratedBlog) error {
	dw.Logger.Logger.InfoContext(ctx, "Writing blog to directory")
	if err := writeMapToFiles(blog.Posts, filepath.Join(dw.outputDir, "posts")); err != nil {
		return err
	}

	// Always write index.html
	if err := os.WriteFile(filepath.Join(dw.outputDir, "index.html"), blog.Index, 0644); err != nil {
		return err
	}

	// Only write tags and tags index if NOT in RawOutput or DisableTags mode
	if !dw.RawOutput.RawOutput && !dw.DisableTags.Disable {
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

	// Write site-wide feed files when the generator produced them.
	if len(blog.RSSFeed) > 0 {
		if err := os.WriteFile(filepath.Join(dw.outputDir, "rss.xml"), blog.RSSFeed, 0644); err != nil {
			return err
		}
	}
	if len(blog.AtomFeed) > 0 {
		if err := os.WriteFile(filepath.Join(dw.outputDir, "atom.xml"), blog.AtomFeed, 0644); err != nil {
			return err
		}
	}

	// Write per-tag feed files when the generator produced them.
	// The tags directory is guaranteed to exist at this point if tag feeds were generated
	// (tags are required for tag feeds). We still guard with MkdirAll for safety.
	if len(blog.TagRSSFeeds) > 0 || len(blog.TagAtomFeeds) > 0 {
		if err := writeMapToFilesExt(blog.TagRSSFeeds, filepath.Join(dw.outputDir, "tags"), ".rss.xml"); err != nil {
			return err
		}
		if err := writeMapToFilesExt(blog.TagAtomFeeds, filepath.Join(dw.outputDir, "tags"), ".atom.xml"); err != nil {
			return err
		}
	}

	dw.Logger.Logger.InfoContext(ctx, "Finished writing to output directory")
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
	return writeMapToFilesExt(data, outputDir, ".html")
}

// writeMapToFilesExt writes a map of filename->content pairs to disk, appending
// the given extension to each key to form the on-disk filename.
//
// For example, with ext=".rss.xml" and key "golang", the file is written as
// "golang.rss.xml" inside outputDir.
//
// The outputDir is created if it doesn't exist, with permissions 0755.
// Files are written with permissions 0644.
//
// Returns an error if directory creation or any file write fails.
func writeMapToFilesExt(data map[string][]byte, outputDir string, ext string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	for filename, content := range data {
		path := filepath.Join(outputDir, filename+ext)
		if err := os.WriteFile(path, content, 0644); err != nil {
			return err
		}
	}
	return nil
}

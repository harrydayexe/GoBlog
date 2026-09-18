// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package outputter

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/harrydayexe/GoBlog/v2/pkg/config"
	"github.com/harrydayexe/GoBlog/v2/pkg/generator"
)

// DirectoryWriter is an Outputter implementation that writes blog content
// as static HTML files to a filesystem directory.
//
// It creates an index.html file, individual post HTML files, (unless
// RawOutput or DisableTags is enabled) a tags subdirectory with tag pages
// and a tags index, and (when an assets directory is configured) an images
// subdirectory containing a copy of the assets.
//
// DirectoryWriter is safe for concurrent use, though concurrent writes to
// the same output directory may result in filesystem race conditions.
type DirectoryWriter struct {
	config.RawOutput
	config.DisableTags
	config.Logger
	config.AssetsDir
	config.BlogRoot
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
//	writer := NewDirectoryWriter("/var/www/blog",
//	    config.WithAssetsDir(assetsFS).AsGeneratorOption(),
//	)
//	writer := NewDirectoryWriter("/var/www/blog",
//	    config.WithBlogRoot("/blog/").AsGeneratorOption(),
//	)
//
// The blog root does not change where files are written — the output tree is
// always flat. It is used only to warn that robots.txt will need relocating
// to the domain root on deploy; see HandleGeneratedBlog.
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
		} else if opt.WithAssetsDirFunc != nil {
			opt.WithAssetsDirFunc(&dw.AssetsDir)
		} else if opt.WithBlogRootFunc != nil {
			opt.WithBlogRootFunc(&dw.BlogRoot)
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
//   - sitemap.xml: the sitemap (only when the generator produced one)
//   - robots.txt: the robots file (only when the generator produced one)
//   - images/...: a recursive copy of the assets directory (only when
//     config.WithAssetsDir was supplied and its root is a readable directory)
//
// Assets are copied in both templated and RawOutput mode, since rendered image
// src attributes point at {BlogRoot}images/ either way. Existing files in the
// images directory are overwritten. Entries that are neither regular files,
// directories, nor symbolic links resolving to regular files inside the assets
// filesystem are skipped with a warning.
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
// whatever the generator produced. sitemap.xml and robots.txt follow the same
// rule.
//
// # robots.txt Placement
//
// The output tree is flat and the blog root affects only link URLs, so with a
// blog root of "/blog/" the output directory *is* the deployed /blog/
// directory and the domain root is its parent — outside this writer's reach.
// robots.txt is therefore always written to <outputDir>/robots.txt, and a
// warning is logged when the blog root is not "/": crawlers only ever fetch
// /robots.txt at the domain root, so the file must be relocated there on
// deploy. sitemap.xml has no such problem — a sitemap at a sub-path
// legitimately covers URLs at or below that sub-path.
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

	// Write the sitemap and robots file when the generator produced them.
	if len(blog.Sitemap) > 0 {
		if err := os.WriteFile(filepath.Join(dw.outputDir, "sitemap.xml"), blog.Sitemap, 0644); err != nil {
			return err
		}
	}
	if len(blog.RobotsTxt) > 0 {
		if err := os.WriteFile(filepath.Join(dw.outputDir, "robots.txt"), blog.RobotsTxt, 0644); err != nil {
			return err
		}
		// The output tree is flat, so robots.txt lands inside the deployed
		// blog root rather than at the domain root where crawlers look for it.
		if root := string(dw.BlogRoot); root != "" && root != "/" {
			dw.Logger.Logger.WarnContext(ctx,
				"robots.txt written inside the blog root; crawlers only read /robots.txt at the domain root, so move it there on deploy",
				slog.String("written", filepath.Join(dw.outputDir, "robots.txt")),
				slog.String("blogRoot", root),
			)
		}
	}

	if dw.AssetsDir.Enabled() {
		if err := dw.copyAssets(ctx, filepath.Join(dw.outputDir, "images")); err != nil {
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

// copyAssets recursively copies the assets filesystem into outputDir,
// preserving subdirectories.
//
// Files are read through the assets filesystem, so a filesystem rooted with
// [os.Root.FS] cannot be used to copy files from outside the assets directory
// via symbolic links. Such links, and any other non-regular entries, are
// skipped with a warning.
//
// Directories are created with permissions 0755 and files written with 0644.
func (dw DirectoryWriter) copyAssets(ctx context.Context, outputDir string) error {
	dw.Logger.Logger.InfoContext(ctx, "Copying assets", slog.String("destination", outputDir))
	fsys := dw.AssetsDir.FS

	return fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("failed to read assets: %w", err)
		}
		dest := filepath.Join(outputDir, filepath.FromSlash(p))

		if d.IsDir() {
			return os.MkdirAll(dest, 0755)
		}

		if !d.Type().IsRegular() {
			fi, err := fs.Stat(fsys, p)
			if err != nil || !fi.Mode().IsRegular() {
				dw.Logger.Logger.WarnContext(ctx, "Skipping asset that is not a regular file", slog.String("path", p))
				return nil
			}
		}

		data, err := fs.ReadFile(fsys, p)
		if err != nil {
			return fmt.Errorf("failed to read asset %q: %w", p, err)
		}
		return os.WriteFile(dest, data, 0644)
	})
}

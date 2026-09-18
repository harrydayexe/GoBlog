// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package parser

import (
	"io/fs"
	"log/slog"
)

// Option is a function which can update the parser config
type Option func(*Config)

// WithLogger returns an Option that sets the structured logger used by the
// parser. When not supplied, the parser falls back to [log/slog.Default] at
// construction time.
//
// Example usage:
//
//	p := parser.New(parser.WithLogger(myLogger))
func WithLogger(l *slog.Logger) Option {
	return func(c *Config) {
		c.Logger = l
	}
}

// WithCodeHighlighting enables or disables the highlighting of ringfenced
// code blocks. Highlighting is enabled by default.
//
// When enabled, the parser outputs CSS class names (e.g. .chroma, .k, .s)
// rather than inline styles. A matching chroma stylesheet must be included
// in your page — see the package documentation for how to generate one with
// chromahtml.WriteCSS.
func WithCodeHighlighting(enable bool) Option {
	return func(c *Config) {
		c.EnableCodeHighlighting = enable
	}
}

// WithFootnote enables the use of PHP Markdown Extra Footnotes.
//
// Footnotes allow you to add references and notes without cluttering the main text.
// Use [^1] in your text and define footnotes with [^1]: Your footnote text.
//
// See the PHP Markdown Extra documentation for syntax details:
// https://michelf.ca/projects/php-markdown/extra/#footnotes
func WithFootnote() Option {
	return func(c *Config) {
		c.EnableFootnote = true
	}
}

// WithBlogRoot sets the root path the blog is served under, such as "/" or
// "/blog/". It is used to rewrite relative image paths to root-absolute asset
// URLs: ![alt](images/foo.png), ![alt](foo.png) and ![[foo.png]] all render
// with src="{BlogRoot}images/foo.png". The default is "/".
//
// Example usage:
//
//	p := parser.New(parser.WithBlogRoot("/blog/"))
func WithBlogRoot(root string) Option {
	return func(c *Config) {
		c.BlogRoot = root
	}
}

// WithAssetsDir sets the filesystem the blog's images live in, the same
// directory served and copied by [github.com/harrydayexe/GoBlog/v2/pkg/config.WithAssetsDir].
//
// The parser reads each referenced image's header from it to emit width and
// height attributes, so browsers can reserve space for the image and avoid
// layout shift. Nothing is written, and only the header of each file is read.
// Images are still measured at most once per parser, no matter how many posts
// reference them.
//
// The option is entirely optional: when it is not supplied, or a file is
// missing, unreadable, or in a format whose header cannot be decoded (SVG,
// AVIF and WebP), the image simply renders without dimensions rather than
// failing the parse.
//
// Prefer a filesystem that cannot escape its root, such as the one returned by
// [os.Root.FS]; [os.DirFS] follows symbolic links that point outside the
// directory.
//
// Example usage:
//
//	root, err := os.OpenRoot("posts/images")
//	if err != nil {
//	    return err
//	}
//	defer root.Close()
//
//	p := parser.New(parser.WithAssetsDir(root.FS()))
func WithAssetsDir(fsys fs.FS) Option {
	return func(c *Config) {
		c.AssetsDir = fsys
	}
}

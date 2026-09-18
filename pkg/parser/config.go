// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package parser

import (
	"io/fs"
	"log/slog"
)

// Config contains all the options for the Parser to use when reading and
// parsing markdown files.
type Config struct {
	// EnableCodeHighlighting controls whether ringfenced code blocks should be
	// highlighted or not
	EnableCodeHighlighting bool

	// EnableFootnote controls whether the parser should allow the use of PHP
	// Markdown Extra Footnotes.
	EnableFootnote bool

	// Logger is the structured logger used by the parser. When nil,
	// [log/slog.Default] is used.
	Logger *slog.Logger

	// BlogRoot is the root path the blog is served under, such as "/" or
	// "/blog/". Relative image paths are rewritten to "{BlogRoot}images/...".
	// An empty BlogRoot is treated as "/".
	BlogRoot string

	// AssetsDir is the filesystem holding the blog's images. It is read, and
	// never written, to measure the intrinsic width and height of every image
	// a post references. When nil, images are rendered without dimensions.
	AssetsDir fs.FS
}

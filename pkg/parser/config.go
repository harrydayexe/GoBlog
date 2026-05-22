// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package parser

// Config contains all the options for the Parser to use when reading and
// parsing markdown files.
type Config struct {
	// EnableCodeHighlighting controls whether ringfenced code blocks should be
	// highlighted or not
	EnableCodeHighlighting bool

	// EnableFootnote controls whether the parser should allow the use of PHP
	// Markdown Extra Footnotes.
	EnableFootnote bool
}

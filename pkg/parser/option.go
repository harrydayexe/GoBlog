package parser

import "github.com/harrydayexe/GoBlog/internal/parser"

// Option is a function which can update the parser config
type Option func(*parser.Config)

// WithCodeHighlightingStyle sets the code highlighting style for ringfenced
// code blocks and enables code highlighting if it wasn't already
//
// Note that currently Chroma is used under-the-hood, so you should refer to the styles in
// their [documentation] and [source code]
//
// [documentation]: https://pkg.go.dev/github.com/alecthomas/chroma
// [source code]: https://github.com/alecthomas/chroma/tree/master/styles
func WithCodeHighlightingStyle(style string) Option {
	return func(c *parser.Config) {
		c.EnableCodeHighlighting = true
		c.CodeHighlightingStyle = style
	}
}

// WithCodeHighlighting enables the highlighting of ringfenced code blocks
//
// Note that currently Chroma is used under-the-hood, so you should refer to the styles in
// their [documentation] and [source code]
//
// [documentation]: https://pkg.go.dev/github.com/alecthomas/chroma
// [source code]: https://github.com/alecthomas/chroma/
func WithCodeHighlighting(enable bool) Option {
	return func(c *parser.Config) {
		c.EnableCodeHighlighting = enable
	}
}

// WithFootnote enables the use of PHP Markdown Extra Footnotes.
func WithFootnote() Option {
	return func(c *parser.Config) {
		c.EnableFootnote = true
	}
}

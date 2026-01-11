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
//
// Footnotes allow you to add references and notes without cluttering the main text.
// Use [^1] in your text and define footnotes with [^1]: Your footnote text.
//
// See the PHP Markdown Extra documentation for syntax details:
// https://michelf.ca/projects/php-markdown/extra/#footnotes
func WithFootnote() Option {
	return func(c *parser.Config) {
		c.EnableFootnote = true
	}
}

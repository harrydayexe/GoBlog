package parser

// Option is a function which can update the parser config
type Option func(*Config)

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

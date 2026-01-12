package parser

// Config contains all the options for the Parser to use when reading and
// parsing markdown files.
type Config struct {
	// EnableCodeHighlighting controls whether ringfenced code blocks should be
	// highlighted or not
	EnableCodeHighlighting bool
	// CodeHighlightingStyle to use for ringfenced code blocks.
	//
	// Note that currently Chroma is used so you should refer to the styles in
	// their [documentation] and [source code]
	//
	// [documentation]: https://pkg.go.dev/github.com/alecthomas/chroma
	// [source code]: https://github.com/alecthomas/chroma/tree/master/styles
	CodeHighlightingStyle string

	// EnableFootnote controls whether the parser should allow the use of PHP
	// Markdown Extra Footnotes.
	EnableFootnote bool
}

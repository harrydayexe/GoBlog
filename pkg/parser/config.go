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

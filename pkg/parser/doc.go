// Package parser reads markdown files with YAML frontmatter and converts them
// to Post objects for the GoBlog system.
//
// The parser uses goldmark for markdown processing and supports:
//   - YAML frontmatter for post metadata
//   - Syntax highlighting for code blocks
//   - Footnotes
//   - Auto-generated heading IDs
//   - HTML sanitization
//
// Basic usage:
//
//	import (
//		"os"
//		"github.com/harrydayexe/GoBlog/pkg/parser"
//	)
//
//	// Parse a single file
//	p := parser.New()
//	post, err := p.ParseFile(os.DirFS("/path/to/posts"), "my-post.md")
//	if err != nil {
//		// handle error
//	}
//
//	// Parse all markdown files in a directory
//	posts, err := p.ParseDirectory(os.DirFS("/path/to/posts"))
//	if err != nil {
//		// handle error - may be ParseErrors with partial results
//	}
//
// # Configuration
//
// The parser can be customized using functional options:
//
//	// Disable code highlighting
//	p := parser.New(parser.WithCodeHighlighting(false))
//
//	// Change syntax highlighting style
//	p := parser.New(parser.WithCodeHighlightingStyle("dracula"))
//
//	// Enable footnote support
//	p := parser.New(parser.WithFootnote())
//
//	// Combine multiple options
//	p := parser.New(
//	    parser.WithCodeHighlightingStyle("monokai"),
//	    parser.WithFootnote(),
//	)
//
// See the Option functions for all available configuration options.
//
// A Parser is safe for concurrent use by multiple goroutines after creation.
package parser

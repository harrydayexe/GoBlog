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
//		"context"
//		"os"
//		"github.com/harrydayexe/GoBlog/v2/pkg/parser"
//	)
//
//	// Parse a single file
//	p := parser.New()
//	post, err := p.ParseFile(context.Background(), os.DirFS("/path/to/posts"), "my-post.md")
//	if err != nil {
//		// handle error
//	}
//
//	// Parse all markdown files in a directory
//	posts, err := p.ParseDirectory(context.Background(), os.DirFS("/path/to/posts"))
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
//	// Enable footnote support
//	p := parser.New(parser.WithFootnote())
//
//	// Combine multiple options
//	p := parser.New(
//	    parser.WithCodeHighlighting(true),
//	    parser.WithFootnote(),
//	)
//
// See the Option functions for all available configuration options.
//
// # Syntax Highlighting CSS
//
// The parser renders highlighted code blocks using CSS classes (via
// chroma's html.WithClasses option) rather than inline styles. This means
// the generated HTML will contain class names like .chroma, .k, .s, etc.,
// but will not be visually styled until a matching stylesheet is included
// in your page.
//
// Generate the stylesheet for a given style at startup and embed it in your
// templates:
//
//	import (
//	    chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
//	    "github.com/alecthomas/chroma/v2/styles"
//	    "strings"
//	)
//
//	formatter := chromahtml.New(chromahtml.WithClasses(true))
//	style := styles.Get("monokai")
//	var sb strings.Builder
//	formatter.WriteCSS(&sb, style)
//	chromaCSS := sb.String() // embed in a <style> tag in your template
//
// The class names follow the Pygments short-name convention. A full reference
// is available in the chroma source:
// https://github.com/alecthomas/chroma/blob/master/types.go
//
// A Parser is safe for concurrent use by multiple goroutines after creation.
package parser

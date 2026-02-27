package parser_test

import (
	"context"
	"fmt"
	"os"

	. "github.com/harrydayexe/GoBlog/v2/pkg/parser"
)

// Example demonstrates basic usage of the parser with default configuration.
// The parser reads markdown files with YAML frontmatter and converts them to Post objects.
func Example() {
	p := New()

	fsys := os.DirFS("testdata")
	post, err := p.ParseFile(context.Background(), fsys, "valid-post.md")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println(post.Title)
	// Output: A Valid Blog Post
}

// Example_withOptions demonstrates configuring the parser with custom options
// such as syntax highlighting styles and footnote support.
func Example_withOptions() {
	// Configure parser with custom options
	p := New(
		WithCodeHighlightingStyle("dracula"),
		WithFootnote(),
	)

	fsys := os.DirFS("testdata")
	post, err := p.ParseFile(context.Background(), fsys, "with-code.md")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Configured parser")
	fmt.Println("Has content:", len(post.Content) > 0)
	// Output: Configured parser
	// Has content: true
}

// ExampleParser_ParseFile demonstrates parsing a single markdown file.
// The file must contain YAML frontmatter with required fields (title, date, description).
func ExampleParser_ParseFile() {
	p := New()
	fsys := os.DirFS("testdata")

	post, err := p.ParseFile(context.Background(), fsys, "valid-post.md")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Title: %s\n", post.Title)
	fmt.Printf("Tags: %d\n", len(post.Tags))
	// Output: Title: A Valid Blog Post
	// Tags: 3
}

// ExampleParser_ParseDirectory demonstrates parsing all markdown files in a directory.
// When errors occur during parsing, partial results may still be returned along with
// a ParseErrors error containing details about which files failed.
func ExampleParser_ParseDirectory() {
	p := New(WithFootnote())
	fsys := os.DirFS("testdata")

	posts, err := p.ParseDirectory(context.Background(), fsys)

	// Partial results may be returned with errors
	if err != nil {
		if parseErrs, ok := err.(ParseErrors); ok {
			fmt.Printf("Parsed with some errors\n")
			fmt.Printf("Valid posts: %d\n", len(posts))
			fmt.Printf("Errors: %d\n", len(parseErrs.Errors))
			return
		}
	}

	fmt.Printf("Valid posts: %d\n", len(posts))
	// Output: Parsed with some errors
	// Valid posts: 3
	// Errors: 4
}

// ExampleWithCodeHighlighting demonstrates disabling syntax highlighting for code blocks.
// By default, code highlighting is enabled with the monokai style.
func ExampleWithCodeHighlighting() {
	// Disable code highlighting
	p := New(WithCodeHighlighting(false))

	fsys := os.DirFS("testdata")
	post, err := p.ParseFile(context.Background(), fsys, "with-code.md")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Highlighting disabled")
	fmt.Println("Post parsed:", post.Title != "")
	// Output: Highlighting disabled
	// Post parsed: true
}

// ExampleWithCodeHighlightingStyle demonstrates using a custom syntax highlighting style.
// The style names come from the Chroma library. Common styles include:
// monokai, dracula, github, solarized-dark, solarized-light, and many others.
func ExampleWithCodeHighlightingStyle() {
	// Use a different syntax highlighting style
	p := New(WithCodeHighlightingStyle("dracula"))

	fsys := os.DirFS("testdata")
	post, err := p.ParseFile(context.Background(), fsys, "with-code.md")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Custom highlighting style")
	fmt.Println("Post parsed:", post.Title != "")
	// Output: Custom highlighting style
	// Post parsed: true
}

// ExampleWithFootnote demonstrates enabling PHP Markdown Extra footnote support.
// Footnotes allow you to add references and notes without cluttering the main text.
// Use [^1] in your text and define footnotes with [^1]: Your footnote text.
func ExampleWithFootnote() {
	// Enable PHP Markdown Extra footnotes
	p := New(WithFootnote())

	fsys := os.DirFS("testdata")
	post, err := p.ParseFile(context.Background(), fsys, "with-footnotes.md")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Footnotes enabled")
	fmt.Println("Post parsed:", post.Title != "")
	// Output: Footnotes enabled
	// Post parsed: true
}

package parser

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/harrydayexe/GoBlog/pkg/models"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	goldmarkhtml "github.com/yuin/goldmark/renderer/html"
	"go.abhg.dev/goldmark/frontmatter"
)

// Parser reads markdown files and converts them to Post objects.
// A Parser is safe for concurrent use after creation.
type Parser struct {
	md goldmark.Markdown
}

// New creates a new Parser with the specified options.
// The parser is configured with:
// - YAML frontmatter parsing
// - Syntax highlighting for code blocks
// - Footnote support
// - Auto-generated heading IDs
// - HTML sanitization (unsafe HTML disabled by default)
func New() *Parser {
	// Configure goldmark with extensions
	md := goldmark.New(
		goldmark.WithExtensions(
			&frontmatter.Extender{},
			extension.Footnote,
			highlighting.NewHighlighting(
				highlighting.WithStyle("monokai"),
				highlighting.WithFormatOptions(
					html.WithClasses(true),
				),
			),
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			goldmarkhtml.WithHardWraps(),
			goldmarkhtml.WithXHTML(),
		),
	)

	p := &Parser{
		md: md,
	}

	return p
}

// ParseFile reads and parses a single markdown file from the filesystem.
// It extracts frontmatter metadata, validates required fields, renders the
// markdown content to HTML, and returns a fully populated Post.
//
// The file must contain YAML frontmatter with at minimum: title, date, and
// description. The markdown body is rendered to HTML with syntax highlighting
// and footnote support.
//
// Returns an error if the file cannot be read, frontmatter is invalid,
// required fields are missing, or markdown rendering fails.
func (p *Parser) ParseFile(fsys fs.FS, path string) (*models.Post, error) {
	// Read file contents
	content, err := fs.ReadFile(fsys, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Create parser context
	ctx := parser.NewContext()

	// Parse markdown (this also extracts frontmatter via the extension)
	var htmlBuf bytes.Buffer
	if err := p.md.Convert(content, &htmlBuf, parser.WithContext(ctx)); err != nil {
		return nil, fmt.Errorf("failed to render markdown: %w", err)
	}

	// Extract frontmatter from context
	var post models.Post
	fmData := frontmatter.Get(ctx)
	if fmData == nil {
		return nil, fmt.Errorf("no frontmatter found in file")
	}

	if err := fmData.Decode(&post); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	// Set source path before validation so error messages include it
	post.SourcePath = path

	// Validate required fields
	if err := post.Validate(); err != nil {
		return nil, err
	}

	// Store raw markdown content (without frontmatter)
	// We need to extract just the body content
	post.RawContent = string(content)

	// Store rendered HTML
	post.Content = htmlBuf.String()
	post.HTMLContent = template.HTML(post.Content)

	// Generate slug from title or filename
	post.GenerateSlug()

	return &post, nil
}

// ParseDirectory walks the filesystem and parses all .md files found.
// It returns a PostList containing all successfully parsed posts, sorted by
// date (newest first).
//
// If any files fail to parse, ParseDirectory continues processing remaining
// files and returns both the valid posts and a ParseErrors containing all
// failures. Callers can check the error and decide whether to proceed with
// partial results or fail entirely.
//
// Only files with .md or .markdown extensions are processed. Other files
// and directories are silently skipped.
func (p *Parser) ParseDirectory(fsys fs.FS) (models.PostList, error) {
	var posts models.PostList
	var parseErrors ParseErrors

	// Walk the filesystem and collect all .md files
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Error accessing path - collect but continue
			parseErrors.Errors = append(parseErrors.Errors, FileError{
				Path: path,
				Err:  err,
			})
			return nil
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Only process markdown files
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".md" && ext != ".markdown" {
			return nil
		}

		// Parse the file
		post, err := p.ParseFile(fsys, path)
		if err != nil {
			// Parsing failed - collect error but continue
			parseErrors.Errors = append(parseErrors.Errors, FileError{
				Path: path,
				Err:  err,
			})
			return nil
		}

		// Successfully parsed - add to collection
		posts = append(posts, post)
		return nil
	})

	// If WalkDir itself failed, return that error
	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	// Sort posts by date (newest first)
	posts.SortByDate()

	// Return posts and errors (if any)
	if parseErrors.HasErrors() {
		return posts, parseErrors
	}

	return posts, nil
}

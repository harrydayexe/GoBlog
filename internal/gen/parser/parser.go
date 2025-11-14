package parser

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	gmhtml "github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"go.abhg.dev/goldmark/frontmatter"

	"github.com/harrydayexe/GoBlog/internal/gen/log"
	"github.com/harrydayexe/GoBlog/pkg/models"
)

// Parser handles markdown parsing and HTML conversion
type Parser struct {
	md     goldmark.Markdown
	logger log.CLILogger
}

// New creates a new parser with goldmark configured for blog posts
func New() *Parser {
	logger := *log.NewCLILogger("PARSER", false)

	// Configure syntax highlighting
	highlighter := highlighting.NewHighlighting(
		highlighting.WithStyle("monokai"),
		highlighting.WithFormatOptions(
			html.WithLineNumbers(true),
			html.WithLinkableLineNumbers(true, ""),
		),
	)

	// Create goldmark instance with extensions
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,         // GitHub Flavored Markdown
			extension.Typographer, // Smart quotes, dashes, etc.
			highlighter,           // Syntax highlighting
			&frontmatter.Extender{}, // Frontmatter support
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(), // Auto-generate heading IDs
		),
		goldmark.WithRendererOptions(
			gmhtml.WithHardWraps(),     // Convert line breaks to <br>
			gmhtml.WithXHTML(),         // Use XHTML-style self-closing tags
			gmhtml.WithUnsafe(),        // Allow raw HTML in markdown
		),
	)

	return &Parser{
		md:     md,
		logger: logger,
	}
}

// ParseFile reads a markdown file and parses it into a Post
func (p *Parser) ParseFile(filePath string) (*models.Post, error) {
	// Read the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	p.logger.Debug("Parsing file: %s", filePath)

	// Parse the markdown and extract frontmatter
	post, err := p.ParseMarkdown(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse markdown in %s: %w", filePath, err)
	}

	// Set the source path
	post.SourcePath = filePath

	// Generate slug if not provided
	post.GenerateSlug()

	// Validate the post
	if err := post.Validate(); err != nil {
		return nil, err
	}

	p.logger.Debug("Successfully parsed: %s (slug: %s)", post.Title, post.Slug)
	return post, nil
}

// ParseMarkdown parses markdown content with frontmatter into a Post
func (p *Parser) ParseMarkdown(content []byte) (*models.Post, error) {
	// Create parser context
	ctx := parser.NewContext()

	// Parse the document
	doc := p.md.Parser().Parse(text.NewReader(content), parser.WithContext(ctx))

	// Extract frontmatter from context
	post := &models.Post{}
	fm := frontmatter.Get(ctx)
	if fm != nil {
		if err := fm.Decode(post); err != nil {
			return nil, fmt.Errorf("failed to decode frontmatter: %w", err)
		}
	}

	// Get the markdown content
	post.RawContent = string(content)

	// Convert markdown to HTML
	var buf bytes.Buffer
	if err := p.md.Renderer().Render(&buf, content, doc); err != nil {
		return nil, fmt.Errorf("failed to render markdown: %w", err)
	}
	htmlContent := buf.String()
	post.Content = htmlContent
	post.HTMLContent = template.HTML(htmlContent)

	return post, nil
}

// discoverPosts finds all markdown files in a directory recursively
func (p *Parser) discoverPosts(inputDir string) ([]string, error) {
	var markdownFiles []string

	p.logger.Info("Discovering markdown files in: %s", inputDir)

	err := filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if it's a markdown file
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".md" || ext == ".markdown" {
			markdownFiles = append(markdownFiles, path)
			p.logger.Debug("Found: %s", path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory %s: %w", inputDir, err)
	}

	p.logger.Info("Found %d markdown file(s)", len(markdownFiles))
	return markdownFiles, nil
}

// ParseAllPosts discovers and parses all posts in a directory
func (p *Parser) ParseAllPosts(inputDir string) (models.PostList, error) {
	// Discover all markdown files
	files, err := p.discoverPosts(inputDir)
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no markdown files found in %s", inputDir)
	}

	// Parse each file
	var posts models.PostList
	var errors []string

	for _, file := range files {
		post, err := p.ParseFile(file)
		if err != nil {
			p.logger.Warn("Skipping %s: %v", file, err)
			errors = append(errors, fmt.Sprintf("%s: %v", file, err))
			continue
		}
		posts = append(posts, post)
	}

	// Log summary
	p.logger.Info("Successfully parsed %d post(s)", len(posts))
	if len(errors) > 0 {
		p.logger.Warn("Failed to parse %d file(s)", len(errors))
	}

	if len(posts) == 0 {
		return nil, fmt.Errorf("no valid posts found (all files had errors)")
	}

	// Sort posts by date (newest first)
	posts.SortByDate()

	return posts, nil
}

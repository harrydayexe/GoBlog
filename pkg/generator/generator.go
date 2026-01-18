package generator

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"

	"github.com/harrydayexe/GoBlog/v2/pkg/config"
	"github.com/harrydayexe/GoBlog/v2/pkg/models"
	"github.com/harrydayexe/GoBlog/v2/pkg/parser"
)

// Generator produces HTML output based on its input configuration.
// It reads markdown files from a configured filesystem and renders them
// as HTML using templates.
//
// A Generator is safe for concurrent use after creation, but Generate
// operations should not be called concurrently on the same instance.
type Generator struct {
	config GeneratorConfig
	logger *slog.Logger
}

// New creates a new Generator with the specified options.
// It returns an error if the configuration is invalid or if required
// resources cannot be initialized.
//
// Options can be provided to customize behavior such as template directories,
// posts per page, and other generation parameters.
func New(posts fs.FS, opts ...config.CommonOption) *Generator {

	// TODO: Validate posts input somehow?
	config := GeneratorConfig{
		PostsDir: posts,
	}

	// Run options on config
	for _, opt := range opts {
		opt(&config.CommonConfig)
	}

	return NewWithConfig(config)
}

func NewWithConfig(config GeneratorConfig) *Generator {
	logger := slog.Default()

	gen := Generator{
		config: config,
		logger: logger,
	}

	return &gen
}

// Generate reads markdown post files from the configured filesystem and
// generates a complete static blog site as HTML.
//
// It returns a GeneratedBlog containing all rendered HTML content including
// individual post pages, tag pages, and the index page. The returned content
// is in-memory only; callers are responsible for writing to disk or serving
// via HTTP as needed.
//
// Generate respects the provided context and will return early with
// context.Canceled or context.DeadlineExceeded if the context is canceled
// or times out.
//
// It returns an error if markdown files cannot be read, parsing fails, or
// template rendering encounters an error.
func (g *Generator) Generate(ctx context.Context) (*GeneratedBlog, error) {
	g.logger.DebugContext(ctx, "Creating parser for generate call")
	p := parser.NewWithConfig(&g.config.ParserConfig)

	posts, err := p.ParseDirectory(ctx, g.config.PostsDir)
	if err != nil {
		return nil, err
	}

	// Step 2: If RawOutput mode, return immediately with raw HTML
	if g.config.RawOutput {
		return g.assembleRawBlog(posts), nil
	}

	// Step 3: Apply templates (TODO: future work)
	return nil, fmt.Errorf("Only raw output is enabled at this point")
}

// DebugConfig logs the current generator configuration at the debug level.
//
// This method is useful for troubleshooting and verifying configuration
// settings during development or when diagnosing issues. The output includes
// all generator configuration details and respects the provided context for
// structured logging.
//
// The log output will only appear if the logger is configured to show debug
// level messages.
func (g *Generator) DebugConfig(ctx context.Context) {
	g.logger.DebugContext(ctx, g.config.String())
}

func (g *Generator) assembleRawBlog(posts models.PostList) *GeneratedBlog {
	blog := NewEmptyGeneratedBlog()

	for _, post := range posts {
		blog.Posts[post.Slug] = post.Content
	}

	return blog
}

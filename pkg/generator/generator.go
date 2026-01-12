package generator

import (
	"context"
	"io/fs"
	"log/slog"
)

// Generator produces HTML output based on its input configuration.
// It reads markdown files from a configured filesystem and renders them
// as HTML using templates.
//
// A Generator is safe for concurrent use after creation, but Generate
// operations should not be called concurrently on the same instance.
type Generator struct {
	config *GeneratorConfig
	logger *slog.Logger
}

// New creates a new Generator with the specified options.
// It returns an error if the configuration is invalid or if required
// resources cannot be initialized.
//
// Options can be provided to customize behavior such as template directories,
// posts per page, and other generation parameters.
func New(posts fs.FS, opts ...Option) (*Generator, error) {
	logger := slog.Default()

	// TODO: Validate posts input somehow?
	config := GeneratorConfig{
		PostsDir: posts,
	}

	// Run options on config
	for _, opt := range opts {
		opt(&config)
	}

	gen := Generator{
		config: &config,
		logger: logger,
	}

	return &gen, nil
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
	return nil, nil
}

// ValidateConfig logs the current config at the debug level
func (g *Generator) DebugConfig(ctx context.Context) {
	g.logger.DebugContext(ctx, g.config.String())
}

package generator

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"time"

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
	PostsDir fs.FS // The filesystem containing the input posts in markdown

	config.RawOutput
	config.SiteTitle
	ParserConfig parser.Config // The config to use when parsing

	logger   *slog.Logger
	renderer *TemplateRenderer
}

func (c Generator) String() string {
	return fmt.Sprintf(`Generator Config
- RawOutput           %t`,
		c.RawOutput,
	)
}

// New creates a new Generator with the specified options.
// It returns an error if the configuration is invalid or if required
// resources cannot be initialized.
//
// Options can be provided to customize behavior such as template directories,
// posts per page, and other generation parameters.
func New(posts fs.FS, renderer *TemplateRenderer, opts ...config.Option) *Generator {
	logger := slog.Default()

	gen := Generator{
		PostsDir: posts,
		logger:   logger,
		renderer: renderer,
	}

	// Run options on config
	for _, opt := range opts {
		if opt.WithRawOutputFunc != nil {
			opt.WithRawOutputFunc(&gen.RawOutput)
		} else if opt.WithSiteTitleFunc != nil {
			opt.WithSiteTitleFunc(&gen.SiteTitle)
		}
	}

	if gen.SiteTitle.SiteTitle == "" {
		gen.SiteTitle = config.SiteTitle{
			SiteTitle: "GoBlog",
		}
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
// # Output Mode Behavior
//
// When RawOutput is disabled (default):
//   - Posts contain fully templated HTML pages
//   - Tags map contains rendered tag pages
//   - Index contains the complete index page with template
//
// When RawOutput is enabled (via config.WithRawOutput()):
//   - Posts contain only the Markdown-to-HTML conversion without templates
//   - Tags map will be empty (tag generation is skipped)
//   - Index will be empty or minimal
//   - Useful for custom integration scenarios
//
// Generate respects the provided context and will return early with
// context.Canceled or context.DeadlineExceeded if the context is canceled
// or times out.
//
// It returns an error if markdown files cannot be read, parsing fails, or
// template rendering encounters an error.
func (g *Generator) Generate(ctx context.Context) (*GeneratedBlog, error) {
	g.logger.DebugContext(ctx, "Creating parser for generate call")
	p := parser.NewWithConfig(&g.ParserConfig)

	posts, err := p.ParseDirectory(ctx, g.PostsDir)
	if err != nil {
		return nil, err
	}

	// Step 2: If RawOutput mode, return immediately with raw HTML
	if g.RawOutput.RawOutput {
		g.logger.InfoContext(ctx, "Raw output enabled, ignoring templates")
		return g.assembleRawBlog(posts), nil
	}

	// Step 3: Apply templates
	return g.assembleBlogWithTemplates(ctx, posts)
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
	g.logger.DebugContext(ctx, g.String())
}

func (g *Generator) assembleRawBlog(posts models.PostList) *GeneratedBlog {
	blog := NewEmptyGeneratedBlog()

	for _, post := range posts {
		blog.Posts[post.Slug] = post.Content
	}

	return blog
}

func (g *Generator) assembleBlogWithTemplates(ctx context.Context, posts models.PostList) (*GeneratedBlog, error) {
	g.logger.DebugContext(ctx, "Rendering posts with templates")
	blog := NewEmptyGeneratedBlog()

	// Sort posts by date descending
	posts.SortByDate()

	// Render individual post pages
	for _, post := range posts {
		data := models.PostPageData{
			BaseData: models.BaseData{
				SiteTitle:   g.SiteTitle.SiteTitle,
				PageTitle:   post.Title,
				Description: post.Description,
				Year:        time.Now().Year(),
			},
			Post: post,
		}

		rendered, err := g.renderer.RenderPost(data)
		if err != nil {
			return nil, fmt.Errorf("failed to render post %s: %w", post.Slug, err)
		}

		blog.Posts[post.Slug] = rendered
	}

	// Render index page
	indexData := models.IndexPageData{
		BaseData: models.BaseData{
			SiteTitle:   g.SiteTitle.SiteTitle,
			PageTitle:   "Home",
			Description: "Recent blog posts",
			Year:        time.Now().Year(),
		},
		Posts:      posts,
		TotalPosts: len(posts),
	}

	index, err := g.renderer.RenderIndex(indexData)
	if err != nil {
		return nil, fmt.Errorf("failed to render index: %w", err)
	}
	blog.Index = index

	// Render tag pages
	allTags := posts.GetAllTags()
	for _, tag := range allTags {
		tagPosts := posts.FilterByTag(tag)

		tagData := models.TagPageData{
			BaseData: models.BaseData{
				SiteTitle:   g.SiteTitle.SiteTitle,
				PageTitle:   "Tag: " + tag,
				Description: fmt.Sprintf("Posts tagged with %s", tag),
				Year:        time.Now().Year(),
			},
			Tag:       tag,
			Posts:     tagPosts,
			PostCount: len(tagPosts),
		}

		rendered, err := g.renderer.RenderTag(tagData)
		if err != nil {
			return nil, fmt.Errorf("failed to render tag page %s: %w", tag, err)
		}

		blog.Tags[tag] = rendered
	}

	return blog, nil
}

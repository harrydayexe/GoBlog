package generator

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/harrydayexe/GoBlog/internal/gen/config"
	"github.com/harrydayexe/GoBlog/internal/gen/feeds"
	"github.com/harrydayexe/GoBlog/internal/gen/log"
	"github.com/harrydayexe/GoBlog/internal/gen/parser"
	"github.com/harrydayexe/GoBlog/internal/gen/sitemap"
	"github.com/harrydayexe/GoBlog/internal/gen/template"
	"github.com/harrydayexe/GoBlog/pkg/models"
)

// Generator orchestrates the static site generation
type Generator struct {
	cfg      config.Config
	parser   *parser.Parser
	template *template.Engine
	logger   log.Logger
}

// New creates a new generator
func New(cfg config.Config, logger log.Logger) (*Generator, error) {
	if logger == nil {
		logger = log.NewCLILogger("GENERATOR", cfg.Verbose)
	}

	// Create parser with logger
	p := parser.New(logger)

	// Create template engine with logger
	tmpl, err := template.New(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize template engine: %w", err)
	}

	return &Generator{
		cfg:      cfg,
		parser:   p,
		template: tmpl,
		logger:   logger,
	}, nil
}

// Generate runs the full site generation process
func (g *Generator) Generate() error {
	g.logger.Info("Starting site generation...")

	// Parse all posts
	g.logger.Info("Parsing markdown files...")
	posts, err := g.parser.ParseAllPosts(g.cfg.InputFolder)
	if err != nil {
		return fmt.Errorf("failed to parse posts: %w", err)
	}

	// Filter published posts
	publishedPosts := posts.FilterPublished()
	g.logger.Info("Found %d published post(s) (excluding %d draft(s))", len(publishedPosts), len(posts)-len(publishedPosts))

	if len(publishedPosts) == 0 {
		return fmt.Errorf("no published posts found (all posts are drafts)")
	}

	// Create output directory
	if err := os.MkdirAll(g.cfg.OutputFolder, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate individual post pages
	if err := g.generatePosts(publishedPosts); err != nil {
		return fmt.Errorf("failed to generate posts: %w", err)
	}

	// Generate index pages with pagination
	if err := g.generateIndex(publishedPosts); err != nil {
		return fmt.Errorf("failed to generate index: %w", err)
	}

	// Generate tag pages
	if err := g.generateTagPages(publishedPosts); err != nil {
		return fmt.Errorf("failed to generate tag pages: %w", err)
	}

	// Generate RSS and Atom feeds
	if err := g.generateFeeds(publishedPosts); err != nil {
		return fmt.Errorf("failed to generate feeds: %w", err)
	}

	// Generate sitemap
	if err := g.generateSitemap(publishedPosts); err != nil {
		return fmt.Errorf("failed to generate sitemap: %w", err)
	}

	// Copy static files
	if err := g.copyStaticFiles(); err != nil {
		g.logger.Warn("Failed to copy static files: %v", err)
		g.logger.Info("Continuing without static files...")
	}

	g.logger.Info("Site generation complete! Output: %s", g.cfg.OutputFolder)
	return nil
}

// generatePosts generates individual HTML pages for each post
func (g *Generator) generatePosts(posts models.PostList) error {
	g.logger.Info("Generating %d post page(s)...", len(posts))

	// Create blog directory
	blogDir := filepath.Join(g.cfg.OutputFolder, g.cfg.BlogPath)
	if err := os.MkdirAll(blogDir, 0755); err != nil {
		return fmt.Errorf("failed to create blog directory: %w", err)
	}

	for _, post := range posts {
		// Render the post
		html, err := g.template.RenderPost(post)
		if err != nil {
			return fmt.Errorf("failed to render post %s: %w", post.Title, err)
		}

		// Write to file
		outputPath := filepath.Join(blogDir, post.Slug+".html")
		if err := template.WriteHTML(outputPath, html); err != nil {
			return fmt.Errorf("failed to write post %s: %w", post.Title, err)
		}

		g.logger.Debug("Generated: %s", outputPath)
	}

	g.logger.Info("Post pages generated successfully")
	return nil
}

// generateIndex generates index pages with pagination
func (g *Generator) generateIndex(posts models.PostList) error {
	g.logger.Info("Generating index pages...")

	// Calculate pagination
	totalPosts := len(posts)
	postsPerPage := g.cfg.PostsPerPage
	totalPages := (totalPosts + postsPerPage - 1) / postsPerPage

	g.logger.Debug("Total posts: %d, Posts per page: %d, Total pages: %d", totalPosts, postsPerPage, totalPages)

	blogDir := filepath.Join(g.cfg.OutputFolder, g.cfg.BlogPath)

	for page := 1; page <= totalPages; page++ {
		// Get posts for this page
		start := (page - 1) * postsPerPage
		end := start + postsPerPage
		if end > totalPosts {
			end = totalPosts
		}
		pagePosts := posts[start:end]

		// Render the index page
		html, err := g.template.RenderIndex(pagePosts, page, totalPages)
		if err != nil {
			return fmt.Errorf("failed to render index page %d: %w", page, err)
		}

		// Determine output path
		var outputPath string
		if page == 1 {
			// First page is the main index
			outputPath = filepath.Join(blogDir, "index.html")
		} else {
			outputPath = filepath.Join(blogDir, fmt.Sprintf("page%d.html", page))
		}

		// Write to file
		if err := template.WriteHTML(outputPath, html); err != nil {
			return fmt.Errorf("failed to write index page %d: %w", page, err)
		}

		g.logger.Debug("Generated: %s", outputPath)
	}

	// Also create index at root if BlogPath is not "/"
	if g.cfg.BlogPath != "/" {
		firstPagePosts := posts
		if len(posts) > postsPerPage {
			firstPagePosts = posts[:postsPerPage]
		}
		html, err := g.template.RenderIndex(firstPagePosts, 1, totalPages)
		if err != nil {
			return fmt.Errorf("failed to render root index: %w", err)
		}

		rootIndex := filepath.Join(g.cfg.OutputFolder, "index.html")
		if err := template.WriteHTML(rootIndex, html); err != nil {
			return fmt.Errorf("failed to write root index: %w", err)
		}
		g.logger.Debug("Generated: %s", rootIndex)
	}

	g.logger.Info("Index pages generated successfully (%d page(s))", totalPages)
	return nil
}

// generateTagPages generates pages for each tag
func (g *Generator) generateTagPages(posts models.PostList) error {
	tags := posts.GetAllTags()
	if len(tags) == 0 {
		g.logger.Info("No tags found, skipping tag pages")
		return nil
	}

	g.logger.Info("Generating %d tag page(s)...", len(tags))

	// Create tags directory
	tagsDir := filepath.Join(g.cfg.OutputFolder, g.cfg.BlogPath, "tags")
	if err := os.MkdirAll(tagsDir, 0755); err != nil {
		return fmt.Errorf("failed to create tags directory: %w", err)
	}

	for _, tag := range tags {
		// Filter posts by tag
		tagPosts := posts.FilterByTag(tag)

		// Render the tag page
		html, err := g.template.RenderTag(tag, tagPosts, posts)
		if err != nil {
			return fmt.Errorf("failed to render tag page %s: %w", tag, err)
		}

		// Write to file
		outputPath := filepath.Join(tagsDir, tag+".html")
		if err := template.WriteHTML(outputPath, html); err != nil {
			return fmt.Errorf("failed to write tag page %s: %w", tag, err)
		}

		g.logger.Debug("Generated: %s (%d post(s))", outputPath, len(tagPosts))
	}

	g.logger.Info("Tag pages generated successfully")
	return nil
}

// copyStaticFiles copies static assets to the output directory
func (g *Generator) copyStaticFiles() error {
	// Check if static folder exists
	if _, err := os.Stat(g.cfg.StaticFolder); os.IsNotExist(err) {
		g.logger.Debug("Static folder %s does not exist, skipping", g.cfg.StaticFolder)
		return nil
	}

	g.logger.Info("Copying static files from %s...", g.cfg.StaticFolder)

	err := filepath.Walk(g.cfg.StaticFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Calculate relative path
		relPath, err := filepath.Rel(g.cfg.StaticFolder, path)
		if err != nil {
			return err
		}

		// Destination path
		destPath := filepath.Join(g.cfg.OutputFolder, relPath)

		// Create destination directory
		destDir := filepath.Dir(destPath)
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", destDir, err)
		}

		// Copy file
		if err := copyFile(path, destPath); err != nil {
			return fmt.Errorf("failed to copy %s: %w", path, err)
		}

		g.logger.Debug("Copied: %s -> %s", path, destPath)
		return nil
	})

	if err != nil {
		return err
	}

	g.logger.Info("Static files copied successfully")
	return nil
}

// generateFeeds generates RSS and Atom feeds
func (g *Generator) generateFeeds(posts models.PostList) error {
	g.logger.Info("Generating RSS and Atom feeds...")

	blogDir := filepath.Join(g.cfg.OutputFolder, g.cfg.BlogPath)

	// Prepare feed configuration
	rssConfig := feeds.RSSConfig{
		Title:       g.cfg.Site.Title,
		Link:        g.cfg.Site.URL,
		Description: g.cfg.Site.Description,
		Language:    "en-us",
		Copyright:   fmt.Sprintf("Copyright © %s", g.cfg.Site.Author),
		Author:      g.cfg.Site.Author,
		BlogPath:    g.cfg.BlogPath,
	}

	atomConfig := feeds.AtomConfig{
		Title:       g.cfg.Site.Title,
		Link:        g.cfg.Site.URL,
		Description: g.cfg.Site.Description,
		Author:      g.cfg.Site.Author,
		BlogPath:    g.cfg.BlogPath,
	}

	// Generate RSS feed
	rssContent, err := feeds.GenerateRSS(posts, rssConfig)
	if err != nil {
		return fmt.Errorf("failed to generate RSS feed: %w", err)
	}

	rssPath := filepath.Join(blogDir, "feed.xml")
	if err := os.WriteFile(rssPath, []byte(rssContent), 0644); err != nil {
		return fmt.Errorf("failed to write RSS feed: %w", err)
	}
	g.logger.Debug("Generated: %s", rssPath)

	// Generate Atom feed
	atomContent, err := feeds.GenerateAtom(posts, atomConfig)
	if err != nil {
		return fmt.Errorf("failed to generate Atom feed: %w", err)
	}

	atomPath := filepath.Join(blogDir, "atom.xml")
	if err := os.WriteFile(atomPath, []byte(atomContent), 0644); err != nil {
		return fmt.Errorf("failed to write Atom feed: %w", err)
	}
	g.logger.Debug("Generated: %s", atomPath)

	g.logger.Info("Feed files generated successfully")
	return nil
}

// generateSitemap generates sitemap.xml
func (g *Generator) generateSitemap(posts models.PostList) error {
	g.logger.Info("Generating sitemap...")

	sitemapConfig := sitemap.SitemapConfig{
		SiteURL:  g.cfg.Site.URL,
		BlogPath: g.cfg.BlogPath,
	}

	sitemapContent, err := sitemap.GenerateSitemap(posts, sitemapConfig)
	if err != nil {
		return fmt.Errorf("failed to generate sitemap: %w", err)
	}

	// Write to root of output folder (sitemap should be at site root)
	sitemapPath := filepath.Join(g.cfg.OutputFolder, "sitemap.xml")
	if err := os.WriteFile(sitemapPath, []byte(sitemapContent), 0644); err != nil {
		return fmt.Errorf("failed to write sitemap: %w", err)
	}
	g.logger.Debug("Generated: %s", sitemapPath)

	g.logger.Info("Sitemap generated successfully")
	return nil
}

// copyFile copies a single file from src to dst
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	return dstFile.Sync()
}

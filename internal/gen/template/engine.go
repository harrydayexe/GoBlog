package template

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"github.com/harrydayexe/GoBlog/internal/gen/config"
	"github.com/harrydayexe/GoBlog/internal/gen/log"
	"github.com/harrydayexe/GoBlog/pkg/models"
)

// Engine handles template loading and rendering
type Engine struct {
	postTemplate  *template.Template
	indexTemplate *template.Template
	tagTemplate   *template.Template
	cfg           config.Config
	logger        log.CLILogger
}

// New creates a new template engine
func New(cfg config.Config) (*Engine, error) {
	engine := &Engine{
		cfg:    cfg,
		logger: *log.NewCLILogger("TEMPLATE", cfg.Verbose),
	}

	// Load templates
	if err := engine.loadTemplates(); err != nil {
		return nil, err
	}

	return engine, nil
}

// loadTemplates loads templates from custom paths or uses defaults
func (e *Engine) loadTemplates() error {
	var err error

	// Load post template
	if e.cfg.Templates.PostTemplate != "" {
		e.logger.Info("Loading custom post template: %s", e.cfg.Templates.PostTemplate)
		e.postTemplate, err = template.ParseFiles(e.cfg.Templates.PostTemplate)
		if err != nil {
			return fmt.Errorf("failed to parse post template: %w", err)
		}
	} else {
		e.logger.Debug("Using default post template")
		e.postTemplate, err = template.New("post").Parse(defaultPostTemplate)
		if err != nil {
			return fmt.Errorf("failed to parse default post template: %w", err)
		}
	}

	// Load index template
	if e.cfg.Templates.IndexTemplate != "" {
		e.logger.Info("Loading custom index template: %s", e.cfg.Templates.IndexTemplate)
		e.indexTemplate, err = template.ParseFiles(e.cfg.Templates.IndexTemplate)
		if err != nil {
			return fmt.Errorf("failed to parse index template: %w", err)
		}
	} else {
		e.logger.Debug("Using default index template")
		e.indexTemplate, err = template.New("index").Funcs(templateFuncs).Parse(defaultIndexTemplate)
		if err != nil {
			return fmt.Errorf("failed to parse default index template: %w", err)
		}
	}

	// Load tag template
	if e.cfg.Templates.TagTemplate != "" {
		e.logger.Info("Loading custom tag template: %s", e.cfg.Templates.TagTemplate)
		e.tagTemplate, err = template.ParseFiles(e.cfg.Templates.TagTemplate)
		if err != nil {
			return fmt.Errorf("failed to parse tag template: %w", err)
		}
	} else {
		e.logger.Debug("Using default tag template")
		e.tagTemplate, err = template.New("tag").Parse(defaultTagTemplate)
		if err != nil {
			return fmt.Errorf("failed to parse default tag template: %w", err)
		}
	}

	e.logger.Info("Templates loaded successfully")
	return nil
}

// PostData holds data for rendering a post page
type PostData struct {
	Post     *models.Post
	Site     config.SiteMetadata
	BlogPath string
}

// IndexData holds data for rendering an index page
type IndexData struct {
	Posts      models.PostList
	Site       config.SiteMetadata
	BlogPath   string
	Page       int
	TotalPages int
	HasNext    bool
	HasPrev    bool
	AllTags    []string
}

// TagData holds data for rendering a tag page
type TagData struct {
	Tag      string
	Posts    models.PostList
	Site     config.SiteMetadata
	BlogPath string
	AllTags  []string
}

// RenderPost renders a single post page
func (e *Engine) RenderPost(post *models.Post) (string, error) {
	data := PostData{
		Post:     post,
		Site:     e.cfg.Site,
		BlogPath: e.cfg.BlogPath,
	}

	var buf bytes.Buffer
	if err := e.postTemplate.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to render post %s: %w", post.Title, err)
	}

	return buf.String(), nil
}

// RenderIndex renders an index/list page with pagination
func (e *Engine) RenderIndex(posts models.PostList, page int, totalPages int) (string, error) {
	data := IndexData{
		Posts:      posts,
		Site:       e.cfg.Site,
		BlogPath:   e.cfg.BlogPath,
		Page:       page,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
		AllTags:    posts.GetAllTags(),
	}

	var buf bytes.Buffer
	if err := e.indexTemplate.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to render index page %d: %w", page, err)
	}

	return buf.String(), nil
}

// RenderTag renders a tag page
func (e *Engine) RenderTag(tag string, posts models.PostList, allPosts models.PostList) (string, error) {
	data := TagData{
		Tag:      tag,
		Posts:    posts,
		Site:     e.cfg.Site,
		BlogPath: e.cfg.BlogPath,
		AllTags:  allPosts.GetAllTags(),
	}

	var buf bytes.Buffer
	if err := e.tagTemplate.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to render tag page %s: %w", tag, err)
	}

	return buf.String(), nil
}

// WriteHTML writes HTML content to a file
func WriteHTML(outputPath string, content string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write the file
	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", outputPath, err)
	}

	return nil
}

// templateFuncs contains custom template functions
var templateFuncs = template.FuncMap{
	"add": func(a, b int) int {
		return a + b
	},
	"sub": func(a, b int) int {
		return a - b
	},
}

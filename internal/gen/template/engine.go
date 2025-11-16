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
	logger        log.Logger
}

// New creates a new template engine
func New(cfg config.Config, logger log.Logger) (*Engine, error) {
	if logger == nil {
		logger = log.NewCLILogger("TEMPLATE", cfg.Verbose)
	}

	engine := &Engine{
		cfg:    cfg,
		logger: logger,
	}

	// Load templates
	if err := engine.loadTemplates(); err != nil {
		return nil, err
	}

	return engine, nil
}

// loadTemplates loads templates from the template directory
func (e *Engine) loadTemplates() error {
	templateDir := e.cfg.TemplateDir

	// Check if template directory exists
	if _, err := os.Stat(templateDir); os.IsNotExist(err) {
		return fmt.Errorf("template directory does not exist: %s", templateDir)
	}

	e.logger.Info("Loading templates from: %s", templateDir)

	// Load post template
	postPath := filepath.Join(templateDir, "post.html")
	if _, err := os.Stat(postPath); err != nil {
		return fmt.Errorf("post template not found: %s", postPath)
	}
	e.logger.Debug("Loading post template: %s", postPath)
	postTmpl, err := template.ParseFiles(postPath)
	if err != nil {
		return fmt.Errorf("failed to parse post template: %w", err)
	}
	e.postTemplate = postTmpl

	// Load index template (with custom functions)
	indexPath := filepath.Join(templateDir, "index.html")
	if _, err := os.Stat(indexPath); err != nil {
		return fmt.Errorf("index template not found: %s", indexPath)
	}
	e.logger.Debug("Loading index template: %s", indexPath)
	indexTmpl, err := template.New("index.html").Funcs(templateFuncs).ParseFiles(indexPath)
	if err != nil {
		return fmt.Errorf("failed to parse index template: %w", err)
	}
	e.indexTemplate = indexTmpl

	// Load tag template
	tagPath := filepath.Join(templateDir, "tag.html")
	if _, err := os.Stat(tagPath); err != nil {
		return fmt.Errorf("tag template not found: %s", tagPath)
	}
	e.logger.Debug("Loading tag template: %s", tagPath)
	tagTmpl, err := template.ParseFiles(tagPath)
	if err != nil {
		return fmt.Errorf("failed to parse tag template: %w", err)
	}
	e.tagTemplate = tagTmpl

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

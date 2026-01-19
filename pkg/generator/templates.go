package generator

import (
	"bytes"
	"html/template"
	"io/fs"
	"log/slog"
	"time"

	"github.com/harrydayexe/GoBlog/v2/pkg/models"
)

// TemplateRenderer handles template loading and rendering.
type TemplateRenderer struct {
	templates *template.Template
}

// NewTemplateRenderer creates a new template renderer from a filesystem.
// templatesFS should point to a directory containing layouts/, pages/, and partials/.
func NewTemplateRenderer(templatesFS fs.FS) (*TemplateRenderer, error) {
	// Define custom template functions
	funcMap := template.FuncMap{
		"formatDate": func(t time.Time) string {
			return t.Format("January 2, 2006")
		},
		"shortDate": func(t time.Time) string {
			return t.Format("Jan 2, 2006")
		},
		"year": func() int {
			return time.Now().Year()
		},
	}

	// Parse all template files
	tmpl := template.New("").Funcs(funcMap)

	// Parse in order: layouts, partials, pages
	patterns := []string{
		"layouts/*.tmpl",
		"partials/*.tmpl",
		"pages/*.tmpl",
	}

	for _, pattern := range patterns {
		matches, err := fs.Glob(templatesFS, pattern)
		if err != nil {
			return nil, err
		}

		for _, match := range matches {
			content, err := fs.ReadFile(templatesFS, match)
			if err != nil {
				return nil, err
			}

			_, err = tmpl.New(match).Parse(string(content))
			if err != nil {
				return nil, err
			}
		}
	}

	return &TemplateRenderer{templates: tmpl}, nil
}

// renderPost renders a single post page.
func (tr *TemplateRenderer) RenderPost(data models.PostPageData) ([]byte, error) {
	var buf bytes.Buffer
	err := tr.templates.ExecuteTemplate(&buf, "pages/post.tmpl", data)
	slog.Debug("Rendered post " + data.Post.Slug)
	return buf.Bytes(), err
}

// renderIndex renders the index/homepage.
func (tr *TemplateRenderer) RenderIndex(data models.IndexPageData) ([]byte, error) {
	var buf bytes.Buffer
	err := tr.templates.ExecuteTemplate(&buf, "pages/index.tmpl", data)
	slog.Debug("Rendered index page", slog.Int("number of posts", len(data.Posts)))
	return buf.Bytes(), err
}

// renderTag renders a tag page.
func (tr *TemplateRenderer) RenderTag(data models.TagPageData) ([]byte, error) {
	var buf bytes.Buffer
	err := tr.templates.ExecuteTemplate(&buf, "pages/tag.tmpl", data)
	slog.Debug("Rendered tag page " + data.Tag)
	return buf.Bytes(), err
}

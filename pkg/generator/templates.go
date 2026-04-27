package generator

import (
	"bytes"
	"html/template"
	"io/fs"
	"log/slog"
	"time"

	"github.com/harrydayexe/GoBlog/v2/pkg/models"
)

// TemplateRenderer parses a tree of HTML templates from an fs.FS and renders
// each page type (post, index, tag, tags-index) into HTML.
//
// Create a renderer once via [NewTemplateRenderer]; the resulting value is
// safe for concurrent use by multiple goroutines. The internal
// *template.Template is not exposed — callers customise rendering by
// supplying a different fs.FS rather than mutating the renderer.
type TemplateRenderer struct {
	templates *template.Template
}

// NewTemplateRenderer parses every *.tmpl file under templatesFS and returns
// a renderer ready to produce HTML for each page type.
//
// templatesFS must contain the following top-level directories:
//
//	pages/    required — must contain post.tmpl, index.tmpl, tag.tmpl,
//	          and tags-index.tmpl
//	partials/ required — each file must {{define}} one named block;
//	          the default templates expect "head", "header", "footer",
//	          and "post-card"
//	layouts/  optional — loaded but not executed by any Render* method;
//	          pages are self-contained documents that inline partials directly
//
// Each *.tmpl file is registered under its full glob path as the template
// name (e.g. "pages/post.tmpl", "partials/head.tmpl"). Pages reference
// partials by their defined name ({{template "head" .}}), and Render*
// methods invoke pages by their path.
//
// The following helpers are available in every template's FuncMap:
//
//	formatDate(t time.Time) string   formats t as "January 2, 2006"
//	shortDate(t time.Time) string    formats t as "Jan 2, 2006"
//	year() int                       returns the current calendar year
//
// Pass [github.com/harrydayexe/GoBlog/v2/pkg/templates.Default] to use the
// built-in templates, or any fs.FS (e.g. os.DirFS("./mytemplates")) for a
// custom theme. Returns an error if any template fails to parse.
func NewTemplateRenderer(templatesFS fs.FS) (*TemplateRenderer, error) {
	// helpers available to all templates
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

// RenderPost renders a single post page by executing pages/post.tmpl with
// the supplied [models.PostPageData]. Returns the rendered HTML or any error
// from template execution.
func (tr *TemplateRenderer) RenderPost(data models.PostPageData) ([]byte, error) {
	var buf bytes.Buffer
	err := tr.templates.ExecuteTemplate(&buf, "pages/post.tmpl", data)
	slog.Debug("Rendered post " + data.Post.Slug)
	return buf.Bytes(), err
}

// RenderIndex renders the index/homepage by executing pages/index.tmpl with
// the supplied [models.IndexPageData]. Returns the rendered HTML or any error
// from template execution.
func (tr *TemplateRenderer) RenderIndex(data models.IndexPageData) ([]byte, error) {
	var buf bytes.Buffer
	err := tr.templates.ExecuteTemplate(&buf, "pages/index.tmpl", data)
	slog.Debug("Rendered index page", slog.Int("number of posts", len(data.Posts)))
	return buf.Bytes(), err
}

// RenderTag renders a tag page by executing pages/tag.tmpl with the supplied
// [models.TagPageData]. Returns the rendered HTML or any error from template
// execution.
func (tr *TemplateRenderer) RenderTag(data models.TagPageData) ([]byte, error) {
	var buf bytes.Buffer
	err := tr.templates.ExecuteTemplate(&buf, "pages/tag.tmpl", data)
	slog.Debug("Rendered tag page " + data.Tag)
	return buf.Bytes(), err
}

// RenderTagsIndex renders the tags index page by executing
// pages/tags-index.tmpl with the supplied [models.TagsIndexPageData]. Returns
// the rendered HTML or any error from template execution.
func (tr *TemplateRenderer) RenderTagsIndex(data models.TagsIndexPageData) ([]byte, error) {
	var buf bytes.Buffer
	err := tr.templates.ExecuteTemplate(&buf, "pages/tags-index.tmpl", data)
	slog.Debug("Rendered tags index page", slog.Int("total tags", data.TotalTags))
	return buf.Bytes(), err
}

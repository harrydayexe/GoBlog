// Package templates provides embedded default templates for the GoBlog system.
//
// The Default variable contains the embedded filesystem with default HTML templates
// used by the generator when no custom templates are provided.
package templates

import (
	"embed"
	"io/fs"
)

//go:embed default/**/*.tmpl
var rawDefault embed.FS

// Default is the embedded default template tree used by the generator when no
// custom templates are supplied. It is rooted at the templates directory so that
// paths such as "pages/post.tmpl" are directly accessible without any "default/"
// prefix, matching the layout expected by
// [github.com/harrydayexe/GoBlog/v2/pkg/generator.NewTemplateRenderer].
//
// Its directory layout — which any custom fs.FS passed to NewTemplateRenderer
// must mirror — is:
//
//	pages/
//	  post.tmpl          executed by TemplateRenderer.RenderPost
//	  index.tmpl         executed by TemplateRenderer.RenderIndex
//	  tag.tmpl           executed by TemplateRenderer.RenderTag
//	  tags-index.tmpl    executed by TemplateRenderer.RenderTagsIndex
//	partials/
//	  head.tmpl          {{define "head"}}
//	  header.tmpl        {{define "header"}}
//	  footer.tmpl        {{define "footer"}}
//	  post-card.tmpl     {{define "post-card"}}
//	layouts/
//	  base.tmpl          loaded but not executed; pages are self-contained
//
// Each page template is a complete HTML document that references partials via
// {{template "head" .}}, {{template "header" .}}, etc. Custom templates must
// follow the same convention: define the four named blocks in their partials/
// directory and reference them from each page template.
var Default fs.FS = func() fs.FS {
	sub, err := fs.Sub(rawDefault, "default")
	if err != nil {
		panic("embedded default templates are missing the 'default/' directory: " + err.Error())
	}
	return sub
}()

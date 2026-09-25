// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package generator provides functionality for generating static HTML blog sites
// from markdown files.
//
// The generator uses the functional options pattern for configuration. Required
// parameters are passed as positional arguments, while optional parameters are
// configured via option functions.
//
// # Basic Usage
//
// Create a generator and generate a blog:
//
//	import (
//		"context"
//		"os"
//		"github.com/harrydayexe/GoBlog/v2/pkg/generator"
//		"github.com/harrydayexe/GoBlog/v2/pkg/config"
//		"github.com/harrydayexe/GoBlog/v2/pkg/templates"
//	)
//
//	fsys := os.DirFS("posts/")
//	renderer, err := generator.NewTemplateRenderer(templates.Default)
//	if err != nil {
//		log.Fatal(err)
//	}
//	gen := generator.New(fsys, renderer)
//
//	// Generate the blog
//	blog, err := gen.Generate(context.Background())
//	if err != nil {
//		log.Fatal(err)
//	}
//
// In raw output mode no templates are applied, so the renderer may be nil:
//
//	gen := generator.New(fsys, nil, config.WithRawOutput())
//
// # Logger injection
//
// Supply a structured logger via config.WithLogger (a BaseOption, so call
// AsGeneratorOption to pass it to generator.New):
//
//	gen := generator.New(fsys, renderer,
//	    config.WithLogger(myLogger).AsGeneratorOption(),
//	)
//
// # Custom Template Functions
//
// Additional template functions can be registered via [config.WithFuncs] and
// passed to [NewTemplateRenderer]. Once registered, the function is available
// in every template as a regular call:
//
//	import (
//		"html/template"
//		"strings"
//	)
//
//	renderer, err := generator.NewTemplateRenderer(
//		templates.Default,
//		config.WithFuncs(template.FuncMap{
//			"upper": strings.ToUpper,
//		}),
//	)
//
// In a template:
//
//	<h1>{{upper .Post.Title}}</h1>
//
// The built-in helpers (formatDate, shortDate, year) remain available unless
// intentionally replaced. Registering a function whose name matches a built-in
// silently replaces that built-in — useful for custom date formats but a
// potential footgun if done accidentally. See [config.WithFuncs] for the full list
// of reserved names.
//
// # Custom Template Data
//
// Arbitrary key-value data from the calling application can be injected into
// every rendered page via [config.WithCustomData]. The data is accessible in
// templates as {{.Custom.key}}:
//
//	gen := generator.New(fsys, renderer,
//		config.WithCustomData(map[string]any{
//			"author":      "Jane Smith",
//			"analyticsID": "UA-12345",
//		}),
//	)
//
// In a template:
//
//	{{with .Custom}}
//	    <meta name="author" content="{{.author}}">
//	{{end}}
//
// .Custom is nil when no WithCustomData option is supplied; templates should
// guard access with {{with .Custom}} or {{if .Custom}} to avoid nil-map errors.
// Multiple WithCustomData calls merge their maps; later values overwrite earlier
// ones for duplicate keys.
//
// # Page Path
//
// Every rendered page receives a Path field on its template data containing the
// URL path for that page. It is accessible in templates as {{.Path}}:
//
//	<a href="{{.Path}}">this page</a>
//
// By default (clean-URL mode) the value includes BlogRoot and no .html suffix:
//   - Index page:        /  (or /blog/ when BlogRoot = "/blog/")
//   - Post page:         /posts/<slug>
//   - Tag page:          /tags/<tag>
//   - Tags index page:   /tags
//   - Series page:       /series/<slug>
//   - Series index page: /series
//
// When [config.WithHTMLPaths] is applied (automatically set by the goblog
// generate CLI), paths include the .html extension to match the files written
// to disk:
//   - Index page:        /index.html  (or /blog.html when BlogRoot = "/blog/")
//   - Post page:         /posts/<slug>.html
//   - Tag page:          /tags/<tag>.html
//   - Tags index page:   /tags.html
//   - Series page:       /series/<slug>.html
//   - Series index page: /series.html
//
// The pkg/server package accepts both clean URLs and .html URLs via its
// built-in StripHTMLExtension middleware, so the server always uses clean-URL
// paths regardless of the generating option.
//
// # Canonical URL
//
// When a base URL is configured via [config.WithBaseURL], each page also
// receives a CanonicalURL field holding the fully-qualified URL for that page
// (the base URL joined with Path). It is empty when no base URL is set, so
// templates must guard on it rather than emitting an empty tag:
//
//	{{if .CanonicalURL}}
//	<link rel="canonical" href="{{.CanonicalURL}}">
//	<meta property="og:url" content="{{.CanonicalURL}}">
//	{{end}}
//
// # Open Graph Type
//
// Each page also receives an OGType field naming its Open Graph object type:
// "article" for post pages and "website" for the index, tag, and tags-index
// pages. Templates emit it as:
//
//	<meta property="og:type" content="{{.OGType}}">
//
// # Article Metadata
//
// Post pages additionally receive an Article field ([models.ArticleMeta])
// holding the post's publication date, last-edited date, author, tags, and
// reading time. It is nil on every other page type, which lets a <head>
// partial shared by all pages emit article-specific markup for posts only:
//
//	{{with .Article}}
//	<meta property="article:published_time" content="{{.PublishedISO}}">
//	{{with .ModifiedISO}}<meta property="article:modified_time" content="{{.}}">{{end}}
//	{{with .Author}}<meta property="article:author" content="{{.}}">{{end}}
//	{{range .Tags}}<meta property="article:tag" content="{{.}}">{{end}}
//	{{end}}
//
// The values a post may not have — the last-edited date, the author, the tags,
// and the reading time when disabled — are empty rather than absent, so
// templates gate each tag on its own value.
//
// # Series
//
// Series are named, ordered collections of posts — a multi-part tutorial, say —
// defined in one site-wide YAML file rather than in post front matter. They are
// opt-in: supply [config.WithSeriesFile] to enable them.
//
//	gen := generator.New(fsys, renderer,
//	    config.WithSeriesFile(fsys, "series.yml"),
//	)
//
// The file lists each series with a name, an optional slug and description, and
// the posts it contains, named by their filename relative to the posts directory
// and listed in reading order:
//
//	series:
//	  - name: "Building a Blog in Go"
//	    description: "A step-by-step guide."
//	    posts:
//	      - building-blog-part-1.md
//	      - building-blog-part-2.md
//
// When series are enabled, GeneratedBlog.Series holds one rendered page per
// series (keyed by slug) and GeneratedBlog.SeriesIndex the page listing them all
// in file order. Post pages additionally receive a Series field
// ([models.PostSeries]) naming the series, the post's 1-based position in it, and
// the previous and next parts; it is nil for a post in no series. Every page
// receives BaseData.SeriesEnabled so shared layouts can show a "Series" nav link.
//
// Validation is strict, because a silently dropped series is worse than a failed
// build: Generate returns an error — naming the series and the offending value —
// when the YAML is malformed or has no top-level "series" key, a series has no
// name or no posts, a listed filename matches no post, a post appears twice in a
// series or in two series, two series share a slug, or the file contains an
// unknown key. Series are independent of tags, so [config.WithDisableTags] does
// not affect them, and a series post keeps its tags. The index page is
// unaffected: series posts still appear there in date order.
//
// Series are not generated under [config.WithRawOutput], which bypasses
// templates entirely.
//
// # Output
//
// The Generator returns all generated content in memory via GeneratedBlog.
// Callers are responsible for I/O operations such as writing
// files to disk or serving content via HTTP.
//
// See the examples_test.go file for more usage examples.
//
// # Concurrency
//
// The Generator is safe for concurrent use once created, though Generate
// operations should not be run concurrently on the same Generator instance.
package generator

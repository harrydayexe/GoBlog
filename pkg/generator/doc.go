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

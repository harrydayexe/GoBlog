package config

// GeneratorOption represents a configuration option that can be applied to
// generator or outputter instances during construction.
//
// Options use the functional options pattern, where each option function
// returns an GeneratorOption struct containing one or more function pointers that
// modify specific configuration fields.
//
// This type should not be constructed directly by users. Instead, use the
// provided option functions like WithRawOutput(), WithDisableTags(),
// WithDisableReadingTime(), WithSiteTitle(), WithEnvironment(), WithCustomData(),
// and WithBaseOption().
type GeneratorOption struct {
	BaseOption

	WithRawOutputFunc          func(v *RawOutput)
	WithDisableTagsFunc        func(v *DisableTags)
	WithDisableReadingTimeFunc func(v *DisableReadingTime)
	WithSiteTitleFunc          func(v *SiteTitle)
	WithEnvironmentFunc        func(v *Environment)
	WithCustomDataFunc         func(v *CustomData)
	WithHTMLPathsFunc          func(v *HTMLPaths)
}

// WithBaseOption wraps a BaseOption as a GeneratorOption so it can be passed
// to generator constructors that accept GeneratorOption values.
// Use this when you have a BaseOption (e.g. from WithBlogRoot) and need to
// supply it alongside other GeneratorOptions.
func WithBaseOption(baseOption BaseOption) GeneratorOption {
	return GeneratorOption{
		BaseOption: baseOption,
	}
}

// RawOutput is a configuration type that controls whether HTML output
// is generated with or without template wrapping.
//
// When RawOutput is true:
//   - The generator produces only Markdown-to-HTML conversion without templates
//   - The outputter skips creating the tags directory
//   - Individual post files contain raw HTML fragments
//
// This type is typically embedded in generator and outputter configuration
// structs and should be set using the WithRawOutput() option function.
type RawOutput struct{ RawOutput bool }

// WithRawOutput returns an Option that enables raw HTML output mode.
//
// When this option is applied to a generator, it will produce HTML content
// without template wrapping - only the Markdown-to-HTML conversion is performed.
// When applied to an outputter, it will skip creating the tags directory.
//
// This is useful for scenarios where you want to integrate GoBlog's HTML
// output into your own templates or existing site structure.
//
// Example usage:
//
//	gen := generator.New(fsys, nil, config.WithRawOutput())
//	writer := outputter.NewDirectoryWriter("output/", config.WithRawOutput())
//
// When using WithRawOutput on the generator, no template renderer is needed
// because templates are bypassed entirely; pass nil as the renderer.
func WithRawOutput() GeneratorOption {
	return GeneratorOption{
		WithRawOutputFunc: func(v *RawOutput) {
			v.RawOutput = true
		},
	}
}

func (o RawOutput) AsOption() GeneratorOption {
	if bool(o.RawOutput) {
		return WithRawOutput()
	}
	return GeneratorOption{
		WithRawOutputFunc: func(v *RawOutput) {
			v.RawOutput = false
		},
	}
}

// DisableTags is a configuration type that controls whether tag pages are
// generated and served.
//
// When Disable is true:
//   - The generator skips rendering individual tag pages and the tags index
//   - Post.Tags slices are cleared in the assembled blog so the default templates do not render tag pills
//   - The outputter skips creating the tags directory
//   - The HTTP server skips registering /tags routes
//   - BaseData.TagsEnabled is set to false for all templates
//
// This type is typically embedded in generator, outputter, and server
// configuration structs and should be set using the WithDisableTags() option function.
type DisableTags struct{ Disable bool }

// WithDisableTags returns a GeneratorOption that disables all tag-related output.
//
// When applied to a generator, it will skip rendering tag pages and the tags
// index. Post tag slices are cleared so that the default templates do not
// render per-post tag pills. When applied to an outputter, it will skip creating the tags
// directory. When applied to the HTTP server, /tags routes are not registered.
//
// Example usage:
//
//	gen := generator.New(fsys, renderer, config.WithDisableTags())
//	writer := outputter.NewDirectoryWriter("output/", config.WithDisableTags())
func WithDisableTags() GeneratorOption {
	return GeneratorOption{
		WithDisableTagsFunc: func(v *DisableTags) {
			v.Disable = true
		},
	}
}

func (o DisableTags) AsOption() GeneratorOption {
	if o.Disable {
		return WithDisableTags()
	}
	return GeneratorOption{
		WithDisableTagsFunc: func(v *DisableTags) {
			v.Disable = false
		},
	}
}

// DisableReadingTime is a configuration type that controls whether estimated
// reading times are computed and surfaced on post pages.
//
// When Disable is true:
//   - Post.ReadingTimeMinutes is left at zero for all posts
//   - The default templates omit the "· N min read" annotation next to the date
//
// This type is typically embedded in generator configuration structs and should
// be set using the WithDisableReadingTime() option function.
type DisableReadingTime struct{ Disable bool }

// WithDisableReadingTime returns a GeneratorOption that disables reading time
// estimation on posts.
//
// When applied to a generator, Post.ReadingTimeMinutes will remain zero for all
// posts. The default templates guard the "· N min read" annotation with
// {{if .Post.ReadingTimeMinutes}}, so setting this option suppresses the display
// without requiring template changes.
//
// Example usage:
//
//	gen := generator.New(fsys, renderer, config.WithDisableReadingTime())
func WithDisableReadingTime() GeneratorOption {
	return GeneratorOption{
		WithDisableReadingTimeFunc: func(v *DisableReadingTime) {
			v.Disable = true
		},
	}
}

func (o DisableReadingTime) AsOption() GeneratorOption {
	if o.Disable {
		return WithDisableReadingTime()
	}
	return GeneratorOption{
		WithDisableReadingTimeFunc: func(v *DisableReadingTime) {
			v.Disable = false
		},
	}
}

// SiteTitle is a configuration type that holds the site's title.
//
// This type is typically embedded in generator configuration structs
// and should be set using the WithSiteTitle() option function.
type SiteTitle struct{ SiteTitle string }

// WithSiteTitle returns an Option that sets the site title.
//
// The site title is used in generated HTML pages and templates.
//
// Example usage:
//
//	gen := generator.New(fsys, renderer, config.WithSiteTitle("My Blog"))
func WithSiteTitle(title string) GeneratorOption {
	return GeneratorOption{
		WithSiteTitleFunc: func(v *SiteTitle) {
			v.SiteTitle = title
		},
	}
}

func (o SiteTitle) AsOption() GeneratorOption {
	return WithSiteTitle(o.SiteTitle)
}

// Environment is a configuration type holding the runtime environment name
// (e.g. "local", "test", "production"). It is exposed to templates via
// models.BaseData.Environment so users can branch on environment.
type Environment struct{ Environment string }

// WithEnvironment returns a GeneratorOption that sets the runtime environment
// surfaced to all page templates via models.BaseData.Environment. Callers are
// responsible for supplying a validated value (e.g. via gowebutilities
// config.ParseConfig with EnvironmentConfig).
//
// Example usage:
//
//	gen := generator.New(fsys, renderer, config.WithEnvironment("production"))
func WithEnvironment(env string) GeneratorOption {
	return GeneratorOption{
		WithEnvironmentFunc: func(v *Environment) {
			v.Environment = env
		},
	}
}

func (o Environment) AsOption() GeneratorOption {
	return WithEnvironment(o.Environment)
}

// HTMLPaths is a configuration type that controls whether BaseData.Path values
// are emitted with a .html file extension.
//
// When Enable is true:
//   - Index page path becomes /index.html (BlogRoot "/") or /<root>.html (other roots)
//   - Post, tag, and tags-index paths have .html appended
//
// This type is typically embedded in generator configuration structs and should
// be set using the WithHTMLPaths() option function.
type HTMLPaths struct{ Enable bool }

// WithHTMLPaths returns a GeneratorOption that switches BaseData.Path to use
// .html-suffixed file paths instead of clean URLs.
//
// This is automatically applied by the goblog generate CLI so that canonical
// URLs in static output match the actual .html files written to disk. Library
// users serving content via pkg/server should leave this option off; the server
// accepts both clean URLs and .html URLs via its built-in StripHTMLExtension
// middleware.
//
// Path values with this option enabled:
//
//	BlogRoot = "/"            BlogRoot = "/blog/"
//	Index:   /index.html      /blog.html
//	Post:    /posts/slug.html  /blog/posts/slug.html
//	Tag:     /tags/go.html    /blog/tags/go.html
//	TagsIdx: /tags.html       /blog/tags.html
//
// Example usage:
//
//	gen := generator.New(fsys, renderer, config.WithHTMLPaths())
func WithHTMLPaths() GeneratorOption {
	return GeneratorOption{
		WithHTMLPathsFunc: func(v *HTMLPaths) {
			v.Enable = true
		},
	}
}

// AsOption converts this HTMLPaths value back into a GeneratorOption.
func (o HTMLPaths) AsOption() GeneratorOption {
	if o.Enable {
		return WithHTMLPaths()
	}
	return GeneratorOption{
		WithHTMLPathsFunc: func(v *HTMLPaths) {
			v.Enable = false
		},
	}
}

// CustomData is a configuration type that holds arbitrary key-value data
// surfaced to all page templates via [github.com/harrydayexe/GoBlog/v2/pkg/models.BaseData].Custom.
//
// This type is typically embedded in generator configuration structs and
// should be set using the [WithCustomData] option function.
type CustomData struct{ Data map[string]any }

// WithCustomData returns a GeneratorOption that merges the supplied map into
// the custom data available to all page templates as {{ .Custom.key }}.
//
// The map is exposed on [github.com/harrydayexe/GoBlog/v2/pkg/models.BaseData].Custom
// and is available in every rendered page: post, index, tag, and tags-index.
//
// Multiple calls to WithCustomData accumulate: keys are merged in the order
// the options are applied, with later values overwriting earlier ones for the
// same key. Templates should guard access with {{ with .Custom }} or
// {{ if .Custom }} when the field may be nil (i.e. when no WithCustomData
// option was supplied).
//
// The same map instance is shared across all pages rendered in a single
// Generate call. Callers must not mutate the map after passing it to
// WithCustomData. Values stored in the map must be safe for concurrent reads
// because the HTTP server may render multiple pages in parallel.
//
// # Security
//
// Store only plain, immutable values (strings, numbers, booleans) in the map.
// Do not pre-wrap values in [html/template.HTML], [html/template.JS],
// [html/template.JSStr], [html/template.URL], [html/template.CSS], or
// [html/template.HTMLAttr]: those types signal to html/template that the value
// is already safe and bypass contextual auto-escaping, creating an XSS sink if
// the value originates from user-controlled input. Let html/template escape
// values at render time instead.
//
// Do not construct custom data from untrusted input at request time. This map
// is intended for static, developer-controlled values only.
//
// Example usage:
//
//	gen := generator.New(fsys, renderer,
//	    config.WithCustomData(map[string]any{
//	        "author":      "Jane Smith",
//	        "analyticsID": "UA-12345",
//	    }),
//	)
//
// In a template:
//
//	{{ with .Custom }}
//	    <meta name="author" content="{{ .author }}">
//	{{ end }}
func WithCustomData(data map[string]any) GeneratorOption {
	return GeneratorOption{
		WithCustomDataFunc: func(v *CustomData) {
			if v.Data == nil {
				v.Data = map[string]any{}
			}
			for k, val := range data {
				v.Data[k] = val
			}
		},
	}
}

// AsOption converts this CustomData value back into a GeneratorOption so it
// can be passed to generator and server constructors that accept GeneratorOption
// values (e.g. when round-tripping through a ServerConfig).
func (o CustomData) AsOption() GeneratorOption {
	return WithCustomData(o.Data)
}

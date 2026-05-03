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
// WithSiteTitle(), WithEnvironment(), and WithBaseOption().
type GeneratorOption struct {
	BaseOption

	WithRawOutputFunc   func(v *RawOutput)
	WithDisableTagsFunc func(v *DisableTags)
	WithSiteTitleFunc   func(v *SiteTitle)
	WithEnvironmentFunc func(v *Environment)
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

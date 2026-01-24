package config

// Option represents a configuration option that can be applied to
// generator or outputter instances during construction.
//
// Options use the functional options pattern, where each option function
// returns an Option struct containing one or more function pointers that
// modify specific configuration fields.
//
// This type should not be constructed directly by users. Instead, use the
// provided option functions like WithRawOutput() and WithTemplatesDir().
type Option struct {
	WithRawOutputFunc func(v *RawOutput)
	WithSiteTitleFunc func(v *SiteTitle)
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
//	gen := generator.New(fsys, config.WithRawOutput())
//	writer := outputter.NewDirectoryWriter("output/", config.WithRawOutput())
func WithRawOutput() Option {
	return Option{
		WithRawOutputFunc: func(v *RawOutput) {
			v.RawOutput = true
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
func WithSiteTitle(title string) Option {
	return Option{
		WithSiteTitleFunc: func(v *SiteTitle) {
			v.SiteTitle = title
		},
	}
}

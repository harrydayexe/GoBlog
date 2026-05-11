package config

import "html/template"

// RendererOption represents a configuration option that can be applied to a
// [github.com/harrydayexe/GoBlog/v2/pkg/generator.TemplateRenderer] during
// construction via [github.com/harrydayexe/GoBlog/v2/pkg/generator.NewTemplateRenderer].
//
// Options use the functional options pattern: each value carries a function
// pointer that modifies a specific renderer setting.
//
// This type should not be constructed directly. Use the provided option
// functions such as [WithFuncs].
type RendererOption struct {
	WithFuncsFunc func(fm *template.FuncMap)
}

// WithFuncs returns a RendererOption that merges the supplied functions into
// the template FuncMap used by [github.com/harrydayexe/GoBlog/v2/pkg/generator.NewTemplateRenderer].
//
// The built-in helpers available to all templates are:
//
//	formatDate(t time.Time) string   formats t as "January 2, 2006"
//	shortDate(t time.Time) string    formats t as "Jan 2, 2006"
//	year() int                       returns the current calendar year
//
// If a key in funcs matches one of those built-in names, the supplied function
// silently replaces the built-in. This allows intentional overrides (for
// example, substituting your own date format), but will also silently break
// the default templates if done accidentally. Check against the list above
// before registering a function to avoid unintentional collisions.
//
// Multiple calls to WithFuncs accumulate: functions are merged in the order
// the options are applied, with later registrations overwriting earlier ones
// for the same key.
//
// Example usage:
//
//	import (
//	    "strings"
//	    "html/template"
//	    "github.com/harrydayexe/GoBlog/v2/pkg/config"
//	    "github.com/harrydayexe/GoBlog/v2/pkg/generator"
//	    "github.com/harrydayexe/GoBlog/v2/pkg/templates"
//	)
//
//	renderer, err := generator.NewTemplateRenderer(
//	    templates.Default,
//	    config.WithFuncs(template.FuncMap{
//	        "upper": strings.ToUpper,
//	    }),
//	)
func WithFuncs(funcs template.FuncMap) RendererOption {
	return RendererOption{
		WithFuncsFunc: func(fm *template.FuncMap) {
			for name, fn := range funcs {
				(*fm)[name] = fn
			}
		},
	}
}

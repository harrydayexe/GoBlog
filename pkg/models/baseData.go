package models

// BaseData contains common data available to all templates.
// This data is included in all page renders.
type BaseData struct {
	// SiteTitle is the name of the blog site.
	// Example: "My Awesome Blog"
	SiteTitle string

	// PageTitle is the title for this specific page.
	// Used in <title> tag and may be shown in header.
	// Example: "How to Use Go Templates"
	PageTitle string

	// Description is the meta description for SEO.
	// Should be 150-160 characters.
	Description string

	// Year is the current year, useful for copyright notices.
	// Example: 2026
	Year int

	// BlogRoot is the path to the root of the blog, useful for hyper links
	// Default: "/"
	// Example: "/blog/"
	BlogRoot string

	// Environment is the runtime environment ("local", "test", or "production").
	// Set via the ENVIRONMENT env var (default "local").
	// Use in templates to gate environment-specific markup:
	//   {{if eq .Environment "production"}}<script src="/analytics.js"></script>{{end}}
	Environment string

	// TagsEnabled indicates whether tag features are active for this blog.
	// When false, the default templates suppress tag-related navigation links
	// and per-post tag pills. Custom templates should also gate tag UI on this
	// field. The Go zero value is false; the Generator sets this to true unless
	// config.WithDisableTags() is applied, so manual constructors must set it
	// explicitly when tags should be visible.
	//
	// Custom templates should gate tag UI on this field:
	//   {{if .TagsEnabled}}<a href="{{.BlogRoot}}tags">Tags</a>{{end}}
	TagsEnabled bool

	// Custom holds arbitrary key-value data injected by the calling application
	// via config.WithCustomData. It is nil when no custom data was configured.
	//
	// Templates should guard access to avoid nil-map panics:
	//   {{with .Custom}}<meta name="author" content="{{.author}}">{{end}}
	//
	// The same map is shared across every page rendered in a single Generate
	// call. Do not mutate the map from inside a template or after passing it
	// to config.WithCustomData.
	//
	// Security: values in this map should be plain strings, numbers, or
	// booleans. Do not store html/template.HTML, html/template.JS, or other
	// pre-sanitised wrapper types — those bypass contextual auto-escaping and
	// become XSS sinks if the underlying value is user-controlled.
	Custom map[string]any
}

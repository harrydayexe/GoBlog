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

	// FeedsEnabled indicates whether RSS and Atom feeds are available for this blog.
	// When true, the default templates render feed discovery <link> tags in the
	// <head> and a visible RSS navigation link. Custom templates should also gate
	// feed UI on this field.
	//
	// The Go zero value is false. The Generator sets this to true when both a
	// base URL is configured (via config.WithBaseURL) and feeds have not been
	// disabled (via config.WithDisableFeeds).
	//
	// Custom templates should gate feed UI on this field:
	//   {{if .FeedsEnabled}}<a href="{{.BlogRoot}}rss.xml">RSS</a>{{end}}
	FeedsEnabled bool

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

	// Path is the site-relative path of this page, including the BlogRoot prefix.
	// Use it to construct canonical URLs or Open Graph meta tags without having
	// to re-concatenate BlogRoot in the template.
	//
	// By default (clean-URL mode) the path has no .html extension:
	//
	//   Examples (BlogRoot = "/blog/"):
	//     Index page:       /blog/
	//     Post page:        /blog/posts/my-first-post
	//     Tag page:         /blog/tags/golang
	//     Tags index:       /blog/tags
	//
	//   Examples (default BlogRoot = "/"):
	//     Index page:       /
	//     Post page:        /posts/my-first-post
	//     Tag page:         /tags/golang
	//     Tags index:       /tags
	//
	// When config.WithHTMLPaths() is applied (used automatically by goblog
	// generate so paths match the .html files written to disk) the extension is
	// included:
	//
	//   Examples (BlogRoot = "/blog/"):
	//     Index page:       /blog.html
	//     Post page:        /blog/posts/my-first-post.html
	//     Tag page:         /blog/tags/golang.html
	//     Tags index:       /blog/tags.html
	//
	//   Examples (default BlogRoot = "/"):
	//     Index page:       /index.html
	//     Post page:        /posts/my-first-post.html
	//     Tag page:         /tags/golang.html
	//     Tags index:       /tags.html
	//
	// Typical usage for an Open Graph URL tag:
	//   <meta property="og:url" content="https://example.com{{.Path}}">
	Path string

	// CanonicalURL is the fully-qualified URL of this page: the site's base URL
	// (config.WithBaseURL) joined with Path.
	//
	// It is empty when no base URL is configured, because a canonical URL
	// cannot be derived from a site-relative path alone. Templates must guard
	// on it so that no empty-valued tag is emitted:
	//
	//   {{if .CanonicalURL}}
	//   <link rel="canonical" href="{{.CanonicalURL}}">
	//   <meta property="og:url" content="{{.CanonicalURL}}">
	//   {{end}}
	//
	// Because it is built from Path, it follows the same clean-URL or
	// .html-suffixed form (see Path and config.WithHTMLPaths).
	//
	//   Examples (BaseURL = "https://example.com", BlogRoot = "/"):
	//     Index page:       https://example.com/
	//     Post page:        https://example.com/posts/my-first-post
	//     Tag page:         https://example.com/tags/golang
	//     Tags index:       https://example.com/tags
	CanonicalURL string

	// OGType is the Open Graph object type for this page, emitted as the
	// og:type meta tag. The Generator sets it to "article" for post pages and
	// "website" for the index, tag, and tags-index pages.
	//
	// The Go zero value is the empty string, so manual constructors that do not
	// set it get no og:type at all. The default templates fall back to
	// "website" in that case:
	//
	//   <meta property="og:type" content="{{or .OGType "website"}}">
	OGType string

	// Article holds the article-specific metadata for a post page: its
	// publication date, author, and tags. The Generator populates it when
	// rendering a post and leaves it nil for the index, tag, and tags-index
	// pages.
	//
	// It lets a <head> partial shared by every page type render article:* Open
	// Graph tags and Schema.org BlogPosting markup for posts only. Templates
	// must guard on it, both to skip that markup on non-post pages and to avoid
	// a nil-pointer error:
	//
	//   {{with .Article}}
	//   <meta property="article:published_time" content="{{.PublishedISO}}">
	//   {{end}}
	Article *ArticleMeta
}

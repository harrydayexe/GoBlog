package generator

// GeneratedBlog contains all the HTML content for a complete static blog site.
//
// It includes individual post pages, the main index page, and tag pages.
// All content is stored as raw HTML bytes ready to be written to files or
// served via HTTP.
//
// Post slugs are derived from markdown filenames. Tag names are extracted from
// post front matter. For more info see [pkg/models/Post].
//
// # Raw Output Mode
//
// When the generator is configured with config.WithRawOutput(), the content
// in GeneratedBlog will contain only the parsed Markdown as HTML without any
// template wrapping:
//   - Posts map contains clean HTML fragments for each post
//   - Tags map will be empty (tag pages are not generated)
//   - Index field will be empty or contain minimal content
//
// This mode is useful for embedding blog content into existing applications,
// custom CMSs, or when you need to apply your own templates programmatically.
type GeneratedBlog struct {
	Posts map[string][]byte // Posts maps a slug to raw HTML bytes for each post
	Index []byte            // Index contains the raw HTML for the blog index page
	Tags  map[string][]byte // Tags maps each tag name to its tag page HTML
}

func NewEmptyGeneratedBlog() *GeneratedBlog {
	return &GeneratedBlog{
		Posts: make(map[string][]byte),
		Tags:  make(map[string][]byte),
	}
}

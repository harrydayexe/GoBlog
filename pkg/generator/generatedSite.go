package generator

// GeneratedBlog contains all the HTML content for a complete static blog site.
//
// It includes individual post pages, the main index page, tag pages, and tags index.
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
//   - TagsIndex will be empty (tags index is not generated)
//   - Index field will be empty or contain minimal content
//
// This mode is useful for embedding blog content into existing applications,
// custom CMSs, or when you need to apply your own templates programmatically.
//
// # Disable Tags Mode
//
// When the generator is configured with config.WithDisableTags(), the content
// in GeneratedBlog will omit all tag-related output while still applying full
// templates to posts and the index:
//   - Posts map contains fully templated HTML pages (with the default templates, tag pills are not rendered)
//   - Tags map will be empty (tag pages are not generated)
//   - TagsIndex will be nil (tags index is not generated)
//   - Index contains the complete templated index page (with the default templates, the Tags nav link is not rendered)
//
// This mode is useful when the blog content is not taxonomy-driven or when a
// custom navigation structure is used in place of GoBlog's built-in tag pages.
type GeneratedBlog struct {
	Posts     map[string][]byte // Posts maps a slug to raw HTML bytes for each post
	Index     []byte            // Index contains the raw HTML for the blog index page
	Tags      map[string][]byte // Tags maps each tag name to its tag page HTML
	TagsIndex []byte            // TagsIndex contains the raw HTML for the tags index page
}

func NewEmptyGeneratedBlog() *GeneratedBlog {
	return &GeneratedBlog{
		Posts: make(map[string][]byte),
		Tags:  make(map[string][]byte),
	}
}

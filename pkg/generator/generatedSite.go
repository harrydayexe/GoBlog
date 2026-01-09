package generator

// GeneratedBlog contains all the HTML content for a complete static blog site.
//
// It includes individual post pages, the main index page, and tag pages.
// All content is stored as raw HTML bytes ready to be written to files or
// served via HTTP.
//
// Post slugs are derived from markdown filenames. Tag names are extracted from
// post front matter. For more info see [pkg/models/Post].
type GeneratedBlog struct {
	Posts map[string][]byte // Posts maps a slug to raw HTML bytes for each post
	Index []byte            // Index contains the raw HTML for the blog index page
	Tags  map[string][]byte // Tags maps each tag name to its tag page HTML
}

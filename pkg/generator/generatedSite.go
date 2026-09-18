// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

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
//
// # Disable Reading Time Mode
//
// When the generator is configured with config.WithDisableReadingTime(), the
// Post.ReadingTimeMinutes field is left at zero for all posts. The default
// templates guard the "· N min read" annotation with
// {{if .Post.ReadingTimeMinutes}}, so the annotation is simply omitted without
// any other changes to the output structure.
//
// # Feed Mode
//
// When the generator is configured with config.WithBaseURL() (and feeds have
// not been disabled via config.WithDisableFeeds()), the following feed fields
// are populated:
//   - RSSFeed / AtomFeed: site-wide RSS 2.0 / Atom XML bytes
//   - TagRSSFeeds / TagAtomFeeds: per-tag RSS 2.0 / Atom XML, keyed by tag name
//
// Feeds are empty (nil) when:
//   - No base URL is configured (feeds are silently skipped)
//   - config.WithDisableFeeds() was applied
//   - config.WithRawOutput() was applied (raw mode skips all template rendering)
//   - config.WithDisableTags() was applied (tag feeds are additionally empty in this mode)
//
// # Sitemap and robots.txt
//
// When the generator is configured with config.WithBaseURL(), a sitemaps.org
// sitemap and a robots.txt are populated alongside the feeds:
//   - Sitemap: sitemap.xml bytes covering the index, every post, the tags
//     index, and every tag page. Tag URLs are omitted under
//     config.WithDisableTags().
//   - RobotsTxt: robots.txt bytes, holding GoBlog's default rule block (or the
//     body from config.WithRobotsTxt) followed by an absolute "Sitemap:" line.
//
// Sitemap is nil when:
//   - No base URL is configured (the sitemap is silently skipped)
//   - config.WithDisableSitemap() was applied
//   - config.WithRawOutput() was applied (raw mode skips all template rendering)
//   - The blog contains no posts (an empty <urlset> is not a valid sitemap)
//
// RobotsTxt is nil when:
//   - No base URL is configured (robots.txt is silently skipped)
//   - config.WithDisableRobotsTxt() was applied
//   - config.WithRawOutput() was applied
//
// The "Sitemap:" line is dropped from RobotsTxt when
// config.WithDisableSitemap() was applied.
type GeneratedBlog struct {
	Posts     map[string][]byte // Posts maps a slug to raw HTML bytes for each post
	Index     []byte            // Index contains the raw HTML for the blog index page
	Tags      map[string][]byte // Tags maps each tag name to its tag page HTML
	TagsIndex []byte            // TagsIndex contains the raw HTML for the tags index page

	RSSFeed      []byte            // RSSFeed contains the site-wide RSS 2.0 feed XML, or nil if feeds are disabled/skipped
	AtomFeed     []byte            // AtomFeed contains the site-wide Atom feed XML, or nil if feeds are disabled/skipped
	TagRSSFeeds  map[string][]byte // TagRSSFeeds maps each tag name to its RSS 2.0 feed XML
	TagAtomFeeds map[string][]byte // TagAtomFeeds maps each tag name to its Atom feed XML

	Sitemap   []byte // Sitemap contains the sitemap.xml bytes, or nil if the sitemap is disabled/skipped
	RobotsTxt []byte // RobotsTxt contains the robots.txt bytes, or nil if robots.txt is disabled/skipped
}

func NewEmptyGeneratedBlog() *GeneratedBlog {
	return &GeneratedBlog{
		Posts:        make(map[string][]byte),
		Tags:         make(map[string][]byte),
		TagRSSFeeds:  make(map[string][]byte),
		TagAtomFeeds: make(map[string][]byte),
	}
}

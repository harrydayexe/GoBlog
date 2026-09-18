// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package generator

import (
	"encoding/xml"
	"fmt"
	"net/url"
	"sort"
	"time"

	"github.com/harrydayexe/GoBlog/v2/pkg/models"
)

// sitemapNamespace is the XML namespace every sitemaps.org sitemap declares.
const sitemapNamespace = "http://www.sitemaps.org/schemas/sitemap/0.9"

// sitemapURLSet is the <urlset> document root of a sitemap.
type sitemapURLSet struct {
	XMLName xml.Name     `xml:"urlset"`
	XMLNS   string       `xml:"xmlns,attr"`
	URLs    []sitemapURL `xml:"url"`
}

// sitemapURL is a single <url> entry.
//
// Only <loc> and <lastmod> are emitted: Google has confirmed it ignores
// <changefreq> and <priority>, so they would be template surface for no gain.
type sitemapURL struct {
	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod"`
}

// buildSitemap generates a sitemaps.org sitemap covering the blog index, every
// post, the tags index, and every tag page.
//
// posts must already be sorted newest-first. tags maps each tag name to the
// posts carrying it, as used to render the tag pages; it is ignored when
// config.WithDisableTags() is set, so no tag URLs are emitted in that mode.
//
// Every URL is built from [Generator.pagePath] and absolutised with
// [Generator.canonicalURL], so the sitemap always matches the paths actually
// emitted regardless of config.WithHTMLPaths and config.WithBlogRoot.
//
// If posts is empty the sitemap is nil: a <urlset> with no entries is not a
// valid sitemap, and an index with no content is not worth submitting.
func (g *Generator) buildSitemap(posts models.PostList, tags map[string]models.PostList) ([]byte, error) {
	if len(posts) == 0 {
		return nil, nil
	}

	tagsEnabled := !g.DisableTags.Disable

	urls := make([]sitemapURL, 0, len(posts)+len(tags)+2)

	// The index and tags index both list every post, so their last-modified
	// instant is the newest effective-updated time across the whole blog.
	newest := newestUpdated(posts)

	urls = append(urls, sitemapURL{
		Loc:     g.canonicalURL(g.pagePath("index", "")),
		LastMod: w3cDateTime(newest),
	})

	for _, post := range posts {
		urls = append(urls, sitemapURL{
			Loc:     g.canonicalURL(g.pagePath("post", post.Slug)),
			LastMod: w3cDateTime(effectiveUpdated(post)),
		})
	}

	if tagsEnabled && len(tags) > 0 {
		urls = append(urls, sitemapURL{
			Loc:     g.canonicalURL(g.pagePath("tagsIndex", "")),
			LastMod: w3cDateTime(newest),
		})

		// Sort the tag names so the sitemap is byte-for-byte reproducible;
		// PostList.GetAllTags iterates a map and is deliberately unordered.
		names := make([]string, 0, len(tags))
		for tag := range tags {
			names = append(names, tag)
		}
		sort.Strings(names)

		for _, tag := range names {
			urls = append(urls, sitemapURL{
				// Tags come verbatim from front matter, so escape them the
				// same way the tag page's own canonical URL is escaped.
				Loc:     g.canonicalURL(g.pagePath("tag", url.PathEscape(tag))),
				LastMod: w3cDateTime(newestUpdated(tags[tag])),
			})
		}
	}

	doc, err := xml.MarshalIndent(sitemapURLSet{XMLNS: sitemapNamespace, URLs: urls}, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshalling sitemap: %w", err)
	}

	out := make([]byte, 0, len(xml.Header)+len(doc)+1)
	out = append(out, xml.Header...)
	out = append(out, doc...)
	out = append(out, '\n')
	return out, nil
}

// sitemapPath returns the site-relative path the sitemap is published at,
// e.g. "/sitemap.xml" or "/blog/sitemap.xml".
func (g *Generator) sitemapPath() string {
	return string(g.BlogRoot) + "sitemap.xml"
}

// newestUpdated returns the most recent effective-updated time across posts,
// or the zero time when posts is empty.
func newestUpdated(posts models.PostList) time.Time {
	var newest time.Time
	for _, p := range posts {
		if eff := effectiveUpdated(p); eff.After(newest) {
			newest = eff
		}
	}
	return newest
}

// w3cDateTime formats t as a W3C datetime, the format sitemaps.org requires
// for <lastmod>.
func w3cDateTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

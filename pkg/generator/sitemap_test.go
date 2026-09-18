// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package generator

import (
	"context"
	"encoding/xml"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/harrydayexe/GoBlog/v2/pkg/config"
	"github.com/harrydayexe/GoBlog/v2/pkg/models"
	"github.com/harrydayexe/GoBlog/v2/pkg/templates"
)

// ---- helpers ----------------------------------------------------------------

// newSitemapGen constructs a minimal Generator for buildSitemap tests.
func newSitemapGen(t *testing.T, opts ...config.GeneratorOption) *Generator {
	t.Helper()
	return New(fstest.MapFS{}, nil, append([]config.GeneratorOption{
		config.WithBaseURL("https://example.com"),
	}, opts...)...)
}

// parsedSitemap is the minimal shape needed to assert on a marshalled sitemap.
type parsedSitemap struct {
	XMLName xml.Name `xml:"urlset"`
	XMLNS   string   `xml:"xmlns,attr"`
	URLs    []struct {
		Loc        string `xml:"loc"`
		LastMod    string `xml:"lastmod"`
		ChangeFreq string `xml:"changefreq"`
		Priority   string `xml:"priority"`
	} `xml:"url"`
}

// parseSitemap unmarshals raw sitemap bytes, failing the test if they are not
// valid XML.
func parseSitemap(t *testing.T, raw []byte) parsedSitemap {
	t.Helper()
	var doc parsedSitemap
	if err := xml.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("sitemap is not valid XML: %v\n%s", err, raw)
	}
	return doc
}

// locs returns the <loc> values of a parsed sitemap in document order.
func locs(doc parsedSitemap) []string {
	out := make([]string, len(doc.URLs))
	for i, u := range doc.URLs {
		out[i] = u.Loc
	}
	return out
}

// lastModFor returns the <lastmod> recorded for the given <loc>.
func lastModFor(t *testing.T, doc parsedSitemap, loc string) string {
	t.Helper()
	for _, u := range doc.URLs {
		if u.Loc == loc {
			return u.LastMod
		}
	}
	t.Fatalf("sitemap has no entry for %q; got %v", loc, locs(doc))
	return ""
}

// assertSameURLs asserts that got and want hold the same URLs, ignoring order.
func assertSameURLs(t *testing.T, got, want []string) {
	t.Helper()
	seen := make(map[string]int, len(got))
	for _, g := range got {
		seen[g]++
	}
	for _, w := range want {
		if seen[w] == 0 {
			t.Errorf("sitemap missing URL %q", w)
		}
		seen[w]--
	}
	for url, n := range seen {
		if n > 0 {
			t.Errorf("sitemap contains unexpected URL %q (%d extra)", url, n)
		}
	}
}

// sitemapPosts returns two tagged posts, newest-first, the older of which has
// been edited after the newer one was published.
func sitemapPosts() models.PostList {
	return models.PostList{
		{
			Title:       "Newer",
			Slug:        "newer",
			Date:        time.Date(2024, time.June, 1, 12, 0, 0, 0, time.UTC),
			Description: "desc",
			Tags:        []string{"go"},
		},
		{
			Title:       "Older",
			Slug:        "older",
			Date:        time.Date(2024, time.January, 1, 9, 0, 0, 0, time.UTC),
			LastEdited:  time.Date(2024, time.August, 1, 9, 0, 0, 0, time.UTC),
			Description: "desc",
			Tags:        []string{"go", "testing"},
		},
	}
}

// tagListsFor groups posts by tag the same way assembleBlogWithTemplates does.
func tagListsFor(posts models.PostList) map[string]models.PostList {
	tags := make(map[string]models.PostList)
	for _, tag := range posts.GetAllTags() {
		tags[tag] = posts.FilterByTag(tag)
	}
	return tags
}

// ---- buildSitemap tests -----------------------------------------------------

func TestBuildSitemap_EmptyPosts(t *testing.T) {
	t.Parallel()

	gen := newSitemapGen(t)
	got, err := gen.buildSitemap(nil, nil)
	if err != nil {
		t.Fatalf("buildSitemap error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil sitemap for empty post list, got:\n%s", got)
	}
}

func TestBuildSitemap_ValidXMLAndNamespace(t *testing.T) {
	t.Parallel()

	gen := newSitemapGen(t)
	posts := sitemapPosts()
	raw, err := gen.buildSitemap(posts, tagListsFor(posts))
	if err != nil {
		t.Fatalf("buildSitemap error: %v", err)
	}

	if !strings.HasPrefix(string(raw), xml.Header) {
		t.Errorf("sitemap does not start with an XML declaration:\n%s", raw)
	}

	doc := parseSitemap(t, raw)
	if doc.XMLNS != sitemapNamespace {
		t.Errorf("xmlns = %q, want %q", doc.XMLNS, sitemapNamespace)
	}
}

// TestBuildSitemap_OmitsChangeFreqAndPriority asserts the two elements Google
// ignores are never emitted.
func TestBuildSitemap_OmitsChangeFreqAndPriority(t *testing.T) {
	t.Parallel()

	gen := newSitemapGen(t)
	posts := sitemapPosts()
	raw, err := gen.buildSitemap(posts, tagListsFor(posts))
	if err != nil {
		t.Fatalf("buildSitemap error: %v", err)
	}

	for _, elem := range []string{"changefreq", "priority"} {
		if strings.Contains(string(raw), elem) {
			t.Errorf("sitemap contains <%s>, which must not be emitted:\n%s", elem, raw)
		}
	}
}

func TestBuildSitemap_URLSet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		opts []config.GeneratorOption
		want []string
	}{
		{
			name: "clean URLs at domain root",
			want: []string{
				"https://example.com/",
				"https://example.com/posts/newer",
				"https://example.com/posts/older",
				"https://example.com/tags",
				"https://example.com/tags/go",
				"https://example.com/tags/testing",
			},
		},
		{
			name: "html paths",
			opts: []config.GeneratorOption{config.WithHTMLPaths()},
			want: []string{
				"https://example.com/index.html",
				"https://example.com/posts/newer.html",
				"https://example.com/posts/older.html",
				"https://example.com/tags.html",
				"https://example.com/tags/go.html",
				"https://example.com/tags/testing.html",
			},
		},
		{
			name: "blog root",
			opts: []config.GeneratorOption{config.WithBlogRoot("/blog/").AsGeneratorOption()},
			want: []string{
				"https://example.com/blog/",
				"https://example.com/blog/posts/newer",
				"https://example.com/blog/posts/older",
				"https://example.com/blog/tags",
				"https://example.com/blog/tags/go",
				"https://example.com/blog/tags/testing",
			},
		},
		{
			name: "tags disabled",
			opts: []config.GeneratorOption{config.WithDisableTags()},
			want: []string{
				"https://example.com/",
				"https://example.com/posts/newer",
				"https://example.com/posts/older",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gen := newSitemapGen(t, tt.opts...)
			posts := sitemapPosts()
			raw, err := gen.buildSitemap(posts, tagListsFor(posts))
			if err != nil {
				t.Fatalf("buildSitemap error: %v", err)
			}
			assertSameURLs(t, locs(parseSitemap(t, raw)), tt.want)
		})
	}
}

// TestBuildSitemap_LastMod covers all four lastmod sources: a post's own date,
// the LastEdited fallback, the newest-post rollup used by the index and tags
// index, and the newest-in-tag rollup used by each tag page.
func TestBuildSitemap_LastMod(t *testing.T) {
	t.Parallel()

	gen := newSitemapGen(t)
	posts := sitemapPosts()
	raw, err := gen.buildSitemap(posts, tagListsFor(posts))
	if err != nil {
		t.Fatalf("buildSitemap error: %v", err)
	}
	doc := parseSitemap(t, raw)

	const (
		newerDate = "2024-06-01T12:00:00Z" // Newer.Date
		olderEdit = "2024-08-01T09:00:00Z" // Older.LastEdited, the newest instant overall
	)

	tests := []struct {
		loc  string
		want string
	}{
		// Post pages use their own effective-updated time.
		{"https://example.com/posts/newer", newerDate},
		{"https://example.com/posts/older", olderEdit}, // LastEdited beats Date
		// Index and tags index roll up to the newest instant across all posts,
		// which here comes from an edit to the *older* post.
		{"https://example.com/", olderEdit},
		{"https://example.com/tags", olderEdit},
		// "go" is on both posts, so it rolls up to the same newest instant;
		// "testing" is only on the older (edited) post.
		{"https://example.com/tags/go", olderEdit},
		{"https://example.com/tags/testing", olderEdit},
	}

	for _, tt := range tests {
		if got := lastModFor(t, doc, tt.loc); got != tt.want {
			t.Errorf("lastmod for %s = %q, want %q", tt.loc, got, tt.want)
		}
	}
}

// TestBuildSitemap_TagLastModIsPerTag asserts that a tag carried only by an
// older post keeps that post's date rather than the site-wide newest.
func TestBuildSitemap_TagLastModIsPerTag(t *testing.T) {
	t.Parallel()

	posts := models.PostList{
		{
			Title: "Newer", Slug: "newer",
			Date: time.Date(2025, time.March, 3, 0, 0, 0, 0, time.UTC),
			Tags: []string{"new-only"},
		},
		{
			Title: "Older", Slug: "older",
			Date: time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
			Tags: []string{"old-only"},
		},
	}

	gen := newSitemapGen(t)
	raw, err := gen.buildSitemap(posts, tagListsFor(posts))
	if err != nil {
		t.Fatalf("buildSitemap error: %v", err)
	}
	doc := parseSitemap(t, raw)

	if got, want := lastModFor(t, doc, "https://example.com/tags/old-only"), "2020-01-01T00:00:00Z"; got != want {
		t.Errorf("lastmod for /tags/old-only = %q, want %q", got, want)
	}
	if got, want := lastModFor(t, doc, "https://example.com/tags/new-only"), "2025-03-03T00:00:00Z"; got != want {
		t.Errorf("lastmod for /tags/new-only = %q, want %q", got, want)
	}
}

// TestBuildSitemap_EscapesXMLSignificantCharacters asserts that slugs and tags
// containing XML- and URL-significant characters round-trip through the
// sitemap unmangled.
func TestBuildSitemap_EscapesXMLSignificantCharacters(t *testing.T) {
	t.Parallel()

	posts := models.PostList{
		{
			Title: "Tricky", Slug: "a&b<c>d",
			Date: time.Date(2024, time.May, 5, 0, 0, 0, 0, time.UTC),
			Tags: []string{"r&d", "c++ tricks"},
		},
	}

	gen := newSitemapGen(t)
	raw, err := gen.buildSitemap(posts, tagListsFor(posts))
	if err != nil {
		t.Fatalf("buildSitemap error: %v", err)
	}

	// Raw '&' and '<' must have been escaped in the serialised document.
	if strings.Contains(string(raw), "a&b") {
		t.Errorf("sitemap contains an unescaped ampersand:\n%s", raw)
	}

	// After unmarshalling, the escaping is transparent and the URLs match the
	// canonical URLs the pages themselves declare.
	doc := parseSitemap(t, raw)
	assertSameURLs(t, locs(doc), []string{
		"https://example.com/",
		"https://example.com/posts/a&b<c>d",
		"https://example.com/tags",
		"https://example.com/tags/r&d",
		"https://example.com/tags/c++%20tricks", // spaces are path-escaped
	})
}

// TestBuildSitemap_Deterministic asserts repeated builds are byte-identical
// despite the unordered tag map.
func TestBuildSitemap_Deterministic(t *testing.T) {
	t.Parallel()

	gen := newSitemapGen(t)
	posts := sitemapPosts()
	tags := tagListsFor(posts)

	first, err := gen.buildSitemap(posts, tags)
	if err != nil {
		t.Fatalf("buildSitemap error: %v", err)
	}
	for range 10 {
		next, err := gen.buildSitemap(posts, tags)
		if err != nil {
			t.Fatalf("buildSitemap error: %v", err)
		}
		if string(next) != string(first) {
			t.Fatalf("buildSitemap is not deterministic:\n%s\n---\n%s", first, next)
		}
	}
}

// ---- Generate() integration tests -------------------------------------------

// sitemapFS is a posts filesystem with a single tagged post.
func sitemapFS() fstest.MapFS {
	return fstest.MapFS{
		"hello.md": &fstest.MapFile{Data: []byte(taggedPost)},
	}
}

// generateSitemapBlog runs a full Generate with the default templates.
func generateSitemapBlog(t *testing.T, opts ...config.GeneratorOption) *GeneratedBlog {
	t.Helper()
	renderer, err := NewTemplateRenderer(templates.Default)
	if err != nil {
		t.Fatalf("NewTemplateRenderer: %v", err)
	}
	blog, err := New(sitemapFS(), renderer, opts...).Generate(context.Background())
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	return blog
}

func TestGenerate_SitemapNilWithoutBaseURL(t *testing.T) {
	t.Parallel()

	blog := generateSitemapBlog(t)
	if blog.Sitemap != nil {
		t.Errorf("expected nil Sitemap without a base URL, got:\n%s", blog.Sitemap)
	}
	if blog.RobotsTxt != nil {
		t.Errorf("expected nil RobotsTxt without a base URL, got:\n%s", blog.RobotsTxt)
	}
}

func TestGenerate_SitemapPopulatedWithBaseURL(t *testing.T) {
	t.Parallel()

	blog := generateSitemapBlog(t, config.WithBaseURL("https://example.com"))
	if len(blog.Sitemap) == 0 {
		t.Fatal("expected a sitemap when a base URL is configured")
	}
	if len(blog.RobotsTxt) == 0 {
		t.Fatal("expected a robots.txt when a base URL is configured")
	}

	assertSameURLs(t, locs(parseSitemap(t, blog.Sitemap)), []string{
		"https://example.com/",
		"https://example.com/posts/hello-world",
		"https://example.com/tags",
		"https://example.com/tags/go",
		"https://example.com/tags/testing",
	})
}

func TestGenerate_SitemapDisabled(t *testing.T) {
	t.Parallel()

	blog := generateSitemapBlog(t,
		config.WithBaseURL("https://example.com"),
		config.WithDisableSitemap(),
	)
	if blog.Sitemap != nil {
		t.Errorf("expected nil Sitemap with WithDisableSitemap, got:\n%s", blog.Sitemap)
	}
	// robots.txt is still produced, just without a Sitemap: line.
	if len(blog.RobotsTxt) == 0 {
		t.Error("expected robots.txt to survive WithDisableSitemap")
	}
}

func TestGenerate_RobotsTxtDisabled(t *testing.T) {
	t.Parallel()

	blog := generateSitemapBlog(t,
		config.WithBaseURL("https://example.com"),
		config.WithDisableRobotsTxt(),
	)
	if blog.RobotsTxt != nil {
		t.Errorf("expected nil RobotsTxt with WithDisableRobotsTxt, got:\n%s", blog.RobotsTxt)
	}
	if len(blog.Sitemap) == 0 {
		t.Error("expected the sitemap to survive WithDisableRobotsTxt")
	}
}

func TestGenerate_SitemapNilInRawOutput(t *testing.T) {
	t.Parallel()

	blog := generateSitemapBlog(t,
		config.WithBaseURL("https://example.com"),
		config.WithRawOutput(),
	)
	if blog.Sitemap != nil || blog.RobotsTxt != nil {
		t.Errorf("expected nil Sitemap and RobotsTxt in raw output mode, got sitemap=%d robots=%d bytes",
			len(blog.Sitemap), len(blog.RobotsTxt))
	}
}

// TestGenerate_SitemapMatchesRenderedCanonicalURLs guards the desync the
// pagePath/canonicalURL indirection exists to prevent: every post URL in the
// sitemap must be the same string the page itself declares as canonical.
func TestGenerate_SitemapMatchesRenderedCanonicalURLs(t *testing.T) {
	t.Parallel()

	blog := generateSitemapBlog(t,
		config.WithBaseURL("https://example.com"),
		config.WithHTMLPaths(),
		config.WithBlogRoot("/blog/").AsGeneratorOption(),
	)

	want := `<link rel="canonical" href="https://example.com/blog/posts/hello-world.html">`
	if !strings.Contains(string(blog.Posts["hello-world"]), want) {
		t.Fatalf("rendered post is missing %s", want)
	}

	sitemapURLs := locs(parseSitemap(t, blog.Sitemap))
	if !slices.Contains(sitemapURLs, "https://example.com/blog/posts/hello-world.html") {
		t.Errorf("sitemap URLs %v do not match the page's canonical URL", sitemapURLs)
	}
}

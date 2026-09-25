// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package generator

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/harrydayexe/GoBlog/v2/pkg/config"
	"github.com/harrydayexe/GoBlog/v2/pkg/models"
	"github.com/harrydayexe/GoBlog/v2/pkg/templates"
)

// ---- helpers ----------------------------------------------------------------

// seriesPost returns a markdown post file with the given title and date.
func seriesPost(title, date string) *fstest.MapFile {
	return &fstest.MapFile{Data: []byte(fmt.Sprintf(
		"---\ntitle: %s\ndate: %s\ndescription: About %s\ntags: [go]\n---\n\n# %s\n\nBody.",
		title, date, title, title,
	))}
}

// seriesFS returns a posts filesystem holding a three-part series in ascending
// date order, plus one post that is in no series, and the given series file
// contents at "series.yml".
func seriesFS(seriesYAML string) fstest.MapFS {
	return fstest.MapFS{
		"part-1.md":   seriesPost("Part One", "2024-01-01"),
		"part-2.md":   seriesPost("Part Two", "2024-02-01"),
		"part-3.md":   seriesPost("Part Three", "2024-03-01"),
		"solo.md":     seriesPost("Solo", "2024-04-01"),
		"series.yml":  &fstest.MapFile{Data: []byte(seriesYAML)},
		"other.yml":   &fstest.MapFile{Data: []byte("series: []\n")},
		"empty.yml":   &fstest.MapFile{Data: []byte("")},
		"nokey.yml":   &fstest.MapFile{Data: []byte("other: 1\n")},
		"badyaml.yml": &fstest.MapFile{Data: []byte("series: [\n  - name: broken\n")},
	}
}

// threePartSeries is a valid series file listing all three parts in order.
const threePartSeries = `series:
  - name: "Building a Blog in Go"
    description: "A step-by-step guide."
    posts:
      - part-1.md
      - part-2.md
      - part-3.md
`

// generateWithSeries generates the blog from fsys with the series file at
// seriesPath, applying any extra options, and fails the test on error.
func generateWithSeries(t *testing.T, fsys fstest.MapFS, seriesPath string, opts ...config.GeneratorOption) *GeneratedBlog {
	t.Helper()

	blog, err := generateWithSeriesErr(t, fsys, seriesPath, opts...)
	if err != nil {
		t.Fatalf("Generate() error = %v, want nil", err)
	}
	return blog
}

// generateWithSeriesErr is generateWithSeries without the error assertion, for
// the validation tests that expect a failure.
func generateWithSeriesErr(t *testing.T, fsys fstest.MapFS, seriesPath string, opts ...config.GeneratorOption) (*GeneratedBlog, error) {
	t.Helper()

	renderer, err := NewTemplateRenderer(templates.Default)
	if err != nil {
		t.Fatalf("NewTemplateRenderer() error = %v", err)
	}

	if seriesPath != "" {
		opts = append(opts, config.WithSeriesFile(fsys, seriesPath))
	}

	gen := New(fsys, renderer, opts...)
	return gen.Generate(context.Background())
}

// postDataFor renders the blog and returns the PostPageData the generator built
// for the post with the given source filename. It re-runs the pipeline against a
// stub renderer so the test can inspect the data rather than the HTML.
func postDataFor(t *testing.T, fsys fstest.MapFS, seriesPath, filename string) models.PostPageData {
	t.Helper()

	renderer, err := NewTemplateRenderer(templates.Default)
	if err != nil {
		t.Fatalf("NewTemplateRenderer() error = %v", err)
	}

	gen := New(fsys, renderer, config.WithSeriesFile(fsys, seriesPath))

	posts, err := gen.parsePosts(context.Background())
	if err != nil {
		t.Fatalf("parsing posts: %v", err)
	}

	series, err := gen.loadSeries(posts)
	if err != nil {
		t.Fatalf("loadSeries() error = %v", err)
	}

	for _, post := range posts {
		if post.SourcePath == filename {
			return models.PostPageData{
				Post:   post,
				Series: gen.postSeriesContexts(series)[filename],
			}
		}
	}

	t.Fatalf("no post parsed from %q", filename)
	return models.PostPageData{}
}

// ---- disabled ---------------------------------------------------------------

// TestSeries_DisabledByDefault asserts that series are off without the option:
// no pages, SeriesEnabled false, and every post's Series nil (case 1).
func TestSeries_DisabledByDefault(t *testing.T) {
	t.Parallel()

	blog := generateWithSeries(t, seriesFS(threePartSeries), "")

	if len(blog.Series) != 0 {
		t.Errorf("Series = %d pages, want 0 when no series file is configured", len(blog.Series))
	}
	if len(blog.SeriesIndex) != 0 {
		t.Errorf("SeriesIndex = %d bytes, want 0 when no series file is configured", len(blog.SeriesIndex))
	}

	// The rendered pages must carry no series markup at all.
	for slug, page := range blog.Posts {
		if strings.Contains(strings.ToLower(string(page)), "series") {
			t.Errorf("post %q mentions series despite series being disabled", slug)
		}
	}
	if strings.Contains(strings.ToLower(string(blog.Index)), "series") {
		t.Error("index page mentions series despite series being disabled")
	}
}

// TestSeries_RawOutputSkipsSeries asserts that raw mode produces no series pages
// even with a series file configured (case 23, generator half).
func TestSeries_RawOutputSkipsSeries(t *testing.T) {
	t.Parallel()

	fsys := seriesFS(threePartSeries)
	blog, err := generateWithSeriesErr(t, fsys, "series.yml", config.WithRawOutput())
	if err != nil {
		t.Fatalf("Generate() error = %v, want nil", err)
	}

	if len(blog.Series) != 0 || len(blog.SeriesIndex) != 0 {
		t.Errorf("raw output produced %d series pages and a %d-byte index, want none",
			len(blog.Series), len(blog.SeriesIndex))
	}
}

// ---- parsing ----------------------------------------------------------------

// TestSeries_EmptyList asserts that "series: []" enables the feature with no
// series in it: the index renders, but no individual pages (case 2).
func TestSeries_EmptyList(t *testing.T) {
	t.Parallel()

	blog := generateWithSeries(t, seriesFS(threePartSeries), "other.yml")

	if len(blog.Series) != 0 {
		t.Errorf("Series = %d pages, want 0 for an empty series list", len(blog.Series))
	}
	if len(blog.SeriesIndex) == 0 {
		t.Error("SeriesIndex is empty; an empty series list should still render the index")
	}
	if !strings.Contains(string(blog.Index), `href="/series"`) {
		t.Error("index page does not show the Series nav link despite series being enabled")
	}
}

// TestSeries_FileOrderWins asserts that a series page lists its posts in file
// order even when that contradicts their dates (cases 3 and 4).
func TestSeries_FileOrderWins(t *testing.T) {
	t.Parallel()

	// Part 1 is the newest post, so date order would reverse the list.
	reversed := `series:
  - name: "Reverse Chronology"
    posts:
      - part-3.md
      - part-2.md
      - part-1.md
`
	fsys := seriesFS(reversed)
	blog := generateWithSeries(t, fsys, "series.yml")

	if len(blog.Series) != 1 {
		t.Fatalf("Series = %d pages, want 1", len(blog.Series))
	}

	page := string(blog.Series["reverse-chronology"])
	if page == "" {
		t.Fatalf("no page rendered for slug %q; got %v", "reverse-chronology", keysOf(blog.Series))
	}

	want := []string{"Part Three", "Part Two", "Part One"}
	assertOrder(t, page, want)

	// The post positions follow the same order.
	data := postDataFor(t, fsys, "series.yml", "part-3.md")
	if data.Series.Position != 1 {
		t.Errorf("part-3 (newest post, listed first) Position = %d, want 1", data.Series.Position)
	}
}

// TestSeries_SlugDerivedFromName asserts the slug and path are derived from the
// name when none is given (case 5).
func TestSeries_SlugDerivedFromName(t *testing.T) {
	t.Parallel()

	blog := generateWithSeries(t, seriesFS(threePartSeries), "series.yml")

	if _, ok := blog.Series["building-a-blog-in-go"]; !ok {
		t.Errorf("no series page keyed %q; got %v", "building-a-blog-in-go", keysOf(blog.Series))
	}
}

// TestSeries_ExplicitSlug asserts an explicit slug is used for the key and the
// page path (case 6).
func TestSeries_ExplicitSlug(t *testing.T) {
	t.Parallel()

	explicit := `series:
  - name: "Docker for Go Developers"
    slug: docker-go
    posts:
      - part-1.md
`
	blog := generateWithSeries(t, seriesFS(explicit), "series.yml")

	if _, ok := blog.Series["docker-go"]; !ok {
		t.Errorf("no series page keyed %q; got %v", "docker-go", keysOf(blog.Series))
	}
	if !strings.Contains(string(blog.SeriesIndex), "/series/docker-go") {
		t.Error("series index does not link to the explicit slug")
	}
}

// TestSeries_ValidationErrors covers every hard error the series file can
// produce (cases 7-14), asserting each message names what the user must fix.
func TestSeries_ValidationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		yaml     string
		file     string
		wantMsgs []string
	}{
		{
			name:     "malformed YAML",
			file:     "badyaml.yml",
			wantMsgs: []string{"badyaml.yml"},
		},
		{
			name:     "missing top-level series key",
			file:     "nokey.yml",
			wantMsgs: []string{"nokey.yml", "series"},
		},
		{
			name:     "empty file",
			file:     "empty.yml",
			wantMsgs: []string{"empty.yml", "series"},
		},
		{
			name: "missing name",
			yaml: "series:\n  - posts:\n      - part-1.md\n",
			// The series has no name to blame, so the message names its index.
			wantMsgs: []string{"index 0", "no name"},
		},
		{
			name:     "empty posts list",
			yaml:     "series:\n  - name: Lonely\n    posts: []\n",
			wantMsgs: []string{"Lonely", "no posts"},
		},
		{
			name:     "missing posts key",
			yaml:     "series:\n  - name: Lonely\n",
			wantMsgs: []string{"Lonely", "no posts"},
		},
		{
			name:     "unknown post filename",
			yaml:     "series:\n  - name: Ghosts\n    posts:\n      - nope.md\n",
			wantMsgs: []string{"Ghosts", "nope.md"},
		},
		{
			name: "post in two series",
			yaml: "series:\n  - name: First\n    posts:\n      - part-1.md\n" +
				"  - name: Second\n    posts:\n      - part-1.md\n",
			wantMsgs: []string{"First", "Second", "part-1.md"},
		},
		{
			name:     "post twice in one series",
			yaml:     "series:\n  - name: Repeats\n    posts:\n      - part-1.md\n      - part-1.md\n",
			wantMsgs: []string{"Repeats", "part-1.md"},
		},
		{
			name: "colliding derived slugs",
			yaml: "series:\n  - name: \"Go Tips\"\n    posts:\n      - part-1.md\n" +
				"  - name: \"go-tips\"\n    posts:\n      - part-2.md\n",
			wantMsgs: []string{"Go Tips", "go-tips"},
		},
		{
			name: "explicit slug colliding with a derived one",
			yaml: "series:\n  - name: \"Go Tips\"\n    posts:\n      - part-1.md\n" +
				"  - name: \"Other\"\n    slug: go-tips\n    posts:\n      - part-2.md\n",
			wantMsgs: []string{"Other", "go-tips"},
		},
		{
			name:     "slug empty once slugified",
			yaml:     "series:\n  - name: Symbols\n    slug: \"!!!\"\n    posts:\n      - part-1.md\n",
			wantMsgs: []string{"Symbols", "!!!"},
		},
		{
			name:     "name empty once slugified",
			yaml:     "series:\n  - name: \"!!!\"\n    posts:\n      - part-1.md\n",
			wantMsgs: []string{"!!!", "slug"},
		},
		{
			// Asserting on the decoder's "field post not found" wording rather
			// than on "post" alone: a series with no posts is also an error
			// mentioning "post", so a looser assertion would still pass if
			// KnownFields(true) were dropped and "post:" silently ignored.
			name:     "unknown key",
			yaml:     "series:\n  - name: Typo\n    post:\n      - part-1.md\n",
			wantMsgs: []string{"field post", "not found", "series.yml"},
		},
		{
			name:     "derived slug is the reserved index slug",
			yaml:     "series:\n  - name: Index\n    posts:\n      - part-1.md\n",
			wantMsgs: []string{"Index", "index", "reserved"},
		},
		{
			name:     "explicit slug is the reserved index slug",
			yaml:     "series:\n  - name: Overview\n    slug: index\n    posts:\n      - part-1.md\n",
			wantMsgs: []string{"Overview", "index", "reserved"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			file := tt.file
			if file == "" {
				file = "series.yml"
			}

			_, err := generateWithSeriesErr(t, seriesFS(tt.yaml), file)
			if err == nil {
				t.Fatal("Generate() error = nil, want an error")
			}
			for _, want := range tt.wantMsgs {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error %q does not mention %q", err, want)
				}
			}
		})
	}
}

// TestSeries_MissingFile asserts a series file that cannot be read fails
// generation rather than silently disabling series.
func TestSeries_MissingFile(t *testing.T) {
	t.Parallel()

	_, err := generateWithSeriesErr(t, seriesFS(threePartSeries), "does-not-exist.yml")
	if err == nil {
		t.Fatal("Generate() error = nil, want an error for a missing series file")
	}
	if !strings.Contains(err.Error(), "does-not-exist.yml") {
		t.Errorf("error %q does not name the missing file", err)
	}
}

// ---- post page data ---------------------------------------------------------

// TestSeries_PostPageData covers the position, total, and neighbour fields for
// every place a post can sit in a series (cases 16-20).
func TestSeries_PostPageData(t *testing.T) {
	t.Parallel()

	fsys := seriesFS(threePartSeries)

	tests := []struct {
		name         string
		filename     string
		wantSeries   bool
		wantPosition int
		wantPrev     string
		wantNext     string
	}{
		{name: "post in no series", filename: "solo.md", wantSeries: false},
		{name: "first of three", filename: "part-1.md", wantSeries: true, wantPosition: 1, wantNext: "Part Two"},
		{name: "middle of three", filename: "part-2.md", wantSeries: true, wantPosition: 2, wantPrev: "Part One", wantNext: "Part Three"},
		{name: "last of three", filename: "part-3.md", wantSeries: true, wantPosition: 3, wantPrev: "Part Two"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data := postDataFor(t, fsys, "series.yml", tt.filename)

			if !tt.wantSeries {
				if data.Series != nil {
					t.Fatalf("Series = %+v, want nil for a post in no series", data.Series)
				}
				return
			}

			if data.Series == nil {
				t.Fatal("Series = nil, want the series context")
			}
			if data.Series.Position != tt.wantPosition {
				t.Errorf("Position = %d, want %d", data.Series.Position, tt.wantPosition)
			}
			if data.Series.Total != 3 {
				t.Errorf("Total = %d, want 3", data.Series.Total)
			}
			if got := titleOf(data.Series.Prev); got != tt.wantPrev {
				t.Errorf("Prev = %q, want %q", got, tt.wantPrev)
			}
			if got := titleOf(data.Series.Next); got != tt.wantNext {
				t.Errorf("Next = %q, want %q", got, tt.wantNext)
			}
			if data.Series.Name != "Building a Blog in Go" {
				t.Errorf("Name = %q, want %q", data.Series.Name, "Building a Blog in Go")
			}
			if data.Series.Path != "/series/building-a-blog-in-go" {
				t.Errorf("Path = %q, want %q", data.Series.Path, "/series/building-a-blog-in-go")
			}
		})
	}
}

// TestSeries_SinglePostSeries asserts a one-post series has no neighbours
// (case 20).
func TestSeries_SinglePostSeries(t *testing.T) {
	t.Parallel()

	single := "series:\n  - name: Standalone\n    posts:\n      - part-1.md\n"
	data := postDataFor(t, seriesFS(single), "series.yml", "part-1.md")

	if data.Series == nil {
		t.Fatal("Series = nil, want the series context")
	}
	if data.Series.Position != 1 || data.Series.Total != 1 {
		t.Errorf("Position/Total = %d/%d, want 1/1", data.Series.Position, data.Series.Total)
	}
	if data.Series.Prev != nil || data.Series.Next != nil {
		t.Error("Prev/Next are set for a single-post series, want both nil")
	}
}

// ---- paths ------------------------------------------------------------------

// TestSeries_PagePaths covers the series paths under a blog root and with HTML
// paths enabled (cases 24 and 25).
func TestSeries_PagePaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		opts            []config.GeneratorOption
		wantSeries      string
		wantSeriesIndex string
	}{
		{
			name:            "default root, clean URLs",
			wantSeries:      "/series/building-a-blog-in-go",
			wantSeriesIndex: "/series",
		},
		{
			name:            "blog root",
			opts:            []config.GeneratorOption{config.WithBlogRoot("/blog/").AsGeneratorOption()},
			wantSeries:      "/blog/series/building-a-blog-in-go",
			wantSeriesIndex: "/blog/series",
		},
		{
			// The series index is a directory index, so it names index.html
			// rather than taking an extension on the directory itself, exactly
			// as the tags index does.
			name:            "HTML paths",
			opts:            []config.GeneratorOption{config.WithHTMLPaths()},
			wantSeries:      "/series/building-a-blog-in-go.html",
			wantSeriesIndex: "/series/index.html",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gen := New(seriesFS(threePartSeries), nil, tt.opts...)

			if got := gen.pagePath("series", "building-a-blog-in-go"); got != tt.wantSeries {
				t.Errorf("pagePath(series) = %q, want %q", got, tt.wantSeries)
			}
			if got := gen.pagePath("seriesIndex", ""); got != tt.wantSeriesIndex {
				t.Errorf("pagePath(seriesIndex) = %q, want %q", got, tt.wantSeriesIndex)
			}
		})
	}
}

// TestSeries_CanonicalURL asserts series pages get a canonical URL when a base
// URL is configured.
func TestSeries_CanonicalURL(t *testing.T) {
	t.Parallel()

	blog := generateWithSeries(t, seriesFS(threePartSeries), "series.yml",
		config.WithBaseURL("https://example.com"))

	want := `<link rel="canonical" href="https://example.com/series/building-a-blog-in-go">`
	if page := string(blog.Series["building-a-blog-in-go"]); !strings.Contains(page, want) {
		t.Errorf("series page does not contain %q", want)
	}
	if !strings.Contains(string(blog.SeriesIndex), `href="https://example.com/series"`) {
		t.Error("series index page has no canonical URL")
	}
}

// ---- descriptions -----------------------------------------------------------

// TestSeries_Description asserts the series page's meta description is the
// author's description when the series file declares one and a generated
// sentence otherwise, and that the page body only shows the author's.
func TestSeries_Description(t *testing.T) {
	t.Parallel()

	const noDescription = `series:
  - name: "Building a Blog in Go"
    posts:
      - part-1.md
      - part-2.md
`

	tests := []struct {
		name     string
		yaml     string
		wantMeta string
		wantBody bool
	}{
		{
			name:     "declared description",
			yaml:     threePartSeries,
			wantMeta: "A step-by-step guide.",
			wantBody: true,
		},
		{
			name:     "no description",
			yaml:     noDescription,
			wantMeta: "Posts in the Building a Blog in Go series",
			wantBody: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			blog := generateWithSeries(t, seriesFS(tt.yaml), "series.yml")
			page := string(blog.Series["building-a-blog-in-go"])

			for _, want := range []string{
				fmt.Sprintf(`<meta name="description" content="%s">`, tt.wantMeta),
				fmt.Sprintf(`<meta property="og:description" content="%s">`, tt.wantMeta),
			} {
				if !strings.Contains(page, want) {
					t.Errorf("series page does not contain %q", want)
				}
			}

			// The generated sentence is a meta-only fallback: the body shows a
			// description only when the author wrote one.
			body := page[strings.Index(page, "<body"):]
			if got := strings.Contains(body, tt.wantMeta); got != tt.wantBody {
				t.Errorf("body contains %q = %t, want %t", tt.wantMeta, got, tt.wantBody)
			}
		})
	}
}

// ---- independence from tags -------------------------------------------------

// TestSeries_IndependentOfDisableTags asserts series pages are still generated
// with tags disabled, and that series posts keep their tags (case 29).
func TestSeries_IndependentOfDisableTags(t *testing.T) {
	t.Parallel()

	blog := generateWithSeries(t, seriesFS(threePartSeries), "series.yml", config.WithDisableTags())

	if len(blog.Series) != 1 {
		t.Errorf("Series = %d pages with --disable-tags, want 1", len(blog.Series))
	}
	if len(blog.SeriesIndex) == 0 {
		t.Error("SeriesIndex is empty with --disable-tags, want the rendered index")
	}
	if len(blog.Tags) != 0 {
		t.Errorf("Tags = %d pages, want 0 with --disable-tags", len(blog.Tags))
	}
}

// TestSeries_IndexPageUnaffected asserts series posts still appear on the main
// index in date order.
func TestSeries_IndexPageUnaffected(t *testing.T) {
	t.Parallel()

	blog := generateWithSeries(t, seriesFS(threePartSeries), "series.yml")

	// Newest first: Solo (April), Part Three, Part Two, Part One.
	assertOrder(t, string(blog.Index), []string{"Solo", "Part Three", "Part Two", "Part One"})
}

// ---- template requirements --------------------------------------------------

// TestSeries_MissingTemplateFails asserts that enabling series with a template
// tree lacking the series pages fails with an error naming the template, rather
// than silently falling back to the built-in templates.
func TestSeries_MissingTemplateFails(t *testing.T) {
	t.Parallel()

	minimal := fstest.MapFS{
		"pages/post.tmpl":       &fstest.MapFile{Data: []byte(`{{.Post.Title}}`)},
		"pages/index.tmpl":      &fstest.MapFile{Data: []byte(`index`)},
		"pages/tag.tmpl":        &fstest.MapFile{Data: []byte(`{{.Tag}}`)},
		"pages/tags-index.tmpl": &fstest.MapFile{Data: []byte(`tags`)},
		"partials/nothing.tmpl": &fstest.MapFile{Data: []byte(`{{define "nothing"}}{{end}}`)},
	}

	renderer, err := NewTemplateRenderer(minimal)
	if err != nil {
		t.Fatalf("NewTemplateRenderer() error = %v", err)
	}

	fsys := seriesFS(threePartSeries)
	gen := New(fsys, renderer, config.WithSeriesFile(fsys, "series.yml"))

	_, err = gen.Generate(context.Background())
	if err == nil {
		t.Fatal("Generate() error = nil, want an error for the missing series template")
	}
	if !strings.Contains(err.Error(), "series") {
		t.Errorf("error %q does not name the missing series template", err)
	}
}

// ---- small helpers ----------------------------------------------------------

// titleOf returns the post's title, or "" when post is nil, so a nil neighbour
// and a neighbour's title can be compared in one assertion.
func titleOf(post *models.Post) string {
	if post == nil {
		return ""
	}
	return post.Title
}

// keysOf returns the keys of a rendered-page map, for failure messages.
func keysOf(pages map[string][]byte) []string {
	keys := make([]string, 0, len(pages))
	for k := range pages {
		keys = append(keys, k)
	}
	return keys
}

// assertOrder asserts that every want appears in page, in the given order.
func assertOrder(t *testing.T, page string, want []string) {
	t.Helper()

	from := 0
	for _, w := range want {
		i := strings.Index(page[from:], w)
		if i < 0 {
			t.Fatalf("page does not contain %q after offset %d, want the order %v", w, from, want)
		}
		from += i + len(w)
	}
}

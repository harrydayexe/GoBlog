// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package models

// SeriesInfo represents one series on the series index page.
type SeriesInfo struct {
	// Name is the series' display name, as written in the series file.
	Name string
	// Slug is the URL segment for the series, either taken from the series
	// file's slug field or derived from Name.
	Slug string
	// Description is the free text from the series file, empty when the series
	// declares none.
	Description string
	// PostCount is the number of posts in the series.
	PostCount int
	// Path is the site-relative path of the series page, including the blog
	// root, so templates can link to it without rebuilding the URL.
	Path string
}

// SeriesIndexPageData is the data passed to pages/series-index.tmpl by
// generator.TemplateRenderer.RenderSeriesIndex. It lists every series in the
// order they appear in the series file, which is the order the author chose.
type SeriesIndexPageData struct {
	BaseData

	// Series is the list of all series with their post counts and paths.
	Series []SeriesInfo

	// TotalSeries is the total number of series.
	TotalSeries int
}

// SeriesPageData is the data passed to pages/series.tmpl by
// generator.TemplateRenderer.RenderSeries. It shows the posts of a single
// series in reading order.
type SeriesPageData struct {
	BaseData

	// Name is the series' display name.
	Name string

	// Slug is the URL segment for the series.
	Slug string

	// Description is the free text from the series file, empty when the series
	// declares none.
	Description string

	// Posts are the posts in the series in reading order: the first entry is
	// part 1. The order comes from the series file, not from the post dates.
	Posts []*Post

	// PostCount is the number of posts in the series.
	PostCount int
}

// PostSeries is the series context of a single post, carried by
// [PostPageData].Series. It tells a post page which series it belongs to and
// where in that series it sits, so the page can render a "Part N of M" box with
// previous and next links.
//
// A post belongs to at most one series, so there is never more than one of
// these per page.
type PostSeries struct {
	// Name is the series' display name.
	Name string

	// Slug is the URL segment for the series.
	Slug string

	// Description is the free text from the series file, empty when the series
	// declares none.
	Description string

	// Path is the site-relative path of the series page, including the blog
	// root.
	Path string

	// Posts are all posts in the series in reading order, so a post page can
	// render the whole table of contents.
	Posts []*Post

	// Position is the 1-based position of the current post within the series.
	Position int

	// Total is the number of posts in the series.
	Total int

	// Prev is the post before this one, nil for the first part.
	Prev *Post

	// Next is the post after this one, nil for the last part.
	Next *Post
}

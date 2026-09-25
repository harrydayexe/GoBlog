// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package models

// PostPageData is the data passed to pages/post.tmpl by
// generator.TemplateRenderer.RenderPost. It renders a single blog post with
// all its metadata.
type PostPageData struct {
	BaseData

	// Post is the blog post to display.
	// See models.Post for available fields.
	Post *Post

	// Series is the post's series context: the series it belongs to, its
	// position within it, and the previous and next parts. See PostSeries.
	//
	// It is nil when the post is in no series and whenever series are disabled
	// (i.e. no config.WithSeriesFile option was supplied), so templates must
	// guard on it:
	//
	//   {{with .Series}}
	//   Part {{.Position}} of {{.Total}} in <a href="{{.Path}}">{{.Name}}</a>
	//   {{end}}
	Series *PostSeries
}

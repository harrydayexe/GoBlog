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
}

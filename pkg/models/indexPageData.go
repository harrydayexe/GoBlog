// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package models

// IndexPageData is the data passed to pages/index.tmpl by
// generator.TemplateRenderer.RenderIndex. It shows a list of recent blog posts.
type IndexPageData struct {
	BaseData

	// Posts is the list of posts to display, typically sorted by date descending.
	// Use range to iterate: {{range .Posts}}...{{end}}
	Posts PostList

	// TotalPosts is the total number of posts in the blog.
	TotalPosts int
}

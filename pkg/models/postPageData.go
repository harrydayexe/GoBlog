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

package models

// PostPageData is passed to the post page template.
// It renders a single blog post with all its metadata.
type PostPageData struct {
	BaseData

	// Post is the blog post to display.
	// See models.Post for available fields.
	Post *Post
}

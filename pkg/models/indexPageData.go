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

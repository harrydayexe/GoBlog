package models

// TagPageData is the data passed to pages/tag.tmpl by
// generator.TemplateRenderer.RenderTag. It shows all posts with a specific tag.
type TagPageData struct {
	BaseData

	// Tag is the name of the tag being displayed.
	// Example: "golang", "tutorial", "web-development"
	Tag string

	// Posts is the list of posts with this tag.
	Posts []*Post

	// PostCount is the number of posts with this tag.
	PostCount int
}

package models

// TagInfo represents information about a single tag.
type TagInfo struct {
	// Name is the tag name.
	Name string
	// PostCount is the number of posts with this tag.
	PostCount int
}

// TagsIndexPageData contains data for rendering the tags index page.
type TagsIndexPageData struct {
	BaseData
	// Tags is the list of all tags with their post counts.
	Tags []TagInfo
	// TotalTags is the total number of tags.
	TotalTags int
}

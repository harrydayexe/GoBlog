package models

// BaseData contains common data available to all templates.
// This data is included in all page renders.
type BaseData struct {
	// SiteTitle is the name of the blog site.
	// Example: "My Awesome Blog"
	SiteTitle string

	// PageTitle is the title for this specific page.
	// Used in <title> tag and may be shown in header.
	// Example: "How to Use Go Templates"
	PageTitle string

	// Description is the meta description for SEO.
	// Should be 150-160 characters.
	Description string

	// Year is the current year, useful for copyright notices.
	// Example: 2026
	Year int
}

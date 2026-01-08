package models

import (
	"fmt"
	"html/template"
	"path/filepath"
	"strings"
	"time"
)

// Post represents a blog post with metadata and content
type Post struct {
	// Frontmatter fields
	Title       string    `yaml:"title"`
	Date        time.Time `yaml:"date"`
	Description string    `yaml:"description"`
	Tags        []string  `yaml:"tags"`

	// Generated fields
	Slug        string        // URL-friendly identifier
	Content     string        // Rendered HTML content
	HTMLContent template.HTML // HTML content for templates (not escaped)
	RawContent  string        // Original markdown content
	SourcePath  string        // Path to source markdown file
	PublishDate time.Time     // Formatted publish date
}

// Validate checks if the post has all required fields
func (p *Post) Validate() error {
	if p.Title == "" {
		return fmt.Errorf("post missing required field: title (source: %s)", p.SourcePath)
	}

	if p.Date.IsZero() {
		return fmt.Errorf("post missing required field: date (source: %s)", p.SourcePath)
	}

	if p.Description == "" {
		return fmt.Errorf("post missing required field: description (source: %s)", p.SourcePath)
	}

	return nil
}

// GenerateSlug creates a URL-friendly slug from the title or filename
func (p *Post) GenerateSlug() {
	if p.Slug != "" {
		return // Already set
	}

	// If we have a title, use it
	if p.Title != "" {
		p.Slug = slugify(p.Title)
		return
	}

	// Fall back to filename without extension
	if p.SourcePath != "" {
		filename := filepath.Base(p.SourcePath)
		p.Slug = slugify(strings.TrimSuffix(filename, filepath.Ext(filename)))
	}
}

// slugify converts a string to a URL-friendly slug
func slugify(s string) string {
	s = strings.ToLower(s)

	// Replace common separators with hyphens
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")

	// Remove any character that's not alphanumeric or hyphen
	var result strings.Builder
	for _, char := range s {
		if (char >= 'a' && char <= 'z') ||
			(char >= '0' && char <= '9') ||
			char == '-' {
			result.WriteRune(char)
		}
	}

	// Remove consecutive hyphens
	slug := result.String()
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}

	// Trim hyphens from start and end
	slug = strings.Trim(slug, "-")

	return slug
}

// HasTag checks if the post has a specific tag
func (p *Post) HasTag(tag string) bool {
	for _, t := range p.Tags {
		if strings.EqualFold(t, tag) {
			return true
		}
	}
	return false
}

// FormattedDate returns the date in a human-readable format
func (p *Post) FormattedDate() string {
	return p.Date.Format("January 2, 2006")
}

// ShortDate returns the date in YYYY-MM-DD format
func (p *Post) ShortDate() string {
	return p.Date.Format("2006-01-02")
}

// PostList is a collection of posts with helper methods
type PostList []*Post

// FilterByTag returns posts that have the specified tag
func (pl PostList) FilterByTag(tag string) PostList {
	var filtered PostList
	for _, post := range pl {
		if post.HasTag(tag) {
			filtered = append(filtered, post)
		}
	}
	return filtered
}

// SortByDate sorts posts by date (newest first)
func (pl PostList) SortByDate() {
	// Simple bubble sort - fine for blog posts
	for i := range pl {
		for j := i + 1; j < len(pl); j++ {
			if pl[i].Date.Before(pl[j].Date) {
				pl[i], pl[j] = pl[j], pl[i]
			}
		}
	}
}

// GetAllTags returns a unique list of all tags across all posts
func (pl PostList) GetAllTags() []string {
	tagSet := make(map[string]bool)
	for _, post := range pl {
		for _, tag := range post.Tags {
			tagSet[tag] = true
		}
	}

	tags := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		tags = append(tags, tag)
	}
	return tags
}

package models_test

import (
	"fmt"
	"time"

	"github.com/harrydayexe/GoBlog/v2/pkg/models"
)

func Example() {
	// Create a new blog post
	post := &models.Post{
		Title:       "Getting Started with Go",
		Date:        time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC),
		Description: "An introduction to Go programming",
		Tags:        []string{"go", "tutorial"},
		Content:     []byte("<p>Welcome to Go programming!</p>"),
	}

	// Generate a URL-friendly slug
	post.GenerateSlug()

	fmt.Println(post.Title)
	fmt.Println(post.Slug)
	fmt.Println(post.FormattedDate())
	// Output: Getting Started with Go
	// getting-started-with-go
	// March 15, 2024
}

func ExamplePost_Validate() {
	// Valid post
	validPost := &models.Post{
		Title:       "My Post",
		Date:        time.Now(),
		Description: "A description",
	}

	if err := validPost.Validate(); err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Valid post")
	}

	// Invalid post (missing title)
	invalidPost := &models.Post{
		Date:        time.Now(),
		Description: "A description",
		SourcePath:  "posts/invalid.md",
	}

	if err := invalidPost.Validate(); err != nil {
		fmt.Println("Invalid: missing title")
	}

	// Output: Valid post
	// Invalid: missing title
}

func ExamplePost_GenerateSlug() {
	post := &models.Post{
		Title: "Hello World! This is My First Post",
	}

	post.GenerateSlug()
	fmt.Println(post.Slug)

	// Slug generation is idempotent
	post.GenerateSlug()
	fmt.Println(post.Slug)

	// Output: hello-world-this-is-my-first-post
	// hello-world-this-is-my-first-post
}

func ExamplePost_HasTag() {
	post := &models.Post{
		Title: "Go Tutorial",
		Tags:  []string{"Go", "Programming", "Tutorial"},
	}

	// Tag comparison is case-insensitive
	fmt.Println(post.HasTag("go"))
	fmt.Println(post.HasTag("GO"))
	fmt.Println(post.HasTag("programming"))
	fmt.Println(post.HasTag("javascript"))

	// Output: true
	// true
	// true
	// false
}

func ExamplePost_FormattedDate() {
	post := &models.Post{
		Date: time.Date(2024, 12, 25, 10, 30, 0, 0, time.UTC),
	}

	fmt.Println(post.FormattedDate())
	// Output: December 25, 2024
}

func ExamplePost_ShortDate() {
	post := &models.Post{
		Date: time.Date(2024, 3, 5, 10, 30, 0, 0, time.UTC),
	}

	fmt.Println(post.ShortDate())
	// Output: 2024-03-05
}

func ExamplePostList_FilterByTag() {
	posts := models.PostList{
		{Title: "Go Basics", Tags: []string{"go", "tutorial"}},
		{Title: "Python Guide", Tags: []string{"python", "tutorial"}},
		{Title: "Advanced Go", Tags: []string{"go", "advanced"}},
	}

	// Filter posts by tag
	goPosts := posts.FilterByTag("go")
	fmt.Printf("Found %d posts about Go\n", len(goPosts))

	for _, post := range goPosts {
		fmt.Println(post.Title)
	}

	// Output: Found 2 posts about Go
	// Go Basics
	// Advanced Go
}

func ExamplePostList_SortByDate() {
	posts := models.PostList{
		{Title: "Oldest Post", Date: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Title: "Newest Post", Date: time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)},
		{Title: "Middle Post", Date: time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC)},
	}

	// Sort by date (newest first)
	posts.SortByDate()

	for _, post := range posts {
		fmt.Printf("%s: %s\n", post.Title, post.ShortDate())
	}

	// Output: Newest Post: 2024-03-15
	// Middle Post: 2024-02-10
	// Oldest Post: 2024-01-01
}

func ExamplePostList_GetAllTags() {
	posts := models.PostList{
		{Tags: []string{"go", "tutorial"}},
		{Tags: []string{"python", "tutorial"}},
		{Tags: []string{"go", "advanced"}},
	}

	allTags := posts.GetAllTags()

	// Note: order is non-deterministic, so we just check the count
	fmt.Printf("Found %d unique tags\n", len(allTags))

	// Output: Found 4 unique tags
}

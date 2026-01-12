// Package models provides data structures for representing blog posts and collections.
//
// The primary types in this package are Post and PostList, which are designed to work
// with the parser package to create a complete blog generation system.
//
// # Basic Usage
//
// Post objects are typically created by the parser package, but can also be constructed
// manually:
//
//	post := &models.Post{
//	    Title:       "My First Blog Post",
//	    Date:        time.Now(),
//	    Description: "An introduction to my blog",
//	    Tags:        []string{"intro", "meta"},
//	    Content:     "<p>Hello, world!</p>",
//	}
//
//	if err := post.Validate(); err != nil {
//	    log.Fatal(err)
//	}
//
//	post.GenerateSlug()
//	fmt.Println(post.Slug) // Output: my-first-blog-post
//
// # Working with PostList
//
// PostList provides methods for filtering, sorting, and working with collections of posts:
//
//	var posts models.PostList = []*models.Post{post1, post2, post3}
//
//	// Sort by date (newest first)
//	posts.SortByDate()
//
//	// Filter by tag
//	goPosts := posts.FilterByTag("go")
//
//	// Get all unique tags
//	allTags := posts.GetAllTags()
//
// # Concurrency Safety
//
// Post and PostList types are not safe for concurrent modification. If you need to
// access posts from multiple goroutines, you must provide your own synchronization.
// However, read-only operations on Post objects (such as calling HasTag or
// FormattedDate) are safe for concurrent use.
package models

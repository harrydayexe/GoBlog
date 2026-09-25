// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

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
// # Template Data
//
// Each page type has a data struct (IndexPageData, PostPageData, TagPageData,
// TagsIndexPageData, SeriesPageData, SeriesIndexPageData) embedding BaseData,
// the fields available to every template. Alongside the site title, blog root, and page path, BaseData
// carries the metadata the default templates use for SEO:
//
//   - CanonicalURL — the page's fully-qualified URL, empty unless a base URL
//     is configured via config.WithBaseURL
//   - OGType — the Open Graph object type ("article" or "website")
//   - Article — an [ArticleMeta] with the post's publish date, author, and
//     tags; nil on every page that is not a post
//
// Each is documented with the template guard it expects. Templates must honour
// those guards: the fields are deliberately empty or nil when the underlying
// value is unknown, so that pages omit a tag rather than emit an empty one.
//
// PostPageData.Series ([PostSeries]) follows the same rule. It carries the
// series a post belongs to, its position within it and the parts either side,
// and is nil both for a post in no series and whenever series are disabled.
//
// # Concurrency Safety
//
// Post and PostList types are not safe for concurrent modification. If you need to
// access posts from multiple goroutines, you must provide your own synchronization.
// However, read-only operations on Post objects (such as calling HasTag or
// FormattedDate) are safe for concurrent use.
package models

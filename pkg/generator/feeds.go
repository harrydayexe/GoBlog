// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package generator

import (
	"fmt"
	"strings"

	"github.com/gorilla/feeds"
	"github.com/harrydayexe/GoBlog/v2/pkg/models"
)

// buildFeeds generates RSS 2.0 and Atom XML bytes for the given list of posts
// using the generator's configured site title, base URL, blog root, and post
// limit.
//
// posts must already be sorted (newest first); they are sliced to the
// generator's FeedPostLimit before the feed is assembled.
//
// Both the RSS 2.0 and Atom representations are returned. If the list is
// empty both slices are nil.
//
// The post URL is derived from the generator's BaseURL and BlogRoot as:
//
//	<BaseURL><BlogRoot>posts/<slug>
//
// and is used as both the item link and its unique identifier (guid/id).
// Each item carries the post's full rendered HTML content.
func (g *Generator) buildFeeds(posts models.PostList) (rss []byte, atom []byte, err error) {
	return g.buildFeedsWithTitle(posts, g.SiteTitle.SiteTitle)
}

// buildFeedsWithTitle is the internal implementation shared by BuildFeeds and
// per-tag feed generation, where the feed title differs from the site title.
func (g *Generator) buildFeedsWithTitle(posts models.PostList, feedTitle string) (rss []byte, atom []byte, err error) {
	if len(posts) == 0 {
		return nil, nil, nil
	}

	limit := g.FeedPostLimit.Limit
	if limit > 0 && len(posts) > limit {
		posts = posts[:limit]
	}

	siteURL := AbsURL(string(g.BaseURL), string(g.BlogRoot))
	absPostURL := func(slug string) string {
		return AbsURL(string(g.BaseURL), string(g.BlogRoot)+"posts/"+slug)
	}

	// The feed's updated time is the date of the newest post.
	updated := posts[0].Date

	feed := &feeds.Feed{
		Title: feedTitle,
		Link:  &feeds.Link{Href: siteURL},
		// Description is required by RSS 2.0.
		Description: feedTitle + " — RSS Feed",
		Created:     updated,
		Updated:     updated,
	}

	feed.Items = make([]*feeds.Item, 0, len(posts))
	for _, post := range posts {
		postURL := absPostURL(post.Slug)

		// Use LastEdited as the updated time when set; otherwise fall back to Date.
		itemUpdated := post.Date
		if !post.LastEdited.IsZero() {
			itemUpdated = post.LastEdited
		}

		item := &feeds.Item{
			// The post URL serves as the unique, permanent identifier.
			Id:          postURL,
			Title:       post.Title,
			Link:        &feeds.Link{Href: postURL},
			Description: post.Description,
			// Full HTML content (not a summary).
			Content: string(post.Content),
			Created: post.Date,
			Updated: itemUpdated,
		}

		if post.Author != "" {
			item.Author = &feeds.Author{Name: post.Author}
		}

		feed.Items = append(feed.Items, item)
	}

	rssStr, err := feed.ToRss()
	if err != nil {
		return nil, nil, fmt.Errorf("generating RSS feed: %w", err)
	}

	atomStr, err := feed.ToAtom()
	if err != nil {
		return nil, nil, fmt.Errorf("generating Atom feed: %w", err)
	}

	return []byte(rssStr), []byte(atomStr), nil
}

// AbsURL returns the absolute URL for a site-relative path by prepending the
// base URL (trailing slash trimmed).
func AbsURL(baseURL, path string) string {
	return strings.TrimRight(baseURL, "/") + path
}

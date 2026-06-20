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

// buildFeeds generates RSS 2.0 and Atom XML bytes for the given list of posts.
//
// posts must already be sorted (newest first) and will be sliced to limit
// entries before building the feed. absPostURL is a function that accepts a
// post slug and returns the fully-qualified URL for that post's page.
//
// The feed's own link (the href used in channel/feed metadata) is built from
// siteURL, which is the value of config.BaseURL with the trailing slash trimmed.
//
// Both the RSS and Atom representations are returned; if the list is empty
// both slices are nil.
func buildFeeds(
	posts models.PostList,
	limit int,
	siteTitle string,
	siteURL string,
	absPostURL func(slug string) string,
) (rss []byte, atom []byte, err error) {
	if len(posts) == 0 {
		return nil, nil, nil
	}

	// Trim to the configured limit.
	if limit > 0 && len(posts) > limit {
		posts = posts[:limit]
	}

	// The feed's updated time is the date of the newest post.
	updated := posts[0].Date

	feed := &feeds.Feed{
		Title: siteTitle,
		Link:  &feeds.Link{Href: siteURL},
		// Description is required by RSS 2.0.
		Description: siteTitle + " — RSS Feed",
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

// absURL returns the absolute URL for a site-relative path by prepending the
// base URL (trailing slash trimmed).
func absURL(baseURL, path string) string {
	return strings.TrimRight(baseURL, "/") + path
}

// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package generator

import (
	"fmt"
	"strings"
	"time"

	"github.com/gorilla/feeds"
	"github.com/harrydayexe/GoBlog/v2/pkg/models"
)

// buildFeeds generates the site-wide RSS 2.0 and Atom feeds from the given
// post list, using the generator's site title, base URL, blog root, and post
// limit.
//
// posts must already be sorted newest-first. Both representations are
// returned; if posts is empty both slices are nil.
func (g *Generator) buildFeeds(posts models.PostList) (rss []byte, atom []byte, err error) {
	return g.buildFeedsForTitle(posts, g.SiteTitle.SiteTitle)
}

// buildTagFeeds generates per-tag RSS 2.0 and Atom feeds for the given post
// list, titling the feed "<SiteTitle> — <tag>".
//
// posts must already be sorted newest-first and pre-filtered to the tag.
// Both representations are returned; if posts is empty both slices are nil.
func (g *Generator) buildTagFeeds(tag string, posts models.PostList) (rss []byte, atom []byte, err error) {
	return g.buildFeedsForTitle(posts, g.SiteTitle.SiteTitle+" — "+tag)
}

// buildFeedsForTitle is the shared implementation for buildFeeds and
// buildTagFeeds. title is the channel/feed title to embed in the output.
func (g *Generator) buildFeedsForTitle(posts models.PostList, title string) (rss []byte, atom []byte, err error) {
	if len(posts) == 0 {
		return nil, nil, nil
	}

	limit := g.FeedPostLimit.Limit
	if limit > 0 && len(posts) > limit {
		posts = posts[:limit]
	}

	siteURL := absURL(string(g.BaseURL), string(g.BlogRoot))
	absPostURL := func(slug string) string {
		return absURL(string(g.BaseURL), string(g.BlogRoot)+"posts/"+slug)
	}

	// The feed's updated time is the most recent effective-updated time across
	// all included posts, per RFC 4287: a feed's <updated> must reflect the
	// most recent instant any entry was significantly modified. Using only
	// posts[0].Date would miss edits to older posts.
	updated := posts[0].Date
	for _, p := range posts {
		if eff := effectiveUpdated(p); eff.After(updated) {
			updated = eff
		}
	}

	feed := &feeds.Feed{
		Title:       title,
		Link:        &feeds.Link{Href: siteURL},
		Description: title + " — RSS Feed",
		Created:     posts[0].Date,
		Updated:     updated,
	}

	feed.Items = make([]*feeds.Item, 0, len(posts))
	for _, post := range posts {
		postURL := absPostURL(post.Slug)

		item := &feeds.Item{
			// The post URL serves as the unique, permanent identifier.
			Id:          postURL,
			Title:       post.Title,
			Link:        &feeds.Link{Href: postURL},
			Description: post.Description,
			Content:     string(post.Content),
			Created:     post.Date,
			Updated:     effectiveUpdated(post),
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

// effectiveUpdated returns the post's last-modified instant: LastEdited when
// set, otherwise the publication Date.
func effectiveUpdated(p *models.Post) time.Time {
	if !p.LastEdited.IsZero() {
		return p.LastEdited
	}
	return p.Date
}

// absURL returns the absolute URL for a site-relative path by prepending the
// base URL (trailing slash trimmed).
func absURL(baseURL, path string) string {
	return strings.TrimRight(baseURL, "/") + path
}

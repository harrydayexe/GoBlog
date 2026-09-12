// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package models

import "time"

// ArticleMeta carries the article-specific metadata needed to describe a post
// to social platforms and search engines: the article:* Open Graph tags and
// the Schema.org BlogPosting fields.
//
// It exists because the <head> is rendered by a partial shared with every page
// type, and Go templates cannot inject markup into a block they invoke. Post
// pages instead hand the partial this struct via BaseData.Article; on all
// other pages the field is nil and the article-specific markup is skipped.
//
// The fields mirror the underlying Post, so a post page template can reach the
// same values through .Post. Prefer .Article in the shared partial, which has
// no Post to read.
type ArticleMeta struct {
	// PublishedTime is the post's publication date, emitted as
	// article:published_time and Schema.org datePublished.
	PublishedTime time.Time

	// Author is the name of the post's author, emitted as article:author.
	// It is empty when the post's front matter declares no author; templates
	// should guard on it so no empty tag is emitted:
	//
	//   {{with .Author}}<meta property="article:author" content="{{.}}">{{end}}
	Author string

	// Tags are the post's tags, emitted as one article:tag per entry. It is
	// empty when the post has no tags, and also when tags are disabled for the
	// blog via config.WithDisableTags.
	Tags []string
}

// PublishedISO returns PublishedTime as an ISO 8601 (RFC 3339) timestamp,
// the format required by article:published_time and Schema.org datePublished.
//
// Example: "2024-06-01T09:30:00Z"
func (a *ArticleMeta) PublishedISO() string {
	return a.PublishedTime.Format(time.RFC3339)
}

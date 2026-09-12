// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package models

import (
	"fmt"
	"time"
)

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

	// ModifiedTime is the date the post was last edited after publication,
	// emitted as article:modified_time and Schema.org dateModified.
	//
	// It is the zero time when the post's front matter declares no lastEdited
	// date. Use ModifiedISO, which returns an empty string in that case, so
	// templates can gate on the value:
	//
	//   {{with .ModifiedISO}}<meta property="article:modified_time" content="{{.}}">{{end}}
	ModifiedTime time.Time

	// Tags are the post's tags, emitted as one article:tag per entry. It is
	// empty when the post has no tags, and also when tags are disabled for the
	// blog via config.WithDisableTags.
	Tags []string

	// ReadingTimeMinutes is the post's estimated reading time, emitted as
	// Schema.org timeRequired.
	//
	// It is zero when reading time estimation is disabled via
	// config.WithDisableReadingTime. Use TimeRequired, which returns an empty
	// string in that case and formats the value as the ISO 8601 duration that
	// Schema.org expects.
	ReadingTimeMinutes int
}

// PublishedISO returns PublishedTime as an ISO 8601 (RFC 3339) timestamp,
// the format required by article:published_time and Schema.org datePublished.
//
// Example: "2024-06-01T09:30:00Z"
func (a *ArticleMeta) PublishedISO() string {
	return a.PublishedTime.Format(time.RFC3339)
}

// ModifiedISO returns ModifiedTime as an ISO 8601 (RFC 3339) timestamp, the
// format required by article:modified_time and Schema.org dateModified.
//
// It returns an empty string when the post has never been edited, so that
// templates gate the tag on the value itself:
//
//	{{with .ModifiedISO}}<meta property="article:modified_time" content="{{.}}">{{end}}
func (a *ArticleMeta) ModifiedISO() string {
	if a.ModifiedTime.IsZero() {
		return ""
	}
	return a.ModifiedTime.Format(time.RFC3339)
}

// TimeRequired returns ReadingTimeMinutes as an ISO 8601 duration, the format
// required by Schema.org timeRequired.
//
// It returns an empty string when no reading time is available, so that
// templates gate the value itself:
//
//	{{with .TimeRequired}}"timeRequired": "{{.}}",{{end}}
//
// Example: "PT5M" for a five minute read.
func (a *ArticleMeta) TimeRequired() string {
	if a.ReadingTimeMinutes <= 0 {
		return ""
	}
	return fmt.Sprintf("PT%dM", a.ReadingTimeMinutes)
}

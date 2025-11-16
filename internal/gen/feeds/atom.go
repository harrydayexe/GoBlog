package feeds

import (
	"encoding/xml"
	"fmt"
	"time"

	"github.com/harrydayexe/GoBlog/pkg/models"
)

// AtomFeed represents the root Atom 1.0 feed
type AtomFeed struct {
	XMLName xml.Name    `xml:"feed"`
	XMLNS   string      `xml:"xmlns,attr"`
	Title   string      `xml:"title"`
	Link    []AtomLink  `xml:"link"`
	ID      string      `xml:"id"`
	Updated string      `xml:"updated"`
	Author  *AtomAuthor `xml:"author,omitempty"`
	Entries []AtomEntry `xml:"entry"`
}

// AtomLink represents an Atom link element
type AtomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr,omitempty"`
	Type string `xml:"type,attr,omitempty"`
}

// AtomAuthor represents an Atom author element
type AtomAuthor struct {
	Name  string `xml:"name"`
	Email string `xml:"email,omitempty"`
}

// AtomEntry represents an Atom entry (post)
type AtomEntry struct {
	Title   string      `xml:"title"`
	Link    []AtomLink  `xml:"link"`
	ID      string      `xml:"id"`
	Updated string      `xml:"updated"`
	Summary string      `xml:"summary,omitempty"`
	Content *AtomContent `xml:"content,omitempty"`
	Author  *AtomAuthor `xml:"author,omitempty"`
}

// AtomContent represents Atom content element
type AtomContent struct {
	Type string `xml:"type,attr"`
	Text string `xml:",chardata"`
}

// AtomConfig contains configuration for Atom feed generation
type AtomConfig struct {
	Title       string
	Link        string
	Description string
	Author      string
	AuthorEmail string
	BlogPath    string
}

// GenerateAtom generates an Atom 1.0 feed from posts
func GenerateAtom(posts []*models.Post, config AtomConfig) (string, error) {
	feed := &AtomFeed{
		XMLNS: "http://www.w3.org/2005/Atom",
		Title: config.Title,
		Link: []AtomLink{
			{
				Href: config.Link + config.BlogPath,
				Rel:  "self",
				Type: "application/atom+xml",
			},
			{
				Href: config.Link,
				Rel:  "alternate",
				Type: "text/html",
			},
		},
		ID:      config.Link + config.BlogPath,
		Entries: make([]AtomEntry, 0, len(posts)),
	}

	// Set updated time to now or most recent post
	if len(posts) > 0 {
		feed.Updated = posts[0].Date.Format(time.RFC3339)
	} else {
		feed.Updated = time.Now().Format(time.RFC3339)
	}

	// Set author if provided
	if config.Author != "" {
		feed.Author = &AtomAuthor{
			Name:  config.Author,
			Email: config.AuthorEmail,
		}
	}

	// Add entries (posts)
	for _, post := range posts {
		postURL := fmt.Sprintf("%s%s/posts/%s", config.Link, config.BlogPath, post.Slug)

		entry := AtomEntry{
			Title: post.Title,
			Link: []AtomLink{
				{
					Href: postURL,
					Rel:  "alternate",
					Type: "text/html",
				},
			},
			ID:      postURL,
			Updated: post.Date.Format(time.RFC3339),
			Summary: post.Description,
		}

		// Add content if available
		if post.HTMLContent != "" {
			entry.Content = &AtomContent{
				Type: "html",
				Text: string(post.HTMLContent),
			}
		}

		// Add author
		if config.Author != "" {
			entry.Author = &AtomAuthor{
				Name:  config.Author,
				Email: config.AuthorEmail,
			}
		}

		feed.Entries = append(feed.Entries, entry)
	}

	// Marshal to XML
	output, err := xml.MarshalIndent(feed, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal Atom feed: %w", err)
	}

	return xml.Header + string(output), nil
}

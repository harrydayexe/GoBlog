package feeds

import (
	"encoding/xml"
	"fmt"
	"time"

	"github.com/harrydayexe/GoBlog/pkg/models"
)

// RSS2Feed represents the root RSS 2.0 feed
type RSS2Feed struct {
	XMLName xml.Name    `xml:"rss"`
	Version string      `xml:"version,attr"`
	Channel *RSS2Channel `xml:"channel"`
}

// RSS2Channel represents the RSS channel
type RSS2Channel struct {
	Title         string      `xml:"title"`
	Link          string      `xml:"link"`
	Description   string      `xml:"description"`
	Language      string      `xml:"language,omitempty"`
	Copyright     string      `xml:"copyright,omitempty"`
	LastBuildDate string      `xml:"lastBuildDate,omitempty"`
	PubDate       string      `xml:"pubDate,omitempty"`
	Generator     string      `xml:"generator,omitempty"`
	Items         []RSS2Item  `xml:"item"`
}

// RSS2Item represents an RSS item (post)
type RSS2Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Author      string `xml:"author,omitempty"`
	PubDate     string `xml:"pubDate"`
	GUID        string `xml:"guid"`
}

// RSSConfig contains configuration for RSS feed generation
type RSSConfig struct {
	Title       string
	Link        string
	Description string
	Language    string
	Copyright   string
	Author      string
	BlogPath    string
}

// GenerateRSS generates an RSS 2.0 feed from posts
func GenerateRSS(posts []*models.Post, config RSSConfig) (string, error) {
	channel := &RSS2Channel{
		Title:       config.Title,
		Link:        config.Link,
		Description: config.Description,
		Language:    config.Language,
		Copyright:   config.Copyright,
		Generator:   "GoBlog",
		Items:       make([]RSS2Item, 0, len(posts)),
	}

	// Set last build date to now
	channel.LastBuildDate = time.Now().Format(time.RFC1123Z)

	// Set pub date to most recent post date
	if len(posts) > 0 {
		channel.PubDate = posts[0].Date.Format(time.RFC1123Z)
	}

	// Add items (posts)
	for _, post := range posts {
		postURL := fmt.Sprintf("%s%s/posts/%s", config.Link, config.BlogPath, post.Slug)

		item := RSS2Item{
			Title:       post.Title,
			Link:        postURL,
			Description: post.Description,
			Author:      config.Author,
			PubDate:     post.Date.Format(time.RFC1123Z),
			GUID:        postURL,
		}

		channel.Items = append(channel.Items, item)
	}

	feed := &RSS2Feed{
		Version: "2.0",
		Channel: channel,
	}

	// Marshal to XML
	output, err := xml.MarshalIndent(feed, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal RSS feed: %w", err)
	}

	return xml.Header + string(output), nil
}

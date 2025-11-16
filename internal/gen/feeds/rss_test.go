package feeds

import (
	"encoding/xml"
	"strings"
	"testing"
	"time"

	"github.com/harrydayexe/GoBlog/pkg/models"
)

func TestGenerateRSS(t *testing.T) {
	t.Parallel()
	t.Run("valid rss feed with posts", func(t *testing.T) {
		t.Parallel()
		posts := []*models.Post{
			{
				Title:       "Test Post 1",
				Slug:        "test-post-1",
				Date:        time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
				Description: "First test post",
			},
			{
				Title:       "Test Post 2",
				Slug:        "test-post-2",
				Date:        time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC),
				Description: "Second test post",
			},
		}

		config := RSSConfig{
			Title:       "Test Blog",
			Link:        "https://example.com",
			Description: "A test blog",
			Language:    "en-us",
			Copyright:   "Copyright 2024",
			Author:      "test@example.com",
			BlogPath:    "/blog",
		}

		result, err := GenerateRSS(posts, config)
		if err != nil {
			t.Fatalf("GenerateRSS failed: %v", err)
		}

		// Check XML header
		if !strings.HasPrefix(result, xml.Header) {
			t.Error("Expected XML header")
		}

		// Parse to validate XML structure
		var feed RSS2Feed
		if err := xml.Unmarshal([]byte(result), &feed); err != nil {
			t.Fatalf("Failed to parse generated XML: %v", err)
		}

		// Verify feed version
		if feed.Version != "2.0" {
			t.Errorf("Expected RSS version %q, got %q", "2.0", feed.Version)
		}

		// Verify channel metadata
		if feed.Channel == nil {
			t.Fatal("Expected channel to be set")
		}

		channel := feed.Channel
		if channel.Title != config.Title {
			t.Errorf("Expected title %q, got %q", config.Title, channel.Title)
		}
		if channel.Link != config.Link {
			t.Errorf("Expected link %q, got %q", config.Link, channel.Link)
		}
		if channel.Description != config.Description {
			t.Errorf("Expected description %q, got %q", config.Description, channel.Description)
		}
		if channel.Language != config.Language {
			t.Errorf("Expected language %q, got %q", config.Language, channel.Language)
		}
		if channel.Copyright != config.Copyright {
			t.Errorf("Expected copyright %q, got %q", config.Copyright, channel.Copyright)
		}
		if channel.Generator != "GoBlog" {
			t.Errorf("Expected generator %q, got %q", "GoBlog", channel.Generator)
		}

		// Verify LastBuildDate is set
		if channel.LastBuildDate == "" {
			t.Error("Expected LastBuildDate to be set")
		}

		// Verify PubDate is set to most recent post
		expectedPubDate := posts[0].Date.Format(time.RFC1123Z)
		if channel.PubDate != expectedPubDate {
			t.Errorf("Expected PubDate %q, got %q", expectedPubDate, channel.PubDate)
		}

		// Verify items
		if len(channel.Items) != len(posts) {
			t.Fatalf("Expected %d items, got %d", len(posts), len(channel.Items))
		}

		// Check first item
		item := channel.Items[0]
		if item.Title != posts[0].Title {
			t.Errorf("Expected item title %q, got %q", posts[0].Title, item.Title)
		}
		if item.Description != posts[0].Description {
			t.Errorf("Expected item description %q, got %q", posts[0].Description, item.Description)
		}
		if item.Author != config.Author {
			t.Errorf("Expected item author %q, got %q", config.Author, item.Author)
		}

		expectedURL := "https://example.com/blog/posts/test-post-1"
		if item.Link != expectedURL {
			t.Errorf("Expected item link %q, got %q", expectedURL, item.Link)
		}
		if item.GUID != expectedURL {
			t.Errorf("Expected item GUID %q, got %q", expectedURL, item.GUID)
		}

		expectedItemPubDate := posts[0].Date.Format(time.RFC1123Z)
		if item.PubDate != expectedItemPubDate {
			t.Errorf("Expected item PubDate %q, got %q", expectedItemPubDate, item.PubDate)
		}
	})

	t.Run("rss feed with empty posts", func(t *testing.T) {
		t.Parallel()
		posts := []*models.Post{}

		config := RSSConfig{
			Title:       "Empty Blog",
			Link:        "https://example.com",
			Description: "An empty blog",
			BlogPath:    "/blog",
		}

		result, err := GenerateRSS(posts, config)
		if err != nil {
			t.Fatalf("GenerateRSS failed: %v", err)
		}

		var feed RSS2Feed
		if err := xml.Unmarshal([]byte(result), &feed); err != nil {
			t.Fatalf("Failed to parse generated XML: %v", err)
		}

		if len(feed.Channel.Items) != 0 {
			t.Errorf("Expected 0 items, got %d", len(feed.Channel.Items))
		}

		// LastBuildDate should still be set
		if feed.Channel.LastBuildDate == "" {
			t.Error("Expected LastBuildDate to be set")
		}

		// PubDate should be empty when no posts
		if feed.Channel.PubDate != "" {
			t.Errorf("Expected PubDate to be empty, got %q", feed.Channel.PubDate)
		}
	})

	t.Run("rss feed with minimal config", func(t *testing.T) {
		t.Parallel()
		posts := []*models.Post{
			{
				Title:       "Test Post",
				Slug:        "test-post",
				Date:        time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
				Description: "Test post",
			},
		}

		config := RSSConfig{
			Title:       "Minimal Blog",
			Link:        "https://example.com",
			Description: "A minimal blog",
			BlogPath:    "/",
			// No Language, Copyright, or Author
		}

		result, err := GenerateRSS(posts, config)
		if err != nil {
			t.Fatalf("GenerateRSS failed: %v", err)
		}

		var feed RSS2Feed
		if err := xml.Unmarshal([]byte(result), &feed); err != nil {
			t.Fatalf("Failed to parse generated XML: %v", err)
		}

		// Optional fields should be empty
		if feed.Channel.Language != "" {
			t.Errorf("Expected language to be empty, got %q", feed.Channel.Language)
		}
		if feed.Channel.Copyright != "" {
			t.Errorf("Expected copyright to be empty, got %q", feed.Channel.Copyright)
		}
		if feed.Channel.Items[0].Author != "" {
			t.Errorf("Expected item author to be empty, got %q", feed.Channel.Items[0].Author)
		}

		// Generator should still be set
		if feed.Channel.Generator != "GoBlog" {
			t.Errorf("Expected generator to be %q, got %q", "GoBlog", feed.Channel.Generator)
		}
	})

	t.Run("rss feed with custom blog path", func(t *testing.T) {
		t.Parallel()
		posts := []*models.Post{
			{
				Title:       "Test Post",
				Slug:        "test",
				Date:        time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
				Description: "Test",
			},
		}

		config := RSSConfig{
			Title:       "Test Blog",
			Link:        "https://example.com",
			Description: "Test",
			BlogPath:    "/custom/path",
		}

		result, err := GenerateRSS(posts, config)
		if err != nil {
			t.Fatalf("GenerateRSS failed: %v", err)
		}

		expectedURL := "https://example.com/custom/path/posts/test"
		if !strings.Contains(result, expectedURL) {
			t.Errorf("Expected result to contain URL %q", expectedURL)
		}
	})

	t.Run("rss feed with multiple posts sorts correctly", func(t *testing.T) {
		t.Parallel()
		oldDate := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
		newDate := time.Date(2024, 12, 1, 12, 0, 0, 0, time.UTC)

		posts := []*models.Post{
			{
				Title:       "New Post",
				Slug:        "new",
				Date:        newDate,
				Description: "Newer post",
			},
			{
				Title:       "Old Post",
				Slug:        "old",
				Date:        oldDate,
				Description: "Older post",
			},
		}

		config := RSSConfig{
			Title:       "Test Blog",
			Link:        "https://example.com",
			Description: "Test",
			BlogPath:    "/blog",
		}

		result, err := GenerateRSS(posts, config)
		if err != nil {
			t.Fatalf("GenerateRSS failed: %v", err)
		}

		var feed RSS2Feed
		if err := xml.Unmarshal([]byte(result), &feed); err != nil {
			t.Fatalf("Failed to parse generated XML: %v", err)
		}

		// PubDate should be the most recent post's date
		expectedPubDate := newDate.Format(time.RFC1123Z)
		if feed.Channel.PubDate != expectedPubDate {
			t.Errorf("Expected PubDate %q, got %q", expectedPubDate, feed.Channel.PubDate)
		}

		// First item should be the newer post
		if feed.Channel.Items[0].Title != "New Post" {
			t.Errorf("Expected first item to be %q, got %q", "New Post", feed.Channel.Items[0].Title)
		}
	})

	t.Run("rss feed with special characters in content", func(t *testing.T) {
		t.Parallel()
		posts := []*models.Post{
			{
				Title:       "Post with <HTML> & Special \"Chars\"",
				Slug:        "special",
				Date:        time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
				Description: "Description with <tags> & entities",
			},
		}

		config := RSSConfig{
			Title:       "Test Blog",
			Link:        "https://example.com",
			Description: "Test",
			BlogPath:    "/blog",
		}

		result, err := GenerateRSS(posts, config)
		if err != nil {
			t.Fatalf("GenerateRSS failed: %v", err)
		}

		// XML should be properly escaped
		var feed RSS2Feed
		if err := xml.Unmarshal([]byte(result), &feed); err != nil {
			t.Fatalf("Failed to parse generated XML with special chars: %v", err)
		}

		// After unmarshaling, special chars should be decoded
		if feed.Channel.Items[0].Title != posts[0].Title {
			t.Errorf("Expected title %q, got %q", posts[0].Title, feed.Channel.Items[0].Title)
		}
	})

	t.Run("rss feed includes all required fields", func(t *testing.T) {
		t.Parallel()
		posts := []*models.Post{
			{
				Title:       "Test Post",
				Slug:        "test",
				Date:        time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
				Description: "Test",
			},
		}

		config := RSSConfig{
			Title:       "Test Blog",
			Link:        "https://example.com",
			Description: "Test blog description",
			Language:    "en",
			Copyright:   "© 2024",
			Author:      "author@example.com",
			BlogPath:    "/blog",
		}

		result, err := GenerateRSS(posts, config)
		if err != nil {
			t.Fatalf("GenerateRSS failed: %v", err)
		}

		// Verify all fields are present in the output
		requiredFields := []string{
			"<title>",
			"<link>",
			"<description>",
			"<language>",
			"<copyright>",
			"<lastBuildDate>",
			"<pubDate>",
			"<generator>",
			"<item>",
			"<guid>",
		}

		for _, field := range requiredFields {
			if !strings.Contains(result, field) {
				t.Errorf("Expected result to contain %q", field)
			}
		}
	})
}

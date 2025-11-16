package feeds

import (
	"encoding/xml"
	"html/template"
	"strings"
	"testing"
	"time"

	"github.com/harrydayexe/GoBlog/pkg/models"
)

func TestGenerateAtom(t *testing.T) {
	t.Run("valid atom feed with posts", func(t *testing.T) {
		posts := []*models.Post{
			{
				Title:       "Test Post 1",
				Slug:        "test-post-1",
				Date:        time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
				Description: "First test post",
				HTMLContent: template.HTML("<p>Content of first post</p>"),
			},
			{
				Title:       "Test Post 2",
				Slug:        "test-post-2",
				Date:        time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC),
				Description: "Second test post",
				HTMLContent: template.HTML("<p>Content of second post</p>"),
			},
		}

		config := AtomConfig{
			Title:       "Test Blog",
			Link:        "https://example.com",
			Description: "A test blog",
			Author:      "Test Author",
			AuthorEmail: "test@example.com",
			BlogPath:    "/blog",
		}

		result, err := GenerateAtom(posts, config)
		if err != nil {
			t.Fatalf("GenerateAtom failed: %v", err)
		}

		// Check XML header
		if !strings.HasPrefix(result, xml.Header) {
			t.Error("Expected XML header")
		}

		// Parse to validate XML structure
		var feed AtomFeed
		if err := xml.Unmarshal([]byte(result), &feed); err != nil {
			t.Fatalf("Failed to parse generated XML: %v", err)
		}

		// Verify feed metadata
		if feed.Title != config.Title {
			t.Errorf("Expected title %q, got %q", config.Title, feed.Title)
		}

		if feed.XMLNS != "http://www.w3.org/2005/Atom" {
			t.Errorf("Expected xmlns %q, got %q", "http://www.w3.org/2005/Atom", feed.XMLNS)
		}

		if feed.Author == nil {
			t.Fatal("Expected author to be set")
		}
		if feed.Author.Name != config.Author {
			t.Errorf("Expected author name %q, got %q", config.Author, feed.Author.Name)
		}
		if feed.Author.Email != config.AuthorEmail {
			t.Errorf("Expected author email %q, got %q", config.AuthorEmail, feed.Author.Email)
		}

		// Verify links
		if len(feed.Link) != 2 {
			t.Fatalf("Expected 2 links, got %d", len(feed.Link))
		}

		selfLink := feed.Link[0]
		if selfLink.Rel != "self" {
			t.Errorf("Expected first link rel %q, got %q", "self", selfLink.Rel)
		}
		if selfLink.Type != "application/atom+xml" {
			t.Errorf("Expected first link type %q, got %q", "application/atom+xml", selfLink.Type)
		}

		// Verify entries
		if len(feed.Entries) != len(posts) {
			t.Fatalf("Expected %d entries, got %d", len(posts), len(feed.Entries))
		}

		// Check first entry
		entry := feed.Entries[0]
		if entry.Title != posts[0].Title {
			t.Errorf("Expected entry title %q, got %q", posts[0].Title, entry.Title)
		}
		if entry.Summary != posts[0].Description {
			t.Errorf("Expected entry summary %q, got %q", posts[0].Description, entry.Summary)
		}
		if entry.Content == nil {
			t.Fatal("Expected entry content to be set")
		}
		if entry.Content.Type != "html" {
			t.Errorf("Expected content type %q, got %q", "html", entry.Content.Type)
		}

		expectedURL := "https://example.com/blog/posts/test-post-1"
		if entry.ID != expectedURL {
			t.Errorf("Expected entry ID %q, got %q", expectedURL, entry.ID)
		}
	})

	t.Run("atom feed with empty posts", func(t *testing.T) {
		posts := []*models.Post{}

		config := AtomConfig{
			Title:    "Empty Blog",
			Link:     "https://example.com",
			BlogPath: "/blog",
		}

		result, err := GenerateAtom(posts, config)
		if err != nil {
			t.Fatalf("GenerateAtom failed: %v", err)
		}

		var feed AtomFeed
		if err := xml.Unmarshal([]byte(result), &feed); err != nil {
			t.Fatalf("Failed to parse generated XML: %v", err)
		}

		if len(feed.Entries) != 0 {
			t.Errorf("Expected 0 entries, got %d", len(feed.Entries))
		}

		// Updated should still be set (to current time)
		if feed.Updated == "" {
			t.Error("Expected updated to be set")
		}
	})

	t.Run("atom feed without author", func(t *testing.T) {
		posts := []*models.Post{
			{
				Title:       "Test Post",
				Slug:        "test-post",
				Date:        time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
				Description: "Test post",
				HTMLContent: template.HTML("<p>Content</p>"),
			},
		}

		config := AtomConfig{
			Title:    "Blog Without Author",
			Link:     "https://example.com",
			BlogPath: "/blog",
			// No Author or AuthorEmail
		}

		result, err := GenerateAtom(posts, config)
		if err != nil {
			t.Fatalf("GenerateAtom failed: %v", err)
		}

		var feed AtomFeed
		if err := xml.Unmarshal([]byte(result), &feed); err != nil {
			t.Fatalf("Failed to parse generated XML: %v", err)
		}

		if feed.Author != nil {
			t.Error("Expected author to be nil when not provided")
		}

		// Entry author should also be nil
		if feed.Entries[0].Author != nil {
			t.Error("Expected entry author to be nil when not provided")
		}
	})

	t.Run("atom feed without html content", func(t *testing.T) {
		posts := []*models.Post{
			{
				Title:       "Text Only Post",
				Slug:        "text-only",
				Date:        time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
				Description: "A post without HTML content",
				// No HTMLContent
			},
		}

		config := AtomConfig{
			Title:    "Test Blog",
			Link:     "https://example.com",
			BlogPath: "/blog",
		}

		result, err := GenerateAtom(posts, config)
		if err != nil {
			t.Fatalf("GenerateAtom failed: %v", err)
		}

		var feed AtomFeed
		if err := xml.Unmarshal([]byte(result), &feed); err != nil {
			t.Fatalf("Failed to parse generated XML: %v", err)
		}

		// Content should be nil when HTMLContent is empty
		if feed.Entries[0].Content != nil {
			t.Error("Expected entry content to be nil when HTMLContent is empty")
		}
	})

	t.Run("atom feed uses most recent post date", func(t *testing.T) {
		recentDate := time.Date(2024, 12, 1, 12, 0, 0, 0, time.UTC)
		posts := []*models.Post{
			{
				Title:       "Recent Post",
				Slug:        "recent",
				Date:        recentDate,
				Description: "Most recent post",
			},
			{
				Title:       "Old Post",
				Slug:        "old",
				Date:        time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
				Description: "Older post",
			},
		}

		config := AtomConfig{
			Title:    "Test Blog",
			Link:     "https://example.com",
			BlogPath: "/blog",
		}

		result, err := GenerateAtom(posts, config)
		if err != nil {
			t.Fatalf("GenerateAtom failed: %v", err)
		}

		var feed AtomFeed
		if err := xml.Unmarshal([]byte(result), &feed); err != nil {
			t.Fatalf("Failed to parse generated XML: %v", err)
		}

		expectedUpdated := recentDate.Format(time.RFC3339)
		if feed.Updated != expectedUpdated {
			t.Errorf("Expected updated %q, got %q", expectedUpdated, feed.Updated)
		}
	})

	t.Run("atom feed with custom blog path", func(t *testing.T) {
		posts := []*models.Post{
			{
				Title:       "Test Post",
				Slug:        "test",
				Date:        time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
				Description: "Test",
			},
		}

		config := AtomConfig{
			Title:    "Test Blog",
			Link:     "https://example.com",
			BlogPath: "/custom/path",
		}

		result, err := GenerateAtom(posts, config)
		if err != nil {
			t.Fatalf("GenerateAtom failed: %v", err)
		}

		expectedURL := "https://example.com/custom/path/posts/test"
		if !strings.Contains(result, expectedURL) {
			t.Errorf("Expected result to contain URL %q", expectedURL)
		}
	})

	t.Run("atom feed with author email only", func(t *testing.T) {
		posts := []*models.Post{
			{
				Title:       "Test Post",
				Slug:        "test",
				Date:        time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
				Description: "Test",
			},
		}

		config := AtomConfig{
			Title:       "Test Blog",
			Link:        "https://example.com",
			BlogPath:    "/blog",
			Author:      "John Doe",
			AuthorEmail: "john@example.com",
		}

		result, err := GenerateAtom(posts, config)
		if err != nil {
			t.Fatalf("GenerateAtom failed: %v", err)
		}

		var feed AtomFeed
		if err := xml.Unmarshal([]byte(result), &feed); err != nil {
			t.Fatalf("Failed to parse generated XML: %v", err)
		}

		if feed.Author.Email != config.AuthorEmail {
			t.Errorf("Expected author email %q, got %q", config.AuthorEmail, feed.Author.Email)
		}
	})
}

package sitemap

import (
	"encoding/xml"
	"strings"
	"testing"
	"time"

	"github.com/harrydayexe/GoBlog/pkg/models"
)

func TestGenerateSitemap(t *testing.T) {
	t.Run("valid sitemap with posts", func(t *testing.T) {
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

		config := SitemapConfig{
			SiteURL:  "https://example.com",
			BlogPath: "/blog",
		}

		result, err := GenerateSitemap(posts, config)
		if err != nil {
			t.Fatalf("GenerateSitemap failed: %v", err)
		}

		// Check XML header
		if !strings.HasPrefix(result, xml.Header) {
			t.Error("Expected XML header")
		}

		// Parse to validate XML structure
		var urlset URLSet
		if err := xml.Unmarshal([]byte(result), &urlset); err != nil {
			t.Fatalf("Failed to parse generated XML: %v", err)
		}

		// Verify xmlns
		if urlset.XMLNS != "http://www.sitemaps.org/schemas/sitemap/0.9" {
			t.Errorf("Expected xmlns %q, got %q", "http://www.sitemaps.org/schemas/sitemap/0.9", urlset.XMLNS)
		}

		// Should have: homepage + blog path + 2 posts = 4 URLs
		expectedURLCount := 4
		if len(urlset.URLs) != expectedURLCount {
			t.Fatalf("Expected %d URLs, got %d", expectedURLCount, len(urlset.URLs))
		}

		// Check homepage
		homepage := urlset.URLs[0]
		if homepage.Loc != "https://example.com/" {
			t.Errorf("Expected homepage loc %q, got %q", "https://example.com/", homepage.Loc)
		}
		if homepage.Priority != 1.0 {
			t.Errorf("Expected homepage priority %.1f, got %.1f", 1.0, homepage.Priority)
		}
		if homepage.ChangeFreq != "daily" {
			t.Errorf("Expected homepage changefreq %q, got %q", "daily", homepage.ChangeFreq)
		}

		// Check blog index
		blogIndex := urlset.URLs[1]
		if blogIndex.Loc != "https://example.com/blog" {
			t.Errorf("Expected blog index loc %q, got %q", "https://example.com/blog", blogIndex.Loc)
		}
		if blogIndex.Priority != 0.9 {
			t.Errorf("Expected blog index priority %.1f, got %.1f", 0.9, blogIndex.Priority)
		}

		// Check first post
		post1 := urlset.URLs[2]
		expectedURL := "https://example.com/blog/posts/test-post-1"
		if post1.Loc != expectedURL {
			t.Errorf("Expected post loc %q, got %q", expectedURL, post1.Loc)
		}
		if post1.Priority != 0.8 {
			t.Errorf("Expected post priority %.1f, got %.1f", 0.8, post1.Priority)
		}
		if post1.ChangeFreq != "monthly" {
			t.Errorf("Expected post changefreq %q, got %q", "monthly", post1.ChangeFreq)
		}

		expectedDate := "2024-01-15"
		if post1.LastMod != expectedDate {
			t.Errorf("Expected post lastmod %q, got %q", expectedDate, post1.LastMod)
		}
	})

	t.Run("sitemap with empty posts", func(t *testing.T) {
		posts := []*models.Post{}

		config := SitemapConfig{
			SiteURL:  "https://example.com",
			BlogPath: "/blog",
		}

		result, err := GenerateSitemap(posts, config)
		if err != nil {
			t.Fatalf("GenerateSitemap failed: %v", err)
		}

		var urlset URLSet
		if err := xml.Unmarshal([]byte(result), &urlset); err != nil {
			t.Fatalf("Failed to parse generated XML: %v", err)
		}

		// Should have: homepage + blog path = 2 URLs
		if len(urlset.URLs) != 2 {
			t.Fatalf("Expected 2 URLs, got %d", len(urlset.URLs))
		}
	})

	t.Run("sitemap with root blog path", func(t *testing.T) {
		posts := []*models.Post{
			{
				Title:       "Test Post",
				Slug:        "test",
				Date:        time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
				Description: "Test",
			},
		}

		config := SitemapConfig{
			SiteURL:  "https://example.com",
			BlogPath: "/",
		}

		result, err := GenerateSitemap(posts, config)
		if err != nil {
			t.Fatalf("GenerateSitemap failed: %v", err)
		}

		var urlset URLSet
		if err := xml.Unmarshal([]byte(result), &urlset); err != nil {
			t.Fatalf("Failed to parse generated XML: %v", err)
		}

		// Should have: homepage + 1 post = 2 URLs (no separate blog index)
		if len(urlset.URLs) != 2 {
			t.Fatalf("Expected 2 URLs (no blog index when path is /), got %d", len(urlset.URLs))
		}

		// First should be homepage
		if urlset.URLs[0].Loc != "https://example.com/" {
			t.Errorf("Expected first URL to be homepage, got %q", urlset.URLs[0].Loc)
		}

		// Second should be the post
		expectedPostURL := "https://example.com//posts/test"
		if urlset.URLs[1].Loc != expectedPostURL {
			t.Errorf("Expected post URL %q, got %q", expectedPostURL, urlset.URLs[1].Loc)
		}
	})

	t.Run("sitemap with empty blog path", func(t *testing.T) {
		posts := []*models.Post{
			{
				Title:       "Test Post",
				Slug:        "test",
				Date:        time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
				Description: "Test",
			},
		}

		config := SitemapConfig{
			SiteURL:  "https://example.com",
			BlogPath: "",
		}

		result, err := GenerateSitemap(posts, config)
		if err != nil {
			t.Fatalf("GenerateSitemap failed: %v", err)
		}

		var urlset URLSet
		if err := xml.Unmarshal([]byte(result), &urlset); err != nil {
			t.Fatalf("Failed to parse generated XML: %v", err)
		}

		// Should have: homepage + 1 post = 2 URLs (no blog index when path is empty)
		if len(urlset.URLs) != 2 {
			t.Fatalf("Expected 2 URLs, got %d", len(urlset.URLs))
		}
	})

	t.Run("sitemap with custom blog path", func(t *testing.T) {
		posts := []*models.Post{
			{
				Title:       "Test Post",
				Slug:        "test",
				Date:        time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
				Description: "Test",
			},
		}

		config := SitemapConfig{
			SiteURL:  "https://example.com",
			BlogPath: "/custom/path",
		}

		result, err := GenerateSitemap(posts, config)
		if err != nil {
			t.Fatalf("GenerateSitemap failed: %v", err)
		}

		expectedURL := "https://example.com/custom/path/posts/test"
		if !strings.Contains(result, expectedURL) {
			t.Errorf("Expected result to contain URL %q", expectedURL)
		}

		// Check blog index path
		expectedBlogIndex := "https://example.com/custom/path"
		if !strings.Contains(result, expectedBlogIndex) {
			t.Errorf("Expected result to contain blog index %q", expectedBlogIndex)
		}
	})

	t.Run("sitemap with multiple posts", func(t *testing.T) {
		posts := []*models.Post{
			{
				Title:       "Post 1",
				Slug:        "post-1",
				Date:        time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
				Description: "First",
			},
			{
				Title:       "Post 2",
				Slug:        "post-2",
				Date:        time.Date(2024, 2, 1, 12, 0, 0, 0, time.UTC),
				Description: "Second",
			},
			{
				Title:       "Post 3",
				Slug:        "post-3",
				Date:        time.Date(2024, 3, 1, 12, 0, 0, 0, time.UTC),
				Description: "Third",
			},
		}

		config := SitemapConfig{
			SiteURL:  "https://example.com",
			BlogPath: "/blog",
		}

		result, err := GenerateSitemap(posts, config)
		if err != nil {
			t.Fatalf("GenerateSitemap failed: %v", err)
		}

		var urlset URLSet
		if err := xml.Unmarshal([]byte(result), &urlset); err != nil {
			t.Fatalf("Failed to parse generated XML: %v", err)
		}

		// homepage + blog index + 3 posts = 5
		if len(urlset.URLs) != 5 {
			t.Fatalf("Expected 5 URLs, got %d", len(urlset.URLs))
		}

		// Check all posts are included
		for i, post := range posts {
			expectedURL := "https://example.com/blog/posts/" + post.Slug
			found := false
			for _, url := range urlset.URLs {
				if url.Loc == expectedURL {
					found = true
					expectedDate := post.Date.Format("2006-01-02")
					if url.LastMod != expectedDate {
						t.Errorf("Post %d: expected lastmod %q, got %q", i, expectedDate, url.LastMod)
					}
					break
				}
			}
			if !found {
				t.Errorf("Post %d URL %q not found in sitemap", i, expectedURL)
			}
		}
	})
}

func TestGenerateSitemapIndex(t *testing.T) {
	t.Run("valid sitemap index", func(t *testing.T) {
		sitemapURLs := []string{
			"https://example.com/sitemap-posts.xml",
			"https://example.com/sitemap-pages.xml",
			"https://example.com/sitemap-tags.xml",
		}

		config := SitemapConfig{
			SiteURL:  "https://example.com",
			BlogPath: "/blog",
		}

		result, err := GenerateSitemapIndex(sitemapURLs, config)
		if err != nil {
			t.Fatalf("GenerateSitemapIndex failed: %v", err)
		}

		// Check XML header
		if !strings.HasPrefix(result, xml.Header) {
			t.Error("Expected XML header")
		}

		// Parse XML to validate structure
		type SitemapEntry struct {
			Loc     string `xml:"loc"`
			LastMod string `xml:"lastmod,omitempty"`
		}

		type SitemapIndex struct {
			XMLName  xml.Name       `xml:"sitemapindex"`
			XMLNS    string         `xml:"xmlns,attr"`
			Sitemaps []SitemapEntry `xml:"sitemap"`
		}

		var index SitemapIndex
		if err := xml.Unmarshal([]byte(result), &index); err != nil {
			t.Fatalf("Failed to parse generated XML: %v", err)
		}

		// Verify xmlns
		if index.XMLNS != "http://www.sitemaps.org/schemas/sitemap/0.9" {
			t.Errorf("Expected xmlns %q, got %q", "http://www.sitemaps.org/schemas/sitemap/0.9", index.XMLNS)
		}

		// Verify all sitemaps are included
		if len(index.Sitemaps) != len(sitemapURLs) {
			t.Fatalf("Expected %d sitemaps, got %d", len(sitemapURLs), len(index.Sitemaps))
		}

		for i, expectedURL := range sitemapURLs {
			if index.Sitemaps[i].Loc != expectedURL {
				t.Errorf("Sitemap %d: expected loc %q, got %q", i, expectedURL, index.Sitemaps[i].Loc)
			}

			// LastMod should be set to current date
			if index.Sitemaps[i].LastMod == "" {
				t.Errorf("Sitemap %d: expected LastMod to be set", i)
			}

			// Verify it's in correct date format (YYYY-MM-DD)
			_, err := time.Parse("2006-01-02", index.Sitemaps[i].LastMod)
			if err != nil {
				t.Errorf("Sitemap %d: LastMod has invalid date format: %v", i, err)
			}
		}
	})

	t.Run("sitemap index with single sitemap", func(t *testing.T) {
		sitemapURLs := []string{
			"https://example.com/sitemap.xml",
		}

		config := SitemapConfig{
			SiteURL:  "https://example.com",
			BlogPath: "/",
		}

		result, err := GenerateSitemapIndex(sitemapURLs, config)
		if err != nil {
			t.Fatalf("GenerateSitemapIndex failed: %v", err)
		}

		if !strings.Contains(result, "https://example.com/sitemap.xml") {
			t.Error("Expected sitemap URL to be in index")
		}
	})

	t.Run("sitemap index with empty urls", func(t *testing.T) {
		sitemapURLs := []string{}

		config := SitemapConfig{
			SiteURL:  "https://example.com",
			BlogPath: "/blog",
		}

		result, err := GenerateSitemapIndex(sitemapURLs, config)
		if err != nil {
			t.Fatalf("GenerateSitemapIndex failed: %v", err)
		}

		type SitemapEntry struct {
			Loc     string `xml:"loc"`
			LastMod string `xml:"lastmod,omitempty"`
		}

		type SitemapIndex struct {
			XMLName  xml.Name       `xml:"sitemapindex"`
			XMLNS    string         `xml:"xmlns,attr"`
			Sitemaps []SitemapEntry `xml:"sitemap"`
		}

		var index SitemapIndex
		if err := xml.Unmarshal([]byte(result), &index); err != nil {
			t.Fatalf("Failed to parse generated XML: %v", err)
		}

		if len(index.Sitemaps) != 0 {
			t.Errorf("Expected 0 sitemaps, got %d", len(index.Sitemaps))
		}
	})

	t.Run("sitemap index with multiple sitemaps", func(t *testing.T) {
		sitemapURLs := []string{
			"https://example.com/sitemap-1.xml",
			"https://example.com/sitemap-2.xml",
			"https://example.com/sitemap-3.xml",
			"https://example.com/sitemap-4.xml",
			"https://example.com/sitemap-5.xml",
		}

		config := SitemapConfig{
			SiteURL:  "https://example.com",
			BlogPath: "/blog",
		}

		result, err := GenerateSitemapIndex(sitemapURLs, config)
		if err != nil {
			t.Fatalf("GenerateSitemapIndex failed: %v", err)
		}

		// Verify all URLs are in the result
		for _, url := range sitemapURLs {
			if !strings.Contains(result, url) {
				t.Errorf("Expected result to contain %q", url)
			}
		}
	})

	t.Run("sitemap index lastmod uses current date", func(t *testing.T) {
		sitemapURLs := []string{
			"https://example.com/sitemap.xml",
		}

		config := SitemapConfig{
			SiteURL:  "https://example.com",
			BlogPath: "/blog",
		}

		result, err := GenerateSitemapIndex(sitemapURLs, config)
		if err != nil {
			t.Fatalf("GenerateSitemapIndex failed: %v", err)
		}

		// Check that today's date appears in the lastmod field
		today := time.Now().Format("2006-01-02")
		if !strings.Contains(result, today) {
			t.Errorf("Expected result to contain today's date %q", today)
		}
	})
}

package sitemap

import (
	"encoding/xml"
	"fmt"
	"time"

	"github.com/harrydayexe/GoBlog/pkg/models"
)

// URLSet represents the root sitemap element
type URLSet struct {
	XMLName xml.Name `xml:"urlset"`
	XMLNS   string   `xml:"xmlns,attr"`
	URLs    []URL    `xml:"url"`
}

// URL represents a single URL entry in the sitemap
type URL struct {
	Loc        string  `xml:"loc"`
	LastMod    string  `xml:"lastmod,omitempty"`
	ChangeFreq string  `xml:"changefreq,omitempty"`
	Priority   float64 `xml:"priority,omitempty"`
}

// SitemapConfig contains configuration for sitemap generation
type SitemapConfig struct {
	SiteURL  string
	BlogPath string
}

// GenerateSitemap generates a sitemap.xml from posts
func GenerateSitemap(posts []*models.Post, config SitemapConfig) (string, error) {
	urlset := &URLSet{
		XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  make([]URL, 0, len(posts)+2), // +2 for index and blog path
	}

	// Add homepage
	urlset.URLs = append(urlset.URLs, URL{
		Loc:        config.SiteURL + "/",
		ChangeFreq: "daily",
		Priority:   1.0,
	})

	// Add blog index if it's not the homepage
	if config.BlogPath != "/" && config.BlogPath != "" {
		urlset.URLs = append(urlset.URLs, URL{
			Loc:        config.SiteURL + config.BlogPath,
			ChangeFreq: "daily",
			Priority:   0.9,
		})
	}

	// Add individual posts
	for _, post := range posts {
		postURL := fmt.Sprintf("%s%s/posts/%s", config.SiteURL, config.BlogPath, post.Slug)

		urlEntry := URL{
			Loc:        postURL,
			LastMod:    post.Date.Format("2006-01-02"),
			ChangeFreq: "monthly",
			Priority:   0.8,
		}

		urlset.URLs = append(urlset.URLs, urlEntry)
	}

	// Marshal to XML
	output, err := xml.MarshalIndent(urlset, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal sitemap: %w", err)
	}

	return xml.Header + string(output), nil
}

// GenerateSitemapIndex generates a sitemap index for large sites
func GenerateSitemapIndex(sitemapURLs []string, config SitemapConfig) (string, error) {
	type SitemapEntry struct {
		Loc     string `xml:"loc"`
		LastMod string `xml:"lastmod,omitempty"`
	}

	type SitemapIndex struct {
		XMLName  xml.Name       `xml:"sitemapindex"`
		XMLNS    string         `xml:"xmlns,attr"`
		Sitemaps []SitemapEntry `xml:"sitemap"`
	}

	index := &SitemapIndex{
		XMLNS:    "http://www.sitemaps.org/schemas/sitemap/0.9",
		Sitemaps: make([]SitemapEntry, 0, len(sitemapURLs)),
	}

	now := time.Now().Format("2006-01-02")
	for _, url := range sitemapURLs {
		index.Sitemaps = append(index.Sitemaps, SitemapEntry{
			Loc:     url,
			LastMod: now,
		})
	}

	// Marshal to XML
	output, err := xml.MarshalIndent(index, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal sitemap index: %w", err)
	}

	return xml.Header + string(output), nil
}

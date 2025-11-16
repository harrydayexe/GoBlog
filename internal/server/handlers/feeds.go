package handlers

import (
	"net/http"

	"github.com/harrydayexe/GoBlog/internal/gen/feeds"
	"github.com/harrydayexe/GoBlog/internal/gen/sitemap"
	"github.com/harrydayexe/GoBlog/internal/server/content"
)

// FeedHandlers handles feed-related HTTP requests
type FeedHandlers struct {
	loader     *content.Loader
	feedConfig FeedConfig
}

// FeedConfig contains configuration for generating feeds
type FeedConfig struct {
	Title       string
	Link        string
	Description string
	Author      string
	AuthorEmail string
	BlogPath    string
}

// NewFeedHandlers creates a new feed handlers instance
func NewFeedHandlers(loader *content.Loader, config FeedConfig) *FeedHandlers {
	return &FeedHandlers{
		loader:     loader,
		feedConfig: config,
	}
}

// HandleRSS serves the RSS feed
func (h *FeedHandlers) HandleRSS(w http.ResponseWriter, r *http.Request) {
	posts := h.loader.GetAll()

	rssConfig := feeds.RSSConfig{
		Title:       h.feedConfig.Title,
		Link:        h.feedConfig.Link,
		Description: h.feedConfig.Description,
		Language:    "en-us",
		Copyright:   "Copyright © " + h.feedConfig.Author,
		Author:      h.feedConfig.Author,
		BlogPath:    h.feedConfig.BlogPath,
	}

	rssContent, err := feeds.GenerateRSS(posts, rssConfig)
	if err != nil {
		http.Error(w, "Failed to generate RSS feed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(rssContent))
}

// HandleAtom serves the Atom feed
func (h *FeedHandlers) HandleAtom(w http.ResponseWriter, r *http.Request) {
	posts := h.loader.GetAll()

	atomConfig := feeds.AtomConfig{
		Title:       h.feedConfig.Title,
		Link:        h.feedConfig.Link,
		Description: h.feedConfig.Description,
		Author:      h.feedConfig.Author,
		AuthorEmail: h.feedConfig.AuthorEmail,
		BlogPath:    h.feedConfig.BlogPath,
	}

	atomContent, err := feeds.GenerateAtom(posts, atomConfig)
	if err != nil {
		http.Error(w, "Failed to generate Atom feed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/atom+xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(atomContent))
}

// HandleSitemap serves the XML sitemap
func (h *FeedHandlers) HandleSitemap(w http.ResponseWriter, r *http.Request) {
	posts := h.loader.GetAll()

	sitemapConfig := sitemap.SitemapConfig{
		SiteURL:  h.feedConfig.Link,
		BlogPath: h.feedConfig.BlogPath,
	}

	sitemapContent, err := sitemap.GenerateSitemap(posts, sitemapConfig)
	if err != nil {
		http.Error(w, "Failed to generate sitemap", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(sitemapContent))
}

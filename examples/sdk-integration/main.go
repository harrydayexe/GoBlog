package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/harrydayexe/GoBlog/pkg/server"
)

// This example shows how to integrate GoBlogServ into your own application
func main() {
	log.Println("Starting custom application with integrated blog...")

	// Create your main HTTP router
	mux := http.NewServeMux()

	// Add your own application routes
	mux.HandleFunc("GET /", homeHandler)
	mux.HandleFunc("GET /about", aboutHandler)
	mux.HandleFunc("GET /contact", contactHandler)

	// Configure the blog server
	blogOpts := server.DefaultOptions()
	blogOpts.ContentPath = "./posts"      // Path to your markdown posts
	blogOpts.EnableCache = true            // Enable caching for better performance
	blogOpts.CacheMaxMB = 100             // 100MB cache
	blogOpts.CacheTTL = 60 * time.Minute  // Cache for 60 minutes
	blogOpts.EnableSearch = true          // Enable full-text search
	blogOpts.SearchIndexPath = "./blog.bleve"
	blogOpts.RebuildIndex = false         // Only rebuild on first run
	blogOpts.PostsPerPage = 10
	blogOpts.Verbose = true               // Enable verbose logging

	// Create the blog server
	blogServer, err := server.New(blogOpts)
	if err != nil {
		log.Fatalf("Failed to create blog server: %v", err)
	}
	defer blogServer.Close()

	// Attach blog routes at /blog
	// This creates:
	//   GET /blog           - Blog index
	//   GET /blog/posts/{slug} - Individual posts
	//   GET /blog/tags/{tag}   - Posts by tag
	//   GET /blog/search       - Search (HTMX partial)
	blogServer.AttachRoutes(mux, "/blog")

	// You can also add a stats endpoint
	mux.HandleFunc("GET /blog/stats", func(w http.ResponseWriter, r *http.Request) {
		stats := blogServer.Stats()
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{
  "total_posts": %d,
  "total_tags": %d,
  "cache_hits": %d,
  "cache_misses": %d,
  "cache_hit_ratio": %.2f,
  "indexed_docs": %d
}`,
			stats.TotalPosts,
			len(stats.AllTags),
			stats.CacheHits,
			stats.CacheMisses,
			stats.CacheHitRatio,
			stats.IndexedDocs,
		)
	})

	// Start the server
	addr := ":8080"
	log.Printf("Server starting on http://localhost%s", addr)
	log.Printf("  - Home:    http://localhost%s/", addr)
	log.Printf("  - About:   http://localhost%s/about", addr)
	log.Printf("  - Blog:    http://localhost%s/blog", addr)
	log.Printf("  - Stats:   http://localhost%s/blog/stats", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>My Application</title>
    <style>
        body { font-family: sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
        nav { margin-bottom: 30px; }
        nav a { margin-right: 20px; text-decoration: none; color: #0066cc; }
        nav a:hover { text-decoration: underline; }
    </style>
</head>
<body>
    <nav>
        <a href="/">Home</a>
        <a href="/about">About</a>
        <a href="/contact">Contact</a>
        <a href="/blog">Blog</a>
    </nav>
    <h1>Welcome to My Application</h1>
    <p>This is a custom application with an integrated blog powered by GoBlogServ SDK.</p>
    <p>Visit the <a href="/blog">blog</a> to see it in action!</p>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>About - My Application</title>
    <style>
        body { font-family: sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
        nav { margin-bottom: 30px; }
        nav a { margin-right: 20px; text-decoration: none; color: #0066cc; }
        nav a:hover { text-decoration: underline; }
    </style>
</head>
<body>
    <nav>
        <a href="/">Home</a>
        <a href="/about">About</a>
        <a href="/contact">Contact</a>
        <a href="/blog">Blog</a>
    </nav>
    <h1>About</h1>
    <p>This application demonstrates how to integrate GoBlogServ as an SDK.</p>
    <p>You can combine your own routes with blog functionality seamlessly.</p>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

func contactHandler(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Contact - My Application</title>
    <style>
        body { font-family: sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
        nav { margin-bottom: 30px; }
        nav a { margin-right: 20px; text-decoration: none; color: #0066cc; }
        nav a:hover { text-decoration: underline; }
    </style>
</head>
<body>
    <nav>
        <a href="/">Home</a>
        <a href="/about">About</a>
        <a href="/contact">Contact</a>
        <a href="/blog">Blog</a>
    </nav>
    <h1>Contact</h1>
    <p>This is your contact page.</p>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

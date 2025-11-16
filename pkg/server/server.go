package server

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/harrydayexe/GoBlog/internal/server/components"
	"github.com/harrydayexe/GoBlog/internal/server/content"
	"github.com/harrydayexe/GoBlog/internal/server/handlers"
	"github.com/harrydayexe/GoBlog/internal/server/search"
	"github.com/harrydayexe/GoBlog/pkg/models"
)

// Server represents the blog server
type Server struct {
	options Options
	loader  *content.Loader
	cache   *content.Cache
	index   *search.Index
	logger  *log.Logger
}

// New creates a new blog server with the given options
func New(opts Options) (*Server, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid options: %w", err)
	}

	// Create logger
	logger := log.New(os.Stdout, "[GoBlogServ] ", log.LstdFlags)

	s := &Server{
		options: opts,
		logger:  logger,
	}

	// Initialize cache if enabled
	if opts.EnableCache {
		cache, err := content.NewCache(opts.CacheMaxMB, opts.CacheTTL)
		if err != nil {
			return nil, fmt.Errorf("failed to create cache: %w", err)
		}
		s.cache = cache
		if opts.Verbose {
			logger.Printf("Cache initialized (max: %dMB, TTL: %v)", opts.CacheMaxMB, opts.CacheTTL)
		}
	}

	// Initialize content loader
	loader, err := content.NewLoader(opts.ContentPath, s.cache)
	if err != nil {
		return nil, fmt.Errorf("failed to create content loader: %w", err)
	}
	s.loader = loader

	if opts.Verbose {
		posts := loader.GetAll()
		logger.Printf("Loaded %d posts from %s", len(posts), opts.ContentPath)
	}

	// Initialize search index if enabled
	if opts.EnableSearch {
		index, err := search.NewIndex(opts.SearchIndexPath, opts.RebuildIndex)
		if err != nil {
			return nil, fmt.Errorf("failed to create search index: %w", err)
		}
		s.index = index

		// Index all posts if rebuilding
		if opts.RebuildIndex {
			posts := loader.GetAll()
			if err := index.IndexPosts(posts); err != nil {
				return nil, fmt.Errorf("failed to index posts: %w", err)
			}
			if opts.Verbose {
				logger.Printf("Indexed %d posts", len(posts))
			}
		}

		if opts.Verbose {
			count, _ := index.Count()
			logger.Printf("Search index ready (%d documents)", count)
		}
	}

	// Set up file watching if enabled
	if opts.WatchFiles {
		// Set reload callback
		if opts.EventHandlers != nil && opts.EventHandlers.OnReload != nil {
			loader.SetReloadCallback(opts.EventHandlers.OnReload)
		}

		// Start watching
		if err := loader.Watch(); err != nil {
			return nil, fmt.Errorf("failed to start file watching: %w", err)
		}

		if opts.Verbose {
			logger.Printf("File watching enabled")
		}
	}

	return s, nil
}

// corsMiddleware creates a CORS middleware from configuration
func corsMiddleware(config *CORSConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			allowed := false
			for _, allowedOrigin := range config.AllowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					break
				}
			}

			if allowed {
				if len(config.AllowedOrigins) == 1 && config.AllowedOrigins[0] == "*" {
					w.Header().Set("Access-Control-Allow-Origin", "*")
				} else {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				}

				if config.AllowCredentials {
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}

				if len(config.AllowedMethods) > 0 {
					w.Header().Set("Access-Control-Allow-Methods", joinStrings(config.AllowedMethods, ", "))
				}

				if len(config.AllowedHeaders) > 0 {
					w.Header().Set("Access-Control-Allow-Headers", joinStrings(config.AllowedHeaders, ", "))
				}

				if config.MaxAge > 0 {
					w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", config.MaxAge))
				}
			}

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// joinStrings joins a slice of strings with a separator
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

// wrapHandler applies middleware to a handler
func wrapHandler(h http.HandlerFunc, middleware []func(http.Handler) http.Handler) http.Handler {
	var handler http.Handler = h
	// Apply middleware in reverse order so they execute in the order they were added
	for i := len(middleware) - 1; i >= 0; i-- {
		handler = middleware[i](handler)
	}
	return handler
}

// AttachRoutes attaches blog routes to the given mux at the specified base path
//
// Example:
//
//	mux := http.NewServeMux()
//	server.AttachRoutes(mux, "/blog")
//
// This will register:
//
//	GET /blog           - Blog index
//	GET /blog/posts/{slug} - Individual post
//	GET /blog/tags/{tag}   - Posts by tag
//	GET /blog/search       - Search (HTMX partial)
func (s *Server) AttachRoutes(mux *http.ServeMux, basePath string) {
	// Ensure base path starts with /
	if basePath == "" || basePath[0] != '/' {
		basePath = "/" + basePath
	}

	// Remove trailing slash
	if len(basePath) > 1 && basePath[len(basePath)-1] == '/' {
		basePath = basePath[:len(basePath)-1]
	}

	// Build middleware stack
	middleware := s.options.Middleware

	// Add CORS middleware if configured
	if s.options.CORS != nil {
		middleware = append([]func(http.Handler) http.Handler{corsMiddleware(s.options.CORS)}, middleware...)
	}

	// Create handlers with event support
	postHandlers := handlers.NewPostHandlers(s.loader)
	searchHandlers := handlers.NewSearchHandlers(s.loader, s.index)
	tagHandlers := handlers.NewTagHandlers(s.loader)
	feedHandlers := handlers.NewFeedHandlers(s.loader, handlers.FeedConfig{
		Title:       s.options.FeedTitle,
		Link:        s.options.SiteURL,
		Description: s.options.FeedDescription,
		Author:      s.options.FeedAuthor,
		AuthorEmail: s.options.FeedAuthorEmail,
		BlogPath:    basePath,
	})

	// Wrap handlers with event callbacks if configured
	indexHandler := postHandlers.HandleIndex
	postHandler := postHandlers.HandlePost
	searchHandler := searchHandlers.HandleSearch

	if s.options.EventHandlers != nil {
		// Wrap post handler to call OnPostView
		if s.options.EventHandlers.OnPostView != nil {
			originalPostHandler := postHandler
			postHandler = func(w http.ResponseWriter, r *http.Request) {
				slug := r.PathValue("slug")
				s.options.EventHandlers.OnPostView(slug, r)
				originalPostHandler(w, r)
			}
		}

		// Wrap search handler to call OnSearch
		if s.options.EventHandlers.OnSearch != nil {
			originalSearchHandler := searchHandler
			searchHandler = func(w http.ResponseWriter, r *http.Request) {
				query := r.URL.Query().Get("q")
				// We'll capture result count in a response wrapper if needed
				// For now just call with query
				s.options.EventHandlers.OnSearch(query, 0, r)
				originalSearchHandler(w, r)
			}
		}
	}

	// Register routes with middleware
	mux.Handle("GET "+basePath, wrapHandler(indexHandler, middleware))
	mux.Handle("GET "+basePath+"/", wrapHandler(indexHandler, middleware))
	mux.Handle("GET "+basePath+"/posts/{slug}", wrapHandler(postHandler, middleware))
	mux.Handle("GET "+basePath+"/tags/{tag}", wrapHandler(tagHandlers.HandleTag, middleware))
	mux.Handle("GET "+basePath+"/search", wrapHandler(searchHandler, middleware))
	mux.Handle("GET "+basePath+"/feed.xml", wrapHandler(feedHandlers.HandleRSS, middleware))
	mux.Handle("GET "+basePath+"/rss.xml", wrapHandler(feedHandlers.HandleRSS, middleware))
	mux.Handle("GET "+basePath+"/atom.xml", wrapHandler(feedHandlers.HandleAtom, middleware))
	mux.Handle("GET "+basePath+"/sitemap.xml", wrapHandler(feedHandlers.HandleSitemap, middleware))

	if s.options.Verbose {
		s.logger.Printf("Routes registered at %s", basePath)
		s.logger.Printf("  - Blog index, posts, tags, search")
		s.logger.Printf("  - RSS feed: %s/feed.xml, %s/rss.xml", basePath, basePath)
		s.logger.Printf("  - Atom feed: %s/atom.xml", basePath)
		s.logger.Printf("  - Sitemap: %s/sitemap.xml", basePath)
		if s.options.CORS != nil {
			s.logger.Printf("CORS enabled for origins: %v", s.options.CORS.AllowedOrigins)
		}
		if len(s.options.Middleware) > 0 {
			s.logger.Printf("Custom middleware: %d handlers", len(s.options.Middleware))
		}
	}
}

// Reload reloads all content from disk
func (s *Server) Reload() error {
	if s.loader == nil {
		return ErrServerNotStarted
	}

	if err := s.loader.Reload(); err != nil {
		return fmt.Errorf("failed to reload content: %w", err)
	}

	// Re-index if search is enabled
	if s.index != nil {
		posts := s.loader.GetAll()
		if err := s.index.IndexPosts(posts); err != nil {
			return fmt.Errorf("failed to re-index posts: %w", err)
		}
	}

	if s.options.Verbose {
		posts := s.loader.GetAll()
		s.logger.Printf("Reloaded %d posts", len(posts))
	}

	return nil
}

// Close closes the server and cleans up resources
func (s *Server) Close() error {
	// Stop file watching if enabled
	if s.loader != nil {
		if err := s.loader.StopWatching(); err != nil {
			s.logger.Printf("Error stopping file watcher: %v", err)
		}
	}

	if s.cache != nil {
		s.cache.Close()
	}

	if s.index != nil {
		if err := s.index.Close(); err != nil {
			return fmt.Errorf("failed to close search index: %w", err)
		}
	}

	return nil
}

// GetLoader returns the content loader (for advanced usage)
func (s *Server) GetLoader() *content.Loader {
	return s.loader
}

// GetIndex returns the search index (for advanced usage)
func (s *Server) GetIndex() *search.Index {
	return s.index
}

// RenderPost renders a single post as HTML (for custom layouts)
// Returns the HTML content that can be embedded in your own templates
func (s *Server) RenderPost(slug string) (string, error) {
	post, err := s.loader.GetBySlug(slug)
	if err != nil {
		return "", ErrPostNotFound
	}

	// Render using templ component
	var buf bytes.Buffer
	err = components.Post(post).Render(context.Background(), &buf)
	if err != nil {
		return "", NewBlogError(ErrCodeContentLoad, "failed to render post", err)
	}

	return buf.String(), nil
}

// RenderPostList renders a list of posts as HTML (for custom layouts)
// Returns the HTML content that can be embedded in your own templates
// If tags is empty, returns all posts. Use page for pagination (0-indexed)
func (s *Server) RenderPostList(tags []string, page int) (string, error) {
	var posts []*models.Post

	if len(tags) > 0 {
		// Filter by tags
		allPosts := s.loader.GetAll()
		for _, post := range allPosts {
			for _, tag := range tags {
				if post.HasTag(tag) {
					posts = append(posts, post)
					break
				}
			}
		}
	} else {
		var err error
		posts, _, err = s.loader.GetPaginated(page, s.options.PostsPerPage)
		if err != nil {
			return "", NewBlogError(ErrCodeContentLoad, "failed to get paginated posts", err)
		}
	}

	var buf bytes.Buffer
	err := components.PostList(posts).Render(context.Background(), &buf)
	if err != nil {
		return "", NewBlogError(ErrCodeContentLoad, "failed to render post list", err)
	}

	return buf.String(), nil
}

// Stats returns server statistics
func (s *Server) Stats() Stats {
	stats := Stats{
		TotalPosts: len(s.loader.GetAll()),
		AllTags:    s.loader.GetAllTags(),
	}

	if s.cache != nil {
		cacheStats := s.cache.Stats()
		stats.CacheHits = cacheStats.Hits
		stats.CacheMisses = cacheStats.Misses
		stats.CacheHitRatio = cacheStats.HitRatio
	}

	if s.index != nil {
		count, _ := s.index.Count()
		stats.IndexedDocs = count
	}

	return stats
}

// Stats represents server statistics
type Stats struct {
	TotalPosts    int
	AllTags       []string
	CacheHits     uint64
	CacheMisses   uint64
	CacheHitRatio float64
	IndexedDocs   uint64
}

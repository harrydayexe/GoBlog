package server

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/harrydayexe/GoBlog/internal/server/content"
	"github.com/harrydayexe/GoBlog/internal/server/handlers"
	"github.com/harrydayexe/GoBlog/internal/server/search"
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

	return s, nil
}

// AttachRoutes attaches blog routes to the given mux at the specified base path
//
// Example:
//   mux := http.NewServeMux()
//   server.AttachRoutes(mux, "/blog")
//
// This will register:
//   GET /blog           - Blog index
//   GET /blog/posts/{slug} - Individual post
//   GET /blog/tags/{tag}   - Posts by tag
//   GET /blog/search       - Search (HTMX partial)
func (s *Server) AttachRoutes(mux *http.ServeMux, basePath string) {
	// Ensure base path starts with /
	if basePath == "" || basePath[0] != '/' {
		basePath = "/" + basePath
	}

	// Remove trailing slash
	if len(basePath) > 1 && basePath[len(basePath)-1] == '/' {
		basePath = basePath[:len(basePath)-1]
	}

	// Create handlers
	postHandlers := handlers.NewPostHandlers(s.loader)
	searchHandlers := handlers.NewSearchHandlers(s.loader, s.index)
	tagHandlers := handlers.NewTagHandlers(s.loader)

	// Register routes
	mux.HandleFunc("GET "+basePath, postHandlers.HandleIndex)
	mux.HandleFunc("GET "+basePath+"/", postHandlers.HandleIndex)
	mux.HandleFunc("GET "+basePath+"/posts/{slug}", postHandlers.HandlePost)
	mux.HandleFunc("GET "+basePath+"/tags/{tag}", tagHandlers.HandleTag)
	mux.HandleFunc("GET "+basePath+"/search", searchHandlers.HandleSearch)

	if s.options.Verbose {
		s.logger.Printf("Routes registered at %s", basePath)
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

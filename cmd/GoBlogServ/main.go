package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/harrydayexe/GoBlog/pkg/server"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Load from environment variables first
	opts, err := server.LoadFromEnv()
	if err != nil {
		log.Fatalf("Failed to load environment variables: %v", err)
	}

	// Define flags with current values from env as defaults
	var (
		port         = flag.Int("port", 8080, "Port to listen on (env: GOBLOG_PORT)")
		host         = flag.String("host", "localhost", "Host to bind to (env: GOBLOG_HOST)")
		content      = flag.String("content", opts.ContentPath, "Path to markdown posts (env: GOBLOG_CONTENT_PATH)")
		config       = flag.String("config", "", "Path to config file (not implemented yet)")
		basePath     = flag.String("base-path", "/blog", "Base URL path for blog routes (env: GOBLOG_BASE_PATH)")
		cacheMaxMB   = flag.Int64("cache-max-mb", opts.CacheMaxMB, "Max cache size in MB (env: GOBLOG_CACHE_MAX_MB)")
		noCache      = flag.Bool("no-cache", !opts.EnableCache, "Disable caching")
		noSearch     = flag.Bool("no-search", !opts.EnableSearch, "Disable search index")
		rebuildIndex = flag.Bool("rebuild-index", opts.RebuildIndex, "Rebuild search index on startup (env: GOBLOG_REBUILD_INDEX)")
		verbose      = flag.Bool("verbose", opts.Verbose, "Enable verbose logging (env: GOBLOG_VERBOSE)")
		versionFlag  = flag.Bool("version", false, "Show version information")
		help         = flag.Bool("help", false, "Show help message")
	)

	flag.Parse()

	if *help {
		printHelp()
		os.Exit(0)
	}

	if *versionFlag {
		fmt.Printf("GoBlogServ v%s\n", version)
		fmt.Printf("Commit:  %s\n", commit)
		fmt.Printf("Built:   %s\n", date)
		os.Exit(0)
	}

	// Warn if config is specified (not yet implemented)
	if *config != "" {
		log.Printf("Warning: -config flag is not yet implemented, using environment variables and command-line flags")
	}

	// Override with command-line flags (flags take precedence)
	opts.ContentPath = *content
	opts.CacheMaxMB = *cacheMaxMB
	opts.EnableCache = !*noCache
	opts.EnableSearch = !*noSearch
	opts.RebuildIndex = *rebuildIndex
	opts.Verbose = *verbose

	// Create the blog server
	log.Printf("GoBlogServ v%s (%s)", version, commit)
	log.Printf("Starting server on %s:%d", *host, *port)
	log.Printf("Content path: %s", *content)
	log.Printf("Blog routes: %s/*", *basePath)

	blogServer, err := server.New(opts)
	if err != nil {
		log.Fatalf("Failed to create blog server: %v", err)
	}
	defer blogServer.Close()

	// Create HTTP server with routes
	mux := http.NewServeMux()

	// Attach blog routes
	blogServer.AttachRoutes(mux, *basePath)

	// Add health check endpoint
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	})

	// Add stats endpoint
	mux.HandleFunc("GET /stats", func(w http.ResponseWriter, r *http.Request) {
		stats := blogServer.Stats()
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"total_posts": %d, "total_tags": %d, "cache_hits": %d, "cache_misses": %d, "cache_hit_ratio": %.2f, "indexed_docs": %d}`,
			stats.TotalPosts,
			len(stats.AllTags),
			stats.CacheHits,
			stats.CacheMisses,
			stats.CacheHitRatio,
			stats.IndexedDocs,
		)
	})

	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", *host, *port)
	httpServer := &http.Server{
		Addr:           addr,
		Handler:        mux,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	// Handle graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Server listening on http://%s", addr)
		log.Printf("Visit http://%s%s to see your blog", addr, *basePath)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-stop

	log.Println("Shutting down server...")
	if err := httpServer.Close(); err != nil {
		log.Printf("Error closing server: %v", err)
	}

	log.Println("Server stopped")
}

func printHelp() {
	fmt.Fprintf(os.Stdout, "GoBlogServ - Opinionated web server for markdown blogs\n\n")
	fmt.Fprintf(os.Stdout, "Version: %s\n", version)
	fmt.Fprintf(os.Stdout, "Commit:  %s\n", commit)
	fmt.Fprintf(os.Stdout, "Built:   %s\n\n", date)
	fmt.Fprintf(os.Stdout, "Usage:\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stdout, "\nEnvironment Variables:\n")
	fmt.Fprintf(os.Stdout, "  GOBLOG_CONTENT_PATH       Path to markdown posts (default: ./posts)\n")
	fmt.Fprintf(os.Stdout, "  GOBLOG_CACHE_ENABLED      Enable caching (default: true)\n")
	fmt.Fprintf(os.Stdout, "  GOBLOG_CACHE_MAX_MB       Max cache size in MB (default: 100)\n")
	fmt.Fprintf(os.Stdout, "  GOBLOG_CACHE_TTL          Cache TTL duration (default: 60m)\n")
	fmt.Fprintf(os.Stdout, "  GOBLOG_SEARCH_ENABLED     Enable search (default: true)\n")
	fmt.Fprintf(os.Stdout, "  GOBLOG_SEARCH_INDEX_PATH  Search index path (default: ./blog.bleve)\n")
	fmt.Fprintf(os.Stdout, "  GOBLOG_REBUILD_INDEX      Rebuild search index on startup (default: false)\n")
	fmt.Fprintf(os.Stdout, "  GOBLOG_POSTS_PER_PAGE     Posts per page (default: 10)\n")
	fmt.Fprintf(os.Stdout, "  GOBLOG_VERBOSE            Enable verbose logging (default: false)\n")
	fmt.Fprintf(os.Stdout, "\nConfiguration Priority:\n")
	fmt.Fprintf(os.Stdout, "  1. Command-line flags (highest priority)\n")
	fmt.Fprintf(os.Stdout, "  2. Environment variables\n")
	fmt.Fprintf(os.Stdout, "  3. Default values (lowest priority)\n")
	fmt.Fprintf(os.Stdout, "\nExamples:\n")
	fmt.Fprintf(os.Stdout, "  # Start server with default settings\n")
	fmt.Fprintf(os.Stdout, "  goblogserv\n\n")
	fmt.Fprintf(os.Stdout, "  # Start on port 3000 with verbose logging\n")
	fmt.Fprintf(os.Stdout, "  goblogserv -port 3000 -verbose\n\n")
	fmt.Fprintf(os.Stdout, "  # Configure via environment variables\n")
	fmt.Fprintf(os.Stdout, "  export GOBLOG_CONTENT_PATH=/var/blog/posts\n")
	fmt.Fprintf(os.Stdout, "  export GOBLOG_VERBOSE=true\n")
	fmt.Fprintf(os.Stdout, "  goblogserv\n\n")
	fmt.Fprintf(os.Stdout, "  # Docker: Use environment variables\n")
	fmt.Fprintf(os.Stdout, "  docker run -e GOBLOG_CONTENT_PATH=/posts -v ./posts:/posts goblog\n\n")
}

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
	var (
		port         = flag.Int("port", 8080, "Port to listen on")
		host         = flag.String("host", "localhost", "Host to bind to")
		content      = flag.String("content", "./posts", "Path to markdown posts")
		config       = flag.String("config", "", "Path to config file (not implemented yet)")
		basePath     = flag.String("base-path", "/blog", "Base URL path for blog routes")
		noCache      = flag.Bool("no-cache", false, "Disable caching")
		noSearch     = flag.Bool("no-search", false, "Disable search index")
		rebuildIndex = flag.Bool("rebuild-index", false, "Rebuild search index on startup")
		verbose      = flag.Bool("verbose", false, "Enable verbose logging")
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
		log.Printf("Warning: -config flag is not yet implemented, using command-line flags")
	}

	// Create server options
	opts := server.DefaultOptions()
	opts.ContentPath = *content
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
	fmt.Fprintf(os.Stdout, "\nExamples:\n")
	fmt.Fprintf(os.Stdout, "  # Start server with default settings\n")
	fmt.Fprintf(os.Stdout, "  goblogserv\n\n")
	fmt.Fprintf(os.Stdout, "  # Start on port 3000 with verbose logging\n")
	fmt.Fprintf(os.Stdout, "  goblogserv -port 3000 -verbose\n\n")
	fmt.Fprintf(os.Stdout, "  # Disable caching and search\n")
	fmt.Fprintf(os.Stdout, "  goblogserv -no-cache -no-search\n\n")
	fmt.Fprintf(os.Stdout, "  # Rebuild search index on startup\n")
	fmt.Fprintf(os.Stdout, "  goblogserv -rebuild-index\n\n")
}

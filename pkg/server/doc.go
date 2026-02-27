// Package server provides an HTTP server for serving generated blog content
// with support for live content updates via atomic handler hot-swapping.
//
// The server implements thread-safe handler replacement, allowing blog content
// to be regenerated and swapped in without dropping in-flight requests or
// requiring server restart.
//
// # Basic Usage
//
// Create and start a server:
//
//	import (
//	    "context"
//	    "log/slog"
//	    "os"
//	    "github.com/harrydayexe/GoBlog/v2/pkg/server"
//	    "github.com/harrydayexe/GoBlog/v2/pkg/config"
//	)
//
//	// Create server with posts from directory
//	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
//	postsFS := os.DirFS("posts/")
//
//	cfg := config.ServerConfig{
//	    Server: []config.ServerOption{
//	        config.WithPort(8080),
//	    },
//	}
//
//	srv, err := server.New(logger, postsFS, cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Start serving (blocks until interrupted)
//	if err := srv.Run(context.Background(), os.Stdout); err != nil {
//	    log.Fatal(err)
//	}
//
// # Live Content Updates
//
// Update blog content while server is running:
//
//	// Load updated posts
//	updatedFS := os.DirFS("posts/")
//
//	// Atomically swap in new content
//	if err := srv.UpdatePosts(updatedFS, context.Background()); err != nil {
//	    log.Printf("Update failed: %v", err)
//	}
//
// The handler swap is atomic - in-flight requests complete with the old handler
// while new requests immediately see the updated content.
//
// # Concurrency
//
// All Server methods are safe for concurrent use by multiple goroutines.
// The handler is stored in an atomic.Value, providing lock-free reads during
// request serving and atomic writes during updates. This design supports
// high-concurrency request handling without lock contention.
//
// Server implements http.Handler interface, delegating to the current handler
// via atomic load operations.
package server

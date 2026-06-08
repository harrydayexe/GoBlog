// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

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
//	    Server: []config.BaseServerOption{
//	        config.WithPort(8080),
//	        config.WithLogger(logger).AsServerOption(),
//	    },
//	}
//
//	srv, err := server.New(nil, postsFS, cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Start serving (blocks until interrupted)
//	if err := srv.Run(context.Background()); err != nil {
//	    log.Fatal(err)
//	}
//
// # HTML Extension Handling
//
// The server automatically accepts requests with or without .html suffixes.
// Requests for /posts/my-post.html are rewritten to /posts/my-post before
// routing, so both forms return the same content. This is handled by
// middleware.NewStripHTMLExtension from github.com/harrydayexe/GoWebUtilities,
// which is applied unconditionally inside Handler(). User-supplied middleware
// added via config.WithMiddleware sees the original .html URL before it is
// stripped.
//
// # Middleware
//
// The server supports pluggable HTTP middleware for cross-cutting concerns
// like logging, metrics, authentication, or rate limiting. Middleware uses
// the standard pattern from github.com/harrydayexe/GoWebUtilities/middleware.
//
// Adding middleware to a server:
//
//	import (
//	    "github.com/harrydayexe/GoBlog/v2/pkg/config"
//	    "github.com/harrydayexe/GoBlog/v2/pkg/server"
//	    "github.com/harrydayexe/GoWebUtilities/logging"
//	    "github.com/harrydayexe/GoWebUtilities/middleware"
//	)
//
//	// Create server with built-in logging middleware
//	cfg := config.ServerConfig{
//	    Server: []config.BaseServerOption{
//	        config.WithPort(8080),
//	        config.WithMiddleware(logging.New(logger)),
//	        config.WithLogger(logger).AsServerOption(),
//	    },
//	}
//
//	srv, err := server.New(nil, postsFS, cfg)
//
// Custom middleware can be added following the standard pattern:
//
//	func customMiddleware(h http.Handler) http.Handler {
//	    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//	        // Pre-request logic
//	        h.ServeHTTP(w, r)
//	        // Post-request logic
//	    })
//	}
//
//	cfg := config.ServerConfig{
//	    Server: []config.BaseServerOption{
//	        config.WithMiddleware(
//	            logging.New(logger),     // Built-in
//	            customMiddleware,        // Custom
//	        ),
//	    },
//	}
//
// Middleware are applied in order: the first middleware in the list is
// executed first (outermost wrapper). The middleware chain is reapplied
// whenever the handler is refreshed via UpdatePosts().
//
// Any middleware compatible with the standard http.Handler interface
// can be used, including third-party middleware packages following the
// func(http.Handler) http.Handler pattern.
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
// For automatic filesystem-triggered reloads, use pkg/watcher alongside
// UpdatePosts. The watcher watches a directory tree for changes and invokes
// a callback (debounced) on each change:
//
//	postsPath := "posts/"
//	w, err := watcher.New(postsPath, config.WithLogger(logger).AsWatcherOption())
//	if err != nil {
//	    log.Fatal(err)
//	}
//	go w.Run(ctx, func(ctx context.Context) {
//	    srv.UpdatePosts(os.DirFS(postsPath), ctx)
//	})
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

package server

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"time"

	"github.com/harrydayexe/GoBlog/v2/pkg/config"
	"github.com/harrydayexe/GoBlog/v2/pkg/generator"
	"github.com/harrydayexe/GoBlog/v2/pkg/templates"
)

// Server is an HTTP server that serves generated blog content with support
// for live content updates.
//
// The server uses atomic.Value to store its HTTP handler, enabling thread-safe
// handler hot-swapping without locks. This allows blog content to be regenerated
// and swapped in while serving requests, without dropping connections or requiring
// server restart.
//
// Server implements http.Handler interface, delegating requests to the current
// handler loaded atomically.
//
// All methods are safe for concurrent use by multiple goroutines.
type Server struct {
	logger   *slog.Logger
	mu       sync.RWMutex // protects postsDir
	postsDir fs.FS

	config.BlogRoot
	config.Port
	config.Host

	handler   atomic.Value // stores http.Handler
	generator *generator.Generator
}

// New creates a new Server instance with the specified configuration.
//
// The logger is used for structured logging throughout the server lifecycle.
// The posts filesystem contains the markdown blog posts to be served.
// The opts parameter configures server behavior via the functional options pattern.
//
// New initializes the server's HTTP handler by generating blog content from
// the posts filesystem. If initial generation fails, an error is returned.
//
// Returns an error if template rendering or initial blog generation fails.
// The server is ready to serve requests immediately upon successful return.
func New(logger *slog.Logger, posts fs.FS, opts config.ServerConfig) (*Server, error) {
	var templatesDir fs.FS
	if opts.TemplateDir != nil {
		logger.Debug("Using custom templates")
		templatesDir = opts.TemplateDir
	} else {
		logger.Debug("Using default templates")
		templatesDir = templates.Default
	}

	renderer, err := generator.NewTemplateRenderer(templatesDir)
	if err != nil {
		return nil, err
	}

	gen := generator.New(posts, renderer, opts.Gen...)

	srv := &Server{
		logger:    logger,
		postsDir:  posts,
		Port:      80,
		generator: gen,
	}

	for _, opt := range opts.Server {
		if opt.WithPortFunc != nil {
			opt.WithPortFunc(&srv.Port)
		} else if opt.WithHostFunc != nil {
			opt.WithHostFunc(&srv.Host)
		} else if opt.WithBlogRootFunc != nil {
			opt.WithBlogRootFunc(&srv.BlogRoot)
		}
	}

	// Initialize handler before returning
	if err := srv.refreshHandler(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to initialize handler: %w", err)
	}

	logger.Debug("Server created successfully")
	return srv, nil
}

// ServeHTTP implements http.Handler, delegating requests to the current handler.
// The handler is loaded atomically, allowing it to be safely updated via
// UpdatePosts while requests are being served.
//
// ServeHTTP is safe for concurrent use by multiple goroutines. It performs
// a lock-free atomic load of the current handler on each request.
//
// If the handler has not been initialized (nil), ServeHTTP returns a
// 503 Service Unavailable error.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h := s.handler.Load()
	if h == nil {
		s.logger.ErrorContext(r.Context(), "handler not initialized")
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
		return
	}
	h.(http.Handler).ServeHTTP(w, r)
}

// Run starts the HTTP server and blocks until interrupted via context cancellation
// or OS signal (SIGINT). It handles graceful shutdown with a 10-second timeout.
//
// The server uses atomic handler swapping, allowing UpdatePosts to be called
// while the server is running without interrupting in-flight requests.
//
// Run is safe for concurrent use, though typically only called once per Server.
// It captures OS interrupt signals (Ctrl+C) and initiates graceful shutdown.
//
// Returns an error only if server initialization fails. Shutdown errors are
// logged to stderr but don't prevent clean exit.
func (s *Server) Run(ctx context.Context, stdout io.Writer) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	httpServer := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", s.Host, s.Port),
		Handler: s,
	}

	var wg sync.WaitGroup

	// Track ListenAndServe goroutine
	wg.Go(func() {
		s.logger.Info(
			"server listening",
			slog.String("address", httpServer.Addr),
		)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
		}
	})

	// Track shutdown goroutine
	wg.Go(func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintf(os.Stderr, "error shutting down http server: %s\n", err)
		}
	})

	wg.Wait()
	return nil
}

// UpdatePosts updates the posts directory and refreshes the HTTP handler with
// the new content. This triggers a complete regeneration of the blog and an
// atomic swap of the HTTP handler.
//
// UpdatePosts is safe to call while the server is running and serving requests.
// The handler swap is atomic, ensuring that requests see either the old or new
// content without any intermediate inconsistent state.
//
// If handler refresh fails, an error is returned and the previous handler
// remains active, continuing to serve the old content.
func (s *Server) UpdatePosts(posts fs.FS, ctx context.Context) error {
	s.mu.Lock()
	s.postsDir = posts
	s.generator.PostsDir = posts
	s.mu.Unlock()

	if err := s.refreshHandler(ctx); err != nil {
		return fmt.Errorf("failed to refresh handler: %w", err)
	}
	return nil
}

// refreshHandler regenerates the blog content and updates the HTTP handler atomically.
// It generates fresh blog content from the current posts directory, creates a new
// handler with the updated content, and swaps it in atomically.
//
// refreshHandler is safe to call concurrently with ServeHTTP. The atomic store
// ensures that in-flight requests either see the old handler or the new handler,
// but never an inconsistent state.
//
// Returns an error if blog generation fails. In case of error, the previous
// handler remains active and continues serving requests.
func (s *Server) refreshHandler(ctx context.Context) error {
	s.logger.DebugContext(ctx, "Refreshing HTTP Handler")
	s.generator.DebugConfig(ctx)

	blog, err := s.generator.Generate(ctx)
	if err != nil {
		return fmt.Errorf("failed to generate blog: %w", err)
	}

	s.logger.DebugContext(ctx, "Creating New Handler for Server")

	handler := Handler(blog, s.logger, s.BlogRoot.AsOption())

	s.handler.Store(handler)

	return nil
}

// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package server

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/harrydayexe/GoBlog/v2/pkg/config"
	"github.com/harrydayexe/GoBlog/v2/pkg/generator"
	"github.com/harrydayexe/GoBlog/v2/pkg/templates"
	"github.com/harrydayexe/GoWebUtilities/middleware"
)

// probeState describes the current lifecycle state of the server, used to
// answer health-check probes.
type probeState int

const (
	probeStarting probeState = iota // server is binding and loading content
	probeReady                      // content loaded, ready to serve traffic
	probeFailed                     // content loading failed; see healthStatus.reason
)

// healthStatus is stored atomically and read by serveHealth.
type healthStatus struct {
	state  probeState
	reason string // non-empty when state == probeFailed
}

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
// When health checks are enabled via [config.WithHealthChecks], the server
// binds the HTTP listener before loading posts and templates. The three
// health-check endpoints (/healthz/live, /healthz/ready, /healthz/startup) are
// intercepted before the middleware stack and always available without auth.
//
// All methods are safe for concurrent use by multiple goroutines.
type Server struct {
	// mu protects postsDir, generator, and generator fields during initialize
	// and refreshHandler.
	mu       sync.RWMutex
	postsDir fs.FS

	config.BlogRoot
	config.Port
	config.Host
	config.Logger
	config.CacheControlTTL
	config.HealthChecks

	handler    atomic.Value // stores http.Handler
	health     atomic.Pointer[healthStatus]
	middleware []middleware.Middleware // middleware chain
	generator  *generator.Generator

	// deferred initialisation inputs — set in New, consumed by initialize.
	templatesDir fs.FS
	rendererOpts []config.RendererOption
	genOpts      []config.GeneratorOption
}

// New creates a new Server instance with the specified configuration.
//
// The posts filesystem contains the markdown blog posts to be served.
// The opts parameter configures server behavior via the functional options pattern.
//
// When health checks are disabled (the default), New initializes the server's
// HTTP handler synchronously by generating blog content from the posts
// filesystem. If initial generation fails, an error is returned and the server
// is not started.
//
// When health checks are enabled via [config.WithHealthChecks], New skips
// content generation and returns immediately. The HTTP handler is initialized
// asynchronously when [Run] is called, so probes can observe the startup state
// via /healthz/ready and /healthz/startup. In this mode New does not return an
// error on content-loading failures; instead, the failure is surfaced through
// the /healthz/ready endpoint.
//
// Returns an error if template rendering or initial blog generation fails (only
// in the synchronous / health-checks-disabled path).
//
// # Logger
//
// Supply a logger via [config.WithLogger] in cfg.Server:
//
//	cfg.Server = append(cfg.Server, config.WithLogger(myLogger).AsServerOption())
//
// Deprecated: the positional logger parameter will be removed in v3.0.0.
// Pass nil and supply the logger via config.WithLogger in cfg.Server instead.
// When both are provided, the config.WithLogger option takes precedence.
func New(logger *slog.Logger, posts fs.FS, opts config.ServerConfig) (*Server, error) {
	srv := &Server{
		postsDir:        posts,
		Port:            8080,
		CacheControlTTL: config.CacheControlTTL{TTL: time.Hour},
	}

	for _, opt := range opts.Server {
		if opt.WithPortFunc != nil {
			opt.WithPortFunc(&srv.Port)
		} else if opt.WithHostFunc != nil {
			opt.WithHostFunc(&srv.Host)
		} else if opt.WithBlogRootFunc != nil {
			opt.WithBlogRootFunc(&srv.BlogRoot)
		} else if opt.WithMiddlewareFunc != nil {
			opt.WithMiddlewareFunc(&srv.middleware)
		} else if opt.WithCacheControlFunc != nil {
			opt.WithCacheControlFunc(&srv.CacheControlTTL)
		} else if opt.WithLoggerFunc != nil {
			opt.WithLoggerFunc(&srv.Logger)
		} else if opt.WithHealthChecksFunc != nil {
			opt.WithHealthChecksFunc(&srv.HealthChecks)
		}
	}

	// Precedence: WithLogger option > positional logger arg > slog.Default().
	if srv.Logger.Logger == nil {
		if logger != nil {
			srv.Logger.Logger = logger
		} else {
			srv.Logger.Logger = slog.Default()
		}
	}

	// Resolve the template filesystem.
	var templatesDir fs.FS
	if opts.TemplateDir != nil {
		srv.Logger.Logger.Debug("Using custom templates")
		templatesDir = opts.TemplateDir
	} else {
		srv.Logger.Logger.Debug("Using default templates")
		templatesDir = templates.Default
	}

	// Build the generator option slice (merging caller options with logger).
	genOpts := make([]config.GeneratorOption, 0, len(opts.Gen)+1)
	genOpts = append(genOpts, opts.Gen...)
	genOpts = append(genOpts, srv.Logger.AsOption().AsGeneratorOption())

	// Store inputs needed by initialize (both sync and async paths use them).
	srv.templatesDir = templatesDir
	srv.rendererOpts = opts.RendererOpts
	srv.genOpts = genOpts

	if srv.HealthChecks.Enabled {
		// Async path: bind first, generate in Run. Mark as starting.
		srv.health.Store(&healthStatus{state: probeStarting})
		srv.Logger.Logger.Debug("Health checks enabled: deferring content initialisation to Run")
		return srv, nil
	}

	// Sync path (default): initialize before returning, same behaviour as before.
	if err := srv.initialize(context.Background()); err != nil {
		return nil, err
	}

	srv.Logger.Logger.Debug("Server created successfully")
	return srv, nil
}

// initialize builds the template renderer and generator, generates the initial
// blog content, and stores the HTTP handler atomically.
//
// Callers from New (sync path) invoke this before the server is published, so
// no lock is needed for the struct itself. Callers from Run (async path) invoke
// this concurrently with an already-published server, so the critical section
// that touches s.generator and calls refreshHandler is protected by s.mu.
func (s *Server) initialize(ctx context.Context) error {
	renderer, err := generator.NewTemplateRenderer(s.templatesDir, s.rendererOpts...)
	if err != nil {
		return fmt.Errorf("failed to create template renderer: %w", err)
	}

	s.mu.Lock()
	posts := s.postsDir
	gen := generator.New(posts, renderer, s.genOpts...)
	s.generator = gen
	s.generator.BlogRoot = s.BlogRoot
	err = s.refreshHandler(ctx)
	s.mu.Unlock()

	if err != nil {
		return fmt.Errorf("failed to initialize handler: %w", err)
	}
	return nil
}

// ServeHTTP implements http.Handler, delegating requests to the current handler.
// The handler is loaded atomically, allowing it to be safely updated via
// UpdatePosts while requests are being served.
//
// When health checks are enabled, ServeHTTP intercepts requests to
// /healthz/live, /healthz/ready, and /healthz/startup before the middleware
// stack, so they are always reachable without authentication and during async
// startup.
//
// ServeHTTP is safe for concurrent use by multiple goroutines. It performs
// a lock-free atomic load of the current handler on each request.
//
// If the handler has not been initialized (nil), ServeHTTP returns a
// 503 Service Unavailable error.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.HealthChecks.Enabled {
		switch r.URL.Path {
		case "/healthz/live", "/healthz/ready", "/healthz/startup":
			s.serveHealth(w, r)
			return
		}
	}

	h := s.handler.Load()
	if h == nil {
		s.Logger.Logger.ErrorContext(r.Context(), "handler not initialized")
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
		return
	}
	h.(http.Handler).ServeHTTP(w, r)
}

// serveHealth handles requests to the /healthz/* endpoints.
// Only GET is accepted; any other method returns 405 Method Not Allowed.
//
// /healthz/live always returns 200 OK.
// /healthz/ready and /healthz/startup return 200 OK once content has loaded
// and 503 Service Unavailable while starting or after a load failure.
func (s *Server) serveHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	switch r.URL.Path {
	case "/healthz/live":
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprint(w, "ok")

	case "/healthz/ready", "/healthz/startup":
		st := s.health.Load()
		if st == nil || st.state != probeReady {
			reason := "starting"
			if st != nil && st.state == probeFailed {
				reason = "not ready: " + st.reason
			}
			http.Error(w, reason, http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprint(w, "ok")
	}
}

// Run starts the HTTP server and blocks until interrupted via context cancellation
// or OS signal (SIGINT, SIGTERM, SIGHUP). It handles graceful shutdown with a
// 10-second timeout.
//
// When health checks are enabled, Run launches content initialisation in a
// background goroutine so that the HTTP listener is available immediately for
// probe traffic. The /healthz/ready and /healthz/startup endpoints return
// 503 until initialisation completes (or 503 with a reason if it fails).
//
// The server uses atomic handler swapping, allowing UpdatePosts to be called
// while the server is running without interrupting in-flight requests.
//
// Run is safe for concurrent use, though typically only called once per Server.
// It captures OS signals and initiates graceful shutdown.
//
// Returns an error if the server fails to bind or listen. Shutdown errors are
// logged to stderr but don't prevent clean exit.
func (s *Server) Run(ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	httpServer := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", s.Host, s.Port),
		Handler: s,
	}

	var wg sync.WaitGroup

	// When health checks are enabled, initialize content asynchronously so the
	// listener binds immediately and probes can reach /healthz/* during startup.
	if s.HealthChecks.Enabled {
		wg.Go(func() {
			if err := s.initialize(ctx); err != nil {
				s.health.Store(&healthStatus{state: probeFailed, reason: err.Error()})
				s.Logger.Logger.ErrorContext(ctx, "async content initialisation failed", slog.Any("error", err))
				return
			}
			s.health.Store(&healthStatus{state: probeReady})
			s.Logger.Logger.DebugContext(ctx, "async content initialisation complete")
		})
	}

	// Track ListenAndServe goroutine. On bind/listen failure, cancel ctx so the
	// shutdown goroutine wakes and wg.Wait() can return.
	var listenErr error
	wg.Go(func() {
		s.Logger.Logger.Info(
			"server listening",
			slog.String("address", httpServer.Addr),
		)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			listenErr = err
			cancel()
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
	return listenErr
}

// UpdatePosts updates the posts directory and refreshes the HTTP handler with
// the new content. This triggers a complete regeneration of the blog and an
// atomic swap of the HTTP handler.
//
// UpdatePosts is safe to call while the server is running and serving requests.
// The handler swap is atomic, ensuring that requests see either the old or new
// content without any intermediate inconsistent state.
//
// If the server has not yet completed its initial content load (health-checks
// async path), UpdatePosts returns an error immediately rather than racing with
// the initialisation goroutine.
//
// If handler refresh fails, an error is returned and the previous handler
// remains active, continuing to serve the old content.
func (s *Server) UpdatePosts(posts fs.FS, ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.generator == nil {
		return fmt.Errorf("server not yet initialised; cannot update posts before initial content load completes")
	}

	s.postsDir = posts
	s.generator.PostsDir = posts
	if err := s.refreshHandler(ctx); err != nil {
		return fmt.Errorf("failed to refresh handler: %w", err)
	}

	// Recover health state after a successful reload (e.g. after a prior failure
	// in watch mode). Only meaningful when health checks are enabled.
	if s.HealthChecks.Enabled {
		s.health.Store(&healthStatus{state: probeReady})
	}

	return nil
}

// refreshHandler regenerates the blog content and updates the HTTP handler atomically.
// It generates fresh blog content from the current posts directory, creates a new
// handler with the updated content, and swaps it in atomically.
//
// Callers must either hold s.mu or be the sole goroutine accessing the server
// (e.g. during New before the Server is published).
//
// Returns an error if blog generation fails. In case of error, the previous
// handler remains active and continues serving requests.
func (s *Server) refreshHandler(ctx context.Context) error {
	s.Logger.Logger.DebugContext(ctx, "Refreshing HTTP Handler")
	s.generator.DebugConfig(ctx)

	blog, err := s.generator.Generate(ctx)
	if err != nil {
		return fmt.Errorf("failed to generate blog: %w", err)
	}

	s.Logger.Logger.DebugContext(ctx, "Creating New Handler for Server")

	handler := Handler(blog, nil, s.BlogRoot.AsOption(), s.Logger.AsOption())

	// Apply middleware stack if configured
	if len(s.middleware) > 0 {
		stack := middleware.CreateStack(s.middleware...)
		handler = stack(handler)
	}

	// Apply cache-control as the outermost layer so it covers every route.
	if s.CacheControlTTL.TTL > 0 {
		handler = middleware.NewCacheControl(s.CacheControlTTL.TTL)(handler)
	}

	s.handler.Store(handler)

	return nil
}

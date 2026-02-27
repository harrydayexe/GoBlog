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
	"time"

	"github.com/harrydayexe/GoBlog/v2/pkg/config"
	"github.com/harrydayexe/GoBlog/v2/pkg/generator"
	"github.com/harrydayexe/GoBlog/v2/pkg/templates"
)

type Server struct {
	logger   *slog.Logger
	postsDir fs.FS

	config.BlogRoot
	config.Port
	config.Host

	handler   http.Handler
	generator *generator.Generator
}

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

	logger.Debug("Server created successfully")
	return srv, nil
}

func (s *Server) Run(ctx context.Context, stdout io.Writer) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	httpServer := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", s.Host, s.Port),
		Handler: s.handler,
	}

	go func() {
		s.logger.Info(
			"server listening",
			slog.String("address", httpServer.Addr),
		)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
		}
	}()
	var wg sync.WaitGroup
	wg.Go(func() {
		<-ctx.Done()
		// make a new context for the Shutdown
		shutdownCtx := context.Background()
		shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintf(os.Stderr, "error shutting down http server: %s\n", err)
		}
	})
	wg.Wait()
	return nil

}

func (s *Server) UpdatePosts(posts fs.FS, ctx context.Context) error {
	s.postsDir = posts
	s.refreshHandler(ctx)
	return nil
}

func (s *Server) refreshHandler(ctx context.Context) error {
	s.logger.DebugContext(ctx, "Refreshing HTTP Handler")
	s.generator.DebugConfig(ctx)

	blog, err := s.generator.Generate(ctx)
	if err != nil {
		return err
	}

	s.logger.DebugContext(ctx, "Creating New Handler for Server")

	handler := Handler(blog, s.BlogRoot.AsOption())

	s.handler = handler

	return nil
}

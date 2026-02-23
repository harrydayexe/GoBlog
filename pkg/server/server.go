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
)

type ServerConfig struct {
	config.Port
}

type Server struct {
	logger   *slog.Logger
	postsDir fs.FS
	config   *ServerConfig

	handler http.Handler
}

func New(logger *slog.Logger, posts fs.FS, opts ...config.ServerOption) *Server {
	srv := &Server{
		logger:   logger,
		postsDir: posts,
		config: &ServerConfig{
			Port: 80,
		},
	}

	for _, opt := range opts {
		if opt.WithPortFunc != nil {
			opt.WithPortFunc(&srv.config.Port)
		}
	}

	return srv
}

func (s *Server) Run(ctx context.Context, stdout io.Writer) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.config.Port),
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

func (s *Server) UpdatePosts(posts fs.FS) error {
	s.postsDir = posts
	return nil
}

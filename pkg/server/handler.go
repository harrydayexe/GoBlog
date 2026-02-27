package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/harrydayexe/GoBlog/v2/pkg/config"
	"github.com/harrydayexe/GoBlog/v2/pkg/generator"
)

// HandlerConfig holds configuration options for the blog HTTP handler.
// It embeds config.BlogRoot to specify the root path where the blog is served.
type HandlerConfig struct {
	config.BlogRoot

	logger *slog.Logger
}

// Handler creates an HTTP handler that serves the generated blog content.
// It accepts a GeneratedBlog, a logger for error reporting, and optional
// configuration options to customize the handler behavior, such as setting
// a custom blog root path.
//
// The handler serves the following routes (assuming default root "/"):
//   - GET / and GET /posts - serves the blog index page
//   - GET /posts/{postName} - serves individual blog posts
//   - GET /tags - serves the tags index page
//   - GET /tags/{tagName} - serves tag-specific pages
//
// The handler is safe for concurrent use by multiple goroutines.
// It does not modify the GeneratedBlog instance.
func Handler(blog *generator.GeneratedBlog, logger *slog.Logger, opts ...config.BaseOption) http.Handler {
	cfg := HandlerConfig{
		BlogRoot: config.BlogRoot("/"),
		logger:   logger,
	}

	for _, opt := range opts {
		if opt.WithBlogRootFunc != nil {
			opt.WithBlogRootFunc(&cfg.BlogRoot)
		}
	}

	trimmed := strings.Trim(string(cfg.BlogRoot), "/")
	if trimmed == "" {
		cfg.BlogRoot = ""
	} else {
		cfg.BlogRoot = config.BlogRoot("/" + trimmed)
	}

	return generateHandler(cfg, blog)
}

func generateHandler(cfg HandlerConfig, blog *generator.GeneratedBlog) http.Handler {
	mux := http.NewServeMux()

	root := fmt.Sprintf("GET %s/", cfg.BlogRoot)
	mux.Handle(root+"posts", handleIndex(cfg, blog))
	mux.Handle(root, handleIndex(cfg, blog))
	mux.Handle(root+"posts/{postName}", handlePost(cfg, blog))

	mux.Handle(root+"tags", handleTagsIndex(cfg, blog))
	mux.Handle(root+"tags/{tagName}", handleTag(cfg, blog))

	return mux
}

func handleIndex(cfg HandlerConfig, blog *generator.GeneratedBlog) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if _, err := w.Write(blog.Index); err != nil {
			cfg.logger.ErrorContext(r.Context(), "failed to write index", "error", err)
			return
		}
	})
}

func handlePost(cfg HandlerConfig, blog *generator.GeneratedBlog) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postName := r.PathValue("postName")
		bits, prs := blog.Posts[postName]
		if !prs {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if _, err := w.Write(bits); err != nil {
			cfg.logger.ErrorContext(r.Context(), "failed to write post", "error", err, "post", postName)
			return
		}
	})
}

func handleTagsIndex(cfg HandlerConfig, blog *generator.GeneratedBlog) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if _, err := w.Write(blog.TagsIndex); err != nil {
			cfg.logger.ErrorContext(r.Context(), "failed to write tags index", "error", err)
			return
		}
	})
}

func handleTag(cfg HandlerConfig, blog *generator.GeneratedBlog) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tagName := r.PathValue("tagName")
		bits, prs := blog.Tags[tagName]
		if !prs {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if _, err := w.Write(bits); err != nil {
			cfg.logger.ErrorContext(r.Context(), "failed to write tag page", "error", err, "tag", tagName)
			return
		}
	})
}

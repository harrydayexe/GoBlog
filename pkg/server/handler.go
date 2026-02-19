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
// It accepts a GeneratedBlog and optional configuration options to customize
// the handler behavior, such as setting a custom blog root path.
//
// The handler serves the following routes (assuming default root "/"):
//   - GET / and GET /posts - serves the blog index page
//   - GET /posts/{postName} - serves individual blog posts
//   - GET /tags - serves the tags index page
//   - GET /tags/{postName} - serves tag-specific pages
//
// The handler is safe for concurrent use by multiple goroutines.
// It does not modify the GeneratedBlog instance.
func Handler(blog *generator.GeneratedBlog, opts ...config.BaseOption) http.Handler {
	cfg := HandlerConfig{
		BlogRoot: config.BlogRoot("/"),
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
	mux.Handle(root+"posts", handleIndex(blog))
	mux.Handle(root, handleIndex(blog))
	mux.Handle(root+"posts/{postName}", handlePost(blog))

	mux.Handle(root+"tags", handleTagsIndex(blog))
	mux.Handle(root+"tags/{postName}", handlePost(blog))

	return mux
}

func handleIndex(blog *generator.GeneratedBlog) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(blog.Index)
	})
}

func handlePost(blog *generator.GeneratedBlog) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postName := r.PathValue("postName")
		bits, prs := blog.Posts[postName]
		if !prs {
			w.WriteHeader(404)
			return
		}

		w.Write(bits)
	})
}

func handleTagsIndex(blog *generator.GeneratedBlog) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(blog.TagsIndex)
	})
}

func handleTag(blog *generator.GeneratedBlog) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postName := r.PathValue("tagName")
		bits, prs := blog.Tags[postName]
		if !prs {
			w.WriteHeader(404)
			return
		}

		w.Write(bits)
	})
}

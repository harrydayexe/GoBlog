// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package server

import (
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"path"
	"strings"

	"github.com/harrydayexe/GoBlog/v2/pkg/config"
	"github.com/harrydayexe/GoBlog/v2/pkg/generator"
	"github.com/harrydayexe/GoWebUtilities/middleware"
)

// HandlerConfig holds configuration options for the blog HTTP handler.
// It embeds config.BlogRoot to specify the root path where the blog is served
// and config.AssetsDir to specify the filesystem images are served from.
type HandlerConfig struct {
	config.BlogRoot
	config.Logger
	config.AssetsDir
}

// Handler creates an HTTP handler that serves the generated blog content.
// It accepts a GeneratedBlog and optional configuration options to customize
// the handler behavior, such as setting a custom blog root path or logger.
//
// The handler serves the following routes (assuming default root "/"):
//   - GET / and GET /posts - serves the blog index page
//   - GET /posts/{postName} - serves individual blog posts
//   - GET /tags - serves the tags index page (only if blog.TagsIndex is non-empty)
//   - GET /tags/{tagName} - serves tag-specific pages (only if blog.Tags is non-empty)
//   - GET /rss.xml - serves the site-wide RSS 2.0 feed (404 when feeds are disabled)
//   - GET /atom.xml - serves the site-wide Atom feed (404 when feeds are disabled)
//   - GET /tags/{tagName}.rss.xml - serves a per-tag RSS 2.0 feed (only if blog.Tags is non-empty)
//   - GET /tags/{tagName}.atom.xml - serves a per-tag Atom feed (only if blog.Tags is non-empty)
//   - GET /images/{path...} - serves files from the assets directory (only if
//     config.WithAssetsDir was supplied and its root is a readable directory)
//
// Asset files are served with [http.FileServerFS], so Content-Type, ETag,
// Last-Modified, conditional and range requests are handled. Requests for a
// directory, or for a path that does not exist in the assets filesystem,
// return 404; directory listings are never served.
//
// Tag routes are registered only when the blog contains tag content. When the
// generator is configured with config.WithDisableTags(), blog.Tags and
// blog.TagsIndex will be empty and the tag routes will not be registered.
//
// Feed routes are always registered, but return 404 when the generator did not
// produce feed content (i.e. when config.WithBaseURL was not set, or
// config.WithDisableFeeds was applied).
//
// The returned handler automatically strips .html suffixes from incoming request
// paths via middleware.NewStripHTMLExtension, so both /posts/foo and
// /posts/foo.html are routed to the same handler.
//
// The handler is safe for concurrent use by multiple goroutines.
// It does not modify the GeneratedBlog instance.
//
// # Logger
//
// Supply a logger via [config.WithLogger] in opts:
//
//	h := server.Handler(blog, nil, config.WithLogger(myLogger), config.WithBlogRoot("/blog/"))
//
// Deprecated: the positional logger parameter will be removed in v3.0.0.
// Pass nil and supply the logger via config.WithLogger in opts instead.
// When both are provided, the config.WithLogger option takes precedence.
func Handler(blog *generator.GeneratedBlog, logger *slog.Logger, opts ...config.BaseOption) http.Handler {
	cfg := HandlerConfig{
		BlogRoot: config.BlogRoot("/"),
	}

	for _, opt := range opts {
		if opt.WithBlogRootFunc != nil {
			opt.WithBlogRootFunc(&cfg.BlogRoot)
		} else if opt.WithLoggerFunc != nil {
			opt.WithLoggerFunc(&cfg.Logger)
		} else if opt.WithAssetsDirFunc != nil {
			opt.WithAssetsDirFunc(&cfg.AssetsDir)
		}
	}

	// Precedence: WithLogger option > positional logger arg > slog.Default().
	if cfg.Logger.Logger == nil {
		if logger != nil {
			cfg.Logger.Logger = logger
		} else {
			cfg.Logger.Logger = slog.Default()
		}
	}

	trimmed := strings.Trim(string(cfg.BlogRoot), "/")
	if trimmed == "" {
		cfg.BlogRoot = ""
	} else {
		cfg.BlogRoot = config.BlogRoot("/" + trimmed)
	}

	cfg.Logger.Logger.Debug("handlerConfig set", slog.String("BlogRoot", string(cfg.BlogRoot)))
	cfg.Logger.Logger.Debug("blog index page", slog.Int("bytes", len(blog.Index)))

	return middleware.NewStripHTMLExtension()(generateHandler(cfg, blog))
}

func generateHandler(cfg HandlerConfig, blog *generator.GeneratedBlog) http.Handler {
	mux := http.NewServeMux()

	root := fmt.Sprintf("GET %s/", cfg.BlogRoot)
	cfg.Logger.Logger.Debug("mux root set", slog.String("root", root))
	mux.Handle(root+"posts", handleIndex(cfg, blog))
	mux.Handle(root+"{$}", handleIndex(cfg, blog))
	mux.Handle(root+"posts/{postName}", handlePost(cfg, blog))

	if len(blog.Tags) > 0 || len(blog.TagsIndex) > 0 {
		mux.Handle(root+"tags", handleTagsIndex(cfg, blog))
		mux.Handle(root+"tags/{tagName}", handleTag(cfg, blog))
	}

	mux.Handle(root+"rss.xml", handleRSSFeed(cfg, blog))
	mux.Handle(root+"atom.xml", handleAtomFeed(cfg, blog))

	if cfg.AssetsDir.Enabled() {
		prefix := string(cfg.BlogRoot) + "/images/"
		mux.Handle(root+"images/", http.StripPrefix(prefix, handleAssets(cfg)))
	}

	return mux
}

// handleAssets serves files from the assets filesystem. Directories return 404
// so that http.FileServerFS never renders a directory listing.
func handleAssets(cfg HandlerConfig) http.Handler {
	fileServer := http.FileServerFS(cfg.AssetsDir.FS)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
		fi, err := fs.Stat(cfg.AssetsDir.FS, name)
		if err != nil || fi.IsDir() {
			http.NotFound(w, r)
			return
		}
		cfg.Logger.Logger.DebugContext(r.Context(), "serving asset", slog.String("path", name))
		fileServer.ServeHTTP(w, r)
	})
}

func handleIndex(cfg HandlerConfig, blog *generator.GeneratedBlog) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.Logger.Logger.DebugContext(r.Context(), "handling index page", slog.Int("bytes", len(blog.Index)))

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if _, err := w.Write(blog.Index); err != nil {
			cfg.Logger.Logger.ErrorContext(r.Context(), "failed to write index", "error", err)
			return
		}
	})
}

func handlePost(cfg HandlerConfig, blog *generator.GeneratedBlog) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.Logger.Logger.DebugContext(r.Context(), "handling post page")

		postName := strings.TrimSuffix(r.PathValue("postName"), ".html")
		bits, prs := blog.Posts[postName]
		if !prs {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		cfg.Logger.Logger.DebugContext(r.Context(), "resolved post name", slog.String("postName", postName))

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if _, err := w.Write(bits); err != nil {
			cfg.Logger.Logger.ErrorContext(r.Context(), "failed to write post", "error", err, "post", postName)
			return
		}
	})
}

func handleTagsIndex(cfg HandlerConfig, blog *generator.GeneratedBlog) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.Logger.Logger.DebugContext(r.Context(), "handling tag index")

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if _, err := w.Write(blog.TagsIndex); err != nil {
			cfg.Logger.Logger.ErrorContext(r.Context(), "failed to write tags index", "error", err)
			return
		}
	})
}

func handleTag(cfg HandlerConfig, blog *generator.GeneratedBlog) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.Logger.Logger.DebugContext(r.Context(), "handling tag page")

		rawName := r.PathValue("tagName")

		// Serve per-tag feeds when the path ends with a feed extension.
		// Go's net/http wildcard {tagName} captures the entire final path
		// segment, including any ".rss.xml" / ".atom.xml" suffix, so feed
		// paths cannot be registered as their own mux patterns and must be
		// demuxed here by inspecting the captured name.
		if tag, ok := strings.CutSuffix(rawName, ".rss.xml"); ok {
			bits := blog.TagRSSFeeds[tag]
			if len(bits) == 0 {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			cfg.Logger.Logger.DebugContext(r.Context(), "serving tag RSS feed", slog.String("tag", tag))
			w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
			if _, err := w.Write(bits); err != nil {
				cfg.Logger.Logger.ErrorContext(r.Context(), "failed to write tag RSS feed", "error", err, "tag", tag)
			}
			return
		}
		if tag, ok := strings.CutSuffix(rawName, ".atom.xml"); ok {
			bits := blog.TagAtomFeeds[tag]
			if len(bits) == 0 {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			cfg.Logger.Logger.DebugContext(r.Context(), "serving tag Atom feed", slog.String("tag", tag))
			w.Header().Set("Content-Type", "application/atom+xml; charset=utf-8")
			if _, err := w.Write(bits); err != nil {
				cfg.Logger.Logger.ErrorContext(r.Context(), "failed to write tag Atom feed", "error", err, "tag", tag)
			}
			return
		}

		tagName := strings.TrimSuffix(rawName, ".html")
		bits, prs := blog.Tags[tagName]
		if !prs {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		cfg.Logger.Logger.DebugContext(r.Context(), "resolved tag name", slog.String("tagName", tagName))

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if _, err := w.Write(bits); err != nil {
			cfg.Logger.Logger.ErrorContext(r.Context(), "failed to write tag page", "error", err, "tag", tagName)
			return
		}
	})
}

func handleRSSFeed(cfg HandlerConfig, blog *generator.GeneratedBlog) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.Logger.Logger.DebugContext(r.Context(), "handling site RSS feed")

		if len(blog.RSSFeed) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
		if _, err := w.Write(blog.RSSFeed); err != nil {
			cfg.Logger.Logger.ErrorContext(r.Context(), "failed to write RSS feed", "error", err)
		}
	})
}

func handleAtomFeed(cfg HandlerConfig, blog *generator.GeneratedBlog) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.Logger.Logger.DebugContext(r.Context(), "handling site Atom feed")

		if len(blog.AtomFeed) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/atom+xml; charset=utf-8")
		if _, err := w.Write(blog.AtomFeed); err != nil {
			cfg.Logger.Logger.ErrorContext(r.Context(), "failed to write Atom feed", "error", err)
		}
	})
}

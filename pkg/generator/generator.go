// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package generator

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/harrydayexe/GoBlog/v2/pkg/config"
	"github.com/harrydayexe/GoBlog/v2/pkg/models"
	"github.com/harrydayexe/GoBlog/v2/pkg/parser"
)

// Open Graph object types assigned to models.BaseData.OGType. Post pages
// describe a single article; every other page describes the site itself.
const (
	ogTypeArticle = "article"
	ogTypeWebsite = "website"
)

// Generator produces HTML output based on its input configuration.
// It reads markdown files from a configured filesystem and renders them
// as HTML using templates.
//
// A Generator is safe for concurrent use after creation, but Generate
// operations should not be called concurrently on the same instance.
type Generator struct {
	PostsDir fs.FS // The filesystem containing the input posts in markdown

	config.RawOutput
	config.DisableTags
	config.DisableReadingTime
	config.SiteTitle
	config.BlogRoot
	config.Environment
	config.CustomData
	config.HTMLPaths
	config.Logger
	config.BaseURL
	config.DisableFeeds
	config.FeedPostLimit
	config.SeriesFile
	ParserConfig parser.Config // The config to use when parsing

	renderer *TemplateRenderer
}

func (c Generator) String() string {
	return fmt.Sprintf(`Generator Config
- RawOutput           %t,
- DisableTags         %t,
- DisableReadingTime  %t,
- SiteTitle           %s,
- BlogRoot            %s,
- Environment         %s,
- CustomData keys     %d,
- HTMLPaths           %t,
- BaseURL             %s,
- DisableFeeds        %t,
- FeedPostLimit       %d,
- Series file         %s`,
		c.RawOutput,
		c.DisableTags.Disable,
		c.DisableReadingTime.Disable,
		c.SiteTitle,
		c.BlogRoot,
		c.Environment.Environment,
		len(c.CustomData.Data),
		c.HTMLPaths.Enable,
		c.BaseURL,
		c.DisableFeeds.Disable,
		c.FeedPostLimit.Limit,
		seriesFileDescription(c.SeriesFile),
	)
}

// seriesFileDescription renders the configured series file for the debug config
// dump, naming the path when series are enabled and "disabled" otherwise.
func seriesFileDescription(s config.SeriesFile) string {
	if !s.Enabled() {
		return "disabled"
	}
	return s.Path
}

// New creates a new Generator with the specified options.
// It returns an error if the configuration is invalid or if required
// resources cannot be initialized.
//
// Optional config.GeneratorOption values control behavior: config.WithRawOutput,
// config.WithDisableTags, config.WithDisableReadingTime, config.WithSiteTitle,
// config.WithBlogRoot, config.WithEnvironment, config.WithCustomData,
// config.WithBaseURL, config.WithDisableFeeds, config.WithFeedPostLimit,
// config.WithSeriesFile.
// The template renderer is supplied as a positional argument, not an option.
func New(posts fs.FS, renderer *TemplateRenderer, opts ...config.GeneratorOption) *Generator {
	gen := Generator{
		PostsDir: posts,
		renderer: renderer,
		BlogRoot: config.BlogRoot("/"),
	}

	// Run options on config
	for _, opt := range opts {
		if opt.WithRawOutputFunc != nil {
			opt.WithRawOutputFunc(&gen.RawOutput)
		} else if opt.WithDisableTagsFunc != nil {
			opt.WithDisableTagsFunc(&gen.DisableTags)
		} else if opt.WithDisableReadingTimeFunc != nil {
			opt.WithDisableReadingTimeFunc(&gen.DisableReadingTime)
		} else if opt.WithSiteTitleFunc != nil {
			opt.WithSiteTitleFunc(&gen.SiteTitle)
		} else if opt.WithBlogRootFunc != nil {
			opt.WithBlogRootFunc(&gen.BlogRoot)
		} else if opt.WithEnvironmentFunc != nil {
			opt.WithEnvironmentFunc(&gen.Environment)
		} else if opt.WithCustomDataFunc != nil {
			opt.WithCustomDataFunc(&gen.CustomData)
		} else if opt.WithHTMLPathsFunc != nil {
			opt.WithHTMLPathsFunc(&gen.HTMLPaths)
		} else if opt.WithBaseURLFunc != nil {
			opt.WithBaseURLFunc(&gen.BaseURL)
		} else if opt.WithDisableFeedsFunc != nil {
			opt.WithDisableFeedsFunc(&gen.DisableFeeds)
		} else if opt.WithFeedPostLimitFunc != nil {
			opt.WithFeedPostLimitFunc(&gen.FeedPostLimit)
		} else if opt.WithSeriesFileFunc != nil {
			opt.WithSeriesFileFunc(&gen.SeriesFile)
		} else if opt.WithLoggerFunc != nil {
			opt.WithLoggerFunc(&gen.Logger)
		}
	}

	if gen.Logger.Logger == nil {
		gen.Logger.Logger = slog.Default()
	}

	if gen.SiteTitle.SiteTitle == "" {
		gen.SiteTitle = config.SiteTitle{
			SiteTitle: "GoBlog",
		}
	}

	if gen.Environment.Environment == "" {
		gen.Environment = config.Environment{Environment: "local"}
	}

	return &gen
}

// Generate reads markdown post files from the configured filesystem and
// generates a complete static blog site as HTML.
//
// It returns a GeneratedBlog containing all rendered HTML content including
// individual post pages, tag pages, and the index page. The returned content
// is in-memory only; callers are responsible for writing to disk or serving
// via HTTP as needed.
//
// # Output Mode Behavior
//
// When RawOutput is disabled (default):
//   - Posts contain fully templated HTML pages
//   - Tags map contains rendered tag pages
//   - Index contains the complete index page with template
//
// When RawOutput is enabled (via config.WithRawOutput()):
//   - Posts contain only the Markdown-to-HTML conversion without templates
//   - Tags map will be empty (tag generation is skipped)
//   - Index will be empty or minimal
//   - Useful for custom integration scenarios
//
// Generate respects the provided context and will return early with
// context.Canceled or context.DeadlineExceeded if the context is canceled
// or times out.
//
// It returns an error if markdown files cannot be read, parsing fails, or
// template rendering encounters an error.
func (g *Generator) Generate(ctx context.Context) (*GeneratedBlog, error) {
	posts, err := g.parsePosts(ctx)
	if err != nil {
		return nil, err
	}

	// Step 2: If RawOutput mode, return immediately with raw HTML
	if g.RawOutput.RawOutput {
		g.Logger.Logger.InfoContext(ctx, "Raw output enabled, ignoring templates")
		return g.assembleRawBlog(posts), nil
	}

	// Step 3: Apply templates
	return g.assembleBlogWithTemplates(ctx, posts)
}

// parsePosts parses every markdown file in the posts filesystem, configuring the
// parser from the generator's own configuration.
func (g *Generator) parsePosts(ctx context.Context) (models.PostList, error) {
	g.Logger.Logger.DebugContext(ctx, "Creating parser for generate call")
	parserCfg := g.ParserConfig
	parserCfg.Logger = g.Logger.Logger
	parserCfg.BlogRoot = string(g.BlogRoot)
	p := parser.NewWithConfig(&parserCfg)

	return p.ParseDirectory(ctx, g.PostsDir)
}

// DebugConfig logs the current generator configuration at the debug level.
//
// This method is useful for troubleshooting and verifying configuration
// settings during development or when diagnosing issues. The output includes
// all generator configuration details and respects the provided context for
// structured logging.
//
// The log output will only appear if the logger is configured to show debug
// level messages.
func (g *Generator) DebugConfig(ctx context.Context) {
	g.Logger.Logger.DebugContext(ctx, g.String())
}

// pagePath returns the BaseData.Path value for a given page.
// kind must be one of "index", "post", "tag", "tagsIndex", "series", or
// "seriesIndex"; name is the slug or tag string (empty for the three index
// kinds).
func (g *Generator) pagePath(kind, name string) string {
	root := string(g.BlogRoot)

	var base string
	switch kind {
	case "index":
		if root == "/" {
			base = "/index"
		} else {
			// "/blog/" → "/blog"
			base = strings.TrimSuffix(root, "/")
		}
	case "post":
		base = root + "posts/" + name
	case "tag":
		base = root + "tags/" + name
	case "tagsIndex":
		base = root + "tags"
	case "series":
		base = root + "series/" + name
	case "seriesIndex":
		// The series index is a directory index: the outputter writes
		// series/index.html, so with HTML paths enabled the path has to name
		// that file rather than "series.html", which is never written.
		if g.HTMLPaths.Enable {
			return root + "series/index.html"
		}
		base = root + "series"
	}

	if g.HTMLPaths.Enable {
		return base + ".html"
	}

	// Clean-URL default: return the base as-is except for the index.
	if kind == "index" {
		return root // "/" or "/blog/"
	}
	return base
}

// canonicalURL returns the fully-qualified URL for a site-relative page path.
// It returns an empty string when no base URL is configured, so that templates
// omit canonical and Open Graph URL tags rather than emitting empty ones.
func (g *Generator) canonicalURL(path string) string {
	if g.BaseURL == "" {
		return ""
	}
	return absURL(string(g.BaseURL), path)
}

func (g *Generator) assembleRawBlog(posts models.PostList) *GeneratedBlog {
	blog := NewEmptyGeneratedBlog()

	for _, post := range posts {
		blog.Posts[post.Slug] = post.Content
	}

	return blog
}

func (g *Generator) assembleBlogWithTemplates(ctx context.Context, posts models.PostList) (*GeneratedBlog, error) {
	g.Logger.Logger.DebugContext(ctx, "Rendering posts with templates")

	// Check if renderer is available
	if g.renderer == nil {
		return nil, fmt.Errorf("template renderer is nil: cannot render templates without a template renderer")
	}

	blog := NewEmptyGeneratedBlog()

	tagsEnabled := !g.DisableTags.Disable

	// Series are opt-in: they exist only when a series file was configured.
	// They are independent of tags, so DisableTags has no bearing here.
	seriesEnabled := g.SeriesFile.Enabled()

	// feeds are active when not explicitly disabled AND a base URL is configured.
	feedsEnabled := !g.DisableFeeds.Disable && g.BaseURL != ""
	if !g.DisableFeeds.Disable && g.BaseURL == "" {
		g.Logger.Logger.InfoContext(ctx, "feeds enabled but no base URL configured; skipping feed generation (use WithBaseURL to enable feeds)")
	}

	// When tags are disabled, clear Post.Tags so template tag pills do not render.
	if !tagsEnabled {
		for _, post := range posts {
			post.Tags = nil
		}
	}

	// Compute reading time for each post unless disabled.
	if !g.DisableReadingTime.Disable {
		for _, post := range posts {
			post.ReadingTimeMinutes = minutesFromWords(wordCount(post.Content))
		}
	}

	// Sort posts by date descending
	posts.SortByDate()

	// Build site-wide feeds from the already-sorted post list.
	if feedsEnabled {
		rss, atom, err := g.buildFeeds(posts)
		if err != nil {
			return nil, fmt.Errorf("building site-wide feeds: %w", err)
		}
		blog.RSSFeed = rss
		blog.AtomFeed = atom
	}

	// Resolve the series before any page is rendered: post pages need to know
	// which series they belong to, and an invalid series file must fail the whole
	// generation rather than produce a half-linked site.
	var series []resolvedSeries
	seriesByPost := map[string]*models.PostSeries{}
	if seriesEnabled {
		var err error
		series, err = g.loadSeries(posts)
		if err != nil {
			return nil, err
		}
		seriesByPost = g.postSeriesContexts(series)

		// Series posts are listed with the post-card partial, which builds its
		// links from Post.BlogRoot.
		for _, s := range series {
			for _, post := range s.Posts {
				post.BlogRoot = string(g.BlogRoot)
			}
		}
	}

	// Render individual post pages
	for _, post := range posts {
		path := g.pagePath("post", post.Slug)
		data := models.PostPageData{
			BaseData: models.BaseData{
				SiteTitle:     g.SiteTitle.SiteTitle,
				PageTitle:     post.Title,
				Description:   post.Description,
				Year:          time.Now().Year(),
				BlogRoot:      string(g.BlogRoot),
				Environment:   g.Environment.Environment,
				TagsEnabled:   tagsEnabled,
				SeriesEnabled: seriesEnabled,
				FeedsEnabled:  feedsEnabled,
				Custom:        g.CustomData.Data,
				Path:          path,
				CanonicalURL:  g.canonicalURL(path),
				OGType:        ogTypeArticle,
				Article: &models.ArticleMeta{
					PublishedTime:      post.Date,
					ModifiedTime:       post.LastEdited,
					Author:             post.Author,
					Tags:               post.Tags,
					ReadingTimeMinutes: post.ReadingTimeMinutes,
				},
			},
			Post:   post,
			Series: seriesByPost[post.SourcePath],
		}

		rendered, err := g.renderer.RenderPost(data)
		if err != nil {
			return nil, fmt.Errorf("failed to render post %s: %w", post.Slug, err)
		}

		blog.Posts[post.Slug] = rendered
	}

	// Enrich posts with BlogRoot for index page
	indexPosts := make([]*models.Post, len(posts))
	for i, post := range posts {
		indexPosts[i] = post
		indexPosts[i].BlogRoot = string(g.BlogRoot)
	}

	// Render index page
	indexPath := g.pagePath("index", "")
	indexData := models.IndexPageData{
		BaseData: models.BaseData{
			SiteTitle:     g.SiteTitle.SiteTitle,
			PageTitle:     "Home",
			Description:   "Recent blog posts",
			Year:          time.Now().Year(),
			BlogRoot:      string(g.BlogRoot),
			Environment:   g.Environment.Environment,
			TagsEnabled:   tagsEnabled,
			SeriesEnabled: seriesEnabled,
			FeedsEnabled:  feedsEnabled,
			Custom:        g.CustomData.Data,
			Path:          indexPath,
			CanonicalURL:  g.canonicalURL(indexPath),
			OGType:        ogTypeWebsite,
		},
		Posts:      indexPosts,
		TotalPosts: len(indexPosts),
	}

	index, err := g.renderer.RenderIndex(indexData)
	if err != nil {
		return nil, fmt.Errorf("failed to render index: %w", err)
	}
	blog.Index = index

	if tagsEnabled {
		// Render tag pages
		allTags := posts.GetAllTags()
		for _, tag := range allTags {
			tagPosts := posts.FilterByTag(tag)

			// Enrich tag posts with BlogRoot
			for _, post := range tagPosts {
				post.BlogRoot = string(g.BlogRoot)
			}

			tagPath := g.pagePath("tag", tag)
			tagData := models.TagPageData{
				BaseData: models.BaseData{
					SiteTitle:     g.SiteTitle.SiteTitle,
					PageTitle:     "Tag: " + tag,
					Description:   fmt.Sprintf("Posts tagged with %s", tag),
					Year:          time.Now().Year(),
					BlogRoot:      string(g.BlogRoot),
					Environment:   g.Environment.Environment,
					TagsEnabled:   true,
					SeriesEnabled: seriesEnabled,
					FeedsEnabled:  feedsEnabled,
					Custom:        g.CustomData.Data,
					Path:          tagPath,
					// Tags come verbatim from front matter, so escape them
					// to keep characters like spaces and '#' out of the URL.
					CanonicalURL: g.canonicalURL(g.pagePath("tag", url.PathEscape(tag))),
					OGType:       ogTypeWebsite,
				},
				Tag:       tag,
				Posts:     tagPosts,
				PostCount: len(tagPosts),
			}

			rendered, err := g.renderer.RenderTag(tagData)
			if err != nil {
				return nil, fmt.Errorf("failed to render tag page %s: %w", tag, err)
			}

			blog.Tags[tag] = rendered

			// Build per-tag feeds alongside the tag page.
			if feedsEnabled {
				tagRSS, tagAtom, err := g.buildTagFeeds(tag, tagPosts)
				if err != nil {
					return nil, fmt.Errorf("building feeds for tag %q: %w", tag, err)
				}
				if tagRSS != nil {
					blog.TagRSSFeeds[tag] = tagRSS
					blog.TagAtomFeeds[tag] = tagAtom
				}
			}
		}

		// Render tags index page
		var tagInfos []models.TagInfo
		for _, tag := range allTags {
			tagPosts := posts.FilterByTag(tag)
			tagInfos = append(tagInfos, models.TagInfo{
				Name:      tag,
				PostCount: len(tagPosts),
			})
		}

		// Sort tags alphabetically (case-insensitive)
		sort.Slice(tagInfos, func(i, j int) bool {
			return strings.ToLower(tagInfos[i].Name) < strings.ToLower(tagInfos[j].Name)
		})

		tagsIndexPath := g.pagePath("tagsIndex", "")
		tagsIndexData := models.TagsIndexPageData{
			BaseData: models.BaseData{
				SiteTitle:     g.SiteTitle.SiteTitle,
				PageTitle:     "All Tags",
				Description:   "Browse all topics covered in this blog",
				Year:          time.Now().Year(),
				BlogRoot:      string(g.BlogRoot),
				Environment:   g.Environment.Environment,
				TagsEnabled:   true,
				SeriesEnabled: seriesEnabled,
				FeedsEnabled:  feedsEnabled,
				Custom:        g.CustomData.Data,
				Path:          tagsIndexPath,
				CanonicalURL:  g.canonicalURL(tagsIndexPath),
				OGType:        ogTypeWebsite,
			},
			Tags:      tagInfos,
			TotalTags: len(tagInfos),
		}

		tagsIndex, err := g.renderer.RenderTagsIndex(tagsIndexData)
		if err != nil {
			return nil, fmt.Errorf("failed to render tags index: %w", err)
		}
		blog.TagsIndex = tagsIndex
	}

	if seriesEnabled {
		// Render one page per series, plus the index listing them all in file
		// order so the author controls the ordering.
		seriesInfos := make([]models.SeriesInfo, 0, len(series))
		for _, s := range series {
			seriesPath := g.pagePath("series", s.Slug)
			seriesData := models.SeriesPageData{
				BaseData: models.BaseData{
					SiteTitle:     g.SiteTitle.SiteTitle,
					PageTitle:     "Series: " + s.Name,
					Description:   seriesDescription(s),
					Year:          time.Now().Year(),
					BlogRoot:      string(g.BlogRoot),
					Environment:   g.Environment.Environment,
					TagsEnabled:   tagsEnabled,
					SeriesEnabled: true,
					FeedsEnabled:  feedsEnabled,
					Custom:        g.CustomData.Data,
					Path:          seriesPath,
					CanonicalURL:  g.canonicalURL(seriesPath),
					OGType:        ogTypeWebsite,
				},
				Name:              s.Name,
				Slug:              s.Slug,
				SeriesDescription: s.Description,
				Posts:             s.Posts,
				PostCount:         len(s.Posts),
			}

			rendered, err := g.renderer.RenderSeries(seriesData)
			if err != nil {
				return nil, fmt.Errorf("failed to render series page %s: %w", s.Slug, err)
			}
			blog.Series[s.Slug] = rendered

			seriesInfos = append(seriesInfos, models.SeriesInfo{
				Name:        s.Name,
				Slug:        s.Slug,
				Description: s.Description,
				PostCount:   len(s.Posts),
				Path:        seriesPath,
			})
		}

		seriesIndexPath := g.pagePath("seriesIndex", "")
		seriesIndexData := models.SeriesIndexPageData{
			BaseData: models.BaseData{
				SiteTitle:     g.SiteTitle.SiteTitle,
				PageTitle:     "All Series",
				Description:   "Browse the multi-part series on this blog",
				Year:          time.Now().Year(),
				BlogRoot:      string(g.BlogRoot),
				Environment:   g.Environment.Environment,
				TagsEnabled:   tagsEnabled,
				SeriesEnabled: true,
				FeedsEnabled:  feedsEnabled,
				Custom:        g.CustomData.Data,
				Path:          seriesIndexPath,
				CanonicalURL:  g.canonicalURL(seriesIndexPath),
				OGType:        ogTypeWebsite,
			},
			Series:      seriesInfos,
			TotalSeries: len(seriesInfos),
		}

		seriesIndex, err := g.renderer.RenderSeriesIndex(seriesIndexData)
		if err != nil {
			return nil, fmt.Errorf("failed to render series index: %w", err)
		}
		blog.SeriesIndex = seriesIndex
	}

	return blog, nil
}

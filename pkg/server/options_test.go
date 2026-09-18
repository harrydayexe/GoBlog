// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package server

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/harrydayexe/GoBlog/v2/pkg/config"
	"github.com/harrydayexe/GoBlog/v2/pkg/templates"
	"github.com/harrydayexe/GoWebUtilities/middleware"
)

// pageTemplateFS renders body for every page type so tests can assert on the
// rendered output without depending on the default templates' markup.
func pageTemplateFS(body string) fstest.MapFS {
	return fstest.MapFS{
		"pages/index.tmpl":      {Data: []byte(body)},
		"pages/post.tmpl":       {Data: []byte(body)},
		"pages/tag.tmpl":        {Data: []byte(body)},
		"pages/tags-index.tmpl": {Data: []byte(body)},
	}
}

// TestNew_Defaults verifies the values New resolves when no options are given.
func TestNew_Defaults(t *testing.T) {
	t.Parallel()

	srv, err := New(makeInternalTestFS(t))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if srv.Port != 8080 {
		t.Errorf("Port = %d, want 8080", srv.Port)
	}
	if srv.CacheControlTTL.TTL != time.Hour {
		t.Errorf("CacheControlTTL.TTL = %v, want 1h", srv.CacheControlTTL.TTL)
	}
	if srv.Logger.Logger == nil {
		t.Error("Logger.Logger is nil, want the default logger")
	}
	if srv.TemplateDir.FS != templates.Default {
		t.Error("TemplateDir.FS is not the built-in template filesystem")
	}
	if srv.HealthChecks.Enabled {
		t.Error("HealthChecks.Enabled = true, want false")
	}
}

// TestNew_OptionsResolveIntoServerConfig verifies that every server option is
// applied to the config embedded in the Server.
func TestNew_OptionsResolveIntoServerConfig(t *testing.T) {
	t.Parallel()

	assets := fstest.MapFS{"pipeline.png": {Data: []byte("\x89PNG\r\n\x1a\n")}}
	noop := middleware.Middleware(func(next http.Handler) http.Handler { return next })

	srv, err := New(makeInternalTestFS(t),
		config.WithPort(9090),
		config.WithHost("127.0.0.1"),
		config.WithCacheControl(30*time.Minute),
		config.WithMiddleware(noop),
		config.WithBlogRoot("/blog/").AsServerOption(),
		config.WithAssetsDir(assets).AsServerOption(),
		config.WithTemplateDir(pageTemplateFS(`ok`)),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if srv.Port != 9090 {
		t.Errorf("Port = %d, want 9090", srv.Port)
	}
	if srv.Host != "127.0.0.1" {
		t.Errorf("Host = %q, want \"127.0.0.1\"", srv.Host)
	}
	if srv.CacheControlTTL.TTL != 30*time.Minute {
		t.Errorf("CacheControlTTL.TTL = %v, want 30m", srv.CacheControlTTL.TTL)
	}
	if len(srv.Middleware) != 1 {
		t.Errorf("len(Middleware) = %d, want 1", len(srv.Middleware))
	}
	if srv.BlogRoot != "/blog/" {
		t.Errorf("BlogRoot = %q, want \"/blog/\"", srv.BlogRoot)
	}
	// The generator has no blog root of its own: the server forwards its
	// resolved value, which is what callers such as the CLI rely on.
	if srv.generator.BlogRoot != "/blog/" {
		t.Errorf("generator BlogRoot = %q, want \"/blog/\"", srv.generator.BlogRoot)
	}
	if !srv.AssetsDir.Enabled() {
		t.Error("AssetsDir is not enabled, want the supplied filesystem")
	}
	if srv.TemplateDir.FS == templates.Default {
		t.Error("TemplateDir.FS is the built-in filesystem, want the supplied one")
	}
}

// TestNew_HealthChecksOption verifies that the health-check option is applied
// before the deferred-initialisation branch is taken.
func TestNew_HealthChecksOption(t *testing.T) {
	t.Parallel()

	srv, err := New(makeInternalTestFS(t), config.WithHealthChecks())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if !srv.HealthChecks.Enabled {
		t.Error("HealthChecks.Enabled = false, want true")
	}
	if srv.generator != nil {
		t.Error("generator was built during New, want initialisation deferred to Run")
	}
}

// TestNew_GeneratorOptionForwarded verifies that a generator option passed
// through AsServerOption reaches the generator the server builds.
func TestNew_GeneratorOptionForwarded(t *testing.T) {
	t.Parallel()

	srv, err := New(makeInternalTestFS(t),
		config.WithSiteTitle("Forwarded Title").AsServerOption(),
		config.WithRawOutput().AsServerOption(),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if srv.generator.SiteTitle.SiteTitle != "Forwarded Title" {
		t.Errorf("generator SiteTitle = %q, want \"Forwarded Title\"", srv.generator.SiteTitle.SiteTitle)
	}
	if !srv.generator.RawOutput.RawOutput {
		t.Error("generator RawOutput = false, want the forwarded option to apply")
	}
}

// TestNew_BlogRootForwardedToGenerator verifies that the blog root resolved
// from the server options reaches the generator, so that the in-page links and
// image paths it emits are prefixed with the blog root rather than "/".
//
// The CLI relies on this forwarding: `goblog serve --root-path` passes the
// blog root as a server option only. Without it the routes would still be
// mounted under the blog root while every generated link pointed at "/".
func TestNew_BlogRootForwardedToGenerator(t *testing.T) {
	t.Parallel()

	posts := fstest.MapFS{
		"test-post.md": {Data: []byte(strings.TrimSpace(`
---
title: Test Post
description: A test post
date: 2024-01-01
---

![diagram](diagram.png)
`))},
	}

	// Emit both the blog root passed to the template and the rendered post
	// body, which is where the parser rewrites relative image paths.
	templateFS := fstest.MapFS{
		"pages/index.tmpl":      {Data: []byte(`index`)},
		"pages/post.tmpl":       {Data: []byte(`{{ .BlogRoot }}|{{ .Post.HTMLContent }}`)},
		"pages/tag.tmpl":        {Data: []byte(`tag`)},
		"pages/tags-index.tmpl": {Data: []byte(`tags`)},
	}

	srv, err := New(posts,
		config.WithBlogRoot("/blog/").AsServerOption(),
		config.WithTemplateDir(templateFS),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if srv.generator.BlogRoot != "/blog/" {
		t.Errorf("generator BlogRoot = %q, want \"/blog/\"", srv.generator.BlogRoot)
	}

	w := httptest.NewRecorder()
	srv.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/blog/posts/test-post", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("GET /blog/posts/test-post status = %d, want 200", w.Code)
	}

	body := w.Body.String()
	if !strings.HasPrefix(body, "/blog/|") {
		t.Errorf("rendered BlogRoot = %q, want the page data to carry \"/blog/\"", body)
	}
	if !strings.Contains(body, `src="/blog/images/diagram.png"`) {
		t.Errorf("rendered body = %q, want the image rewritten under \"/blog/images/\"", body)
	}
}

// TestNew_TemplateDirAndRendererOptionForwarded verifies that a custom template
// filesystem is used and that renderer options passed through AsServerOption
// register their functions with the renderer the server builds.
func TestNew_TemplateDirAndRendererOptionForwarded(t *testing.T) {
	t.Parallel()

	srv, err := New(makeInternalTestFS(t),
		config.WithTemplateDir(pageTemplateFS(`{{ shout .SiteTitle }}`)),
		config.WithFuncs(template.FuncMap{"shout": strings.ToUpper}).AsServerOption(),
		config.WithSiteTitle("custom title").AsServerOption(),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	w := httptest.NewRecorder()
	srv.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("GET / status = %d, want 200", w.Code)
	}
	if got := strings.TrimSpace(w.Body.String()); got != "CUSTOM TITLE" {
		t.Errorf("GET / body = %q, want \"CUSTOM TITLE\"", got)
	}
}

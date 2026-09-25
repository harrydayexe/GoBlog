// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package server_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/harrydayexe/GoBlog/v2/pkg/config"
	"github.com/harrydayexe/GoBlog/v2/pkg/generator"
	"github.com/harrydayexe/GoBlog/v2/pkg/server"
)

// seriesBlog returns a GeneratedBlog carrying one series page and a series
// index, as the generator produces when a series file is configured.
func seriesBlog() *generator.GeneratedBlog {
	blog := generator.NewEmptyGeneratedBlog()
	blog.Index = []byte("<html>index</html>")
	blog.Series["building-a-blog-in-go"] = []byte("<html>series</html>")
	blog.SeriesIndex = []byte("<html>series index</html>")
	return blog
}

// TestHandler_SeriesRoutes asserts both clean and .html series URLs resolve
// (case 26), and that an unknown series is a 404 (case 28).
func TestHandler_SeriesRoutes(t *testing.T) {
	t.Parallel()

	h := server.Handler(seriesBlog())

	tests := []struct {
		path     string
		wantCode int
		wantBody string
	}{
		{path: "/series", wantCode: http.StatusOK, wantBody: "series index"},
		{path: "/series.html", wantCode: http.StatusOK, wantBody: "series index"},
		{path: "/series/building-a-blog-in-go", wantCode: http.StatusOK, wantBody: "<html>series</html>"},
		{path: "/series/building-a-blog-in-go.html", wantCode: http.StatusOK, wantBody: "<html>series</html>"},
		{path: "/series/unknown", wantCode: http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			t.Parallel()

			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tt.path, nil))

			if rec.Code != tt.wantCode {
				t.Fatalf("GET %s status = %d, want %d", tt.path, rec.Code, tt.wantCode)
			}
			if tt.wantBody != "" && !strings.Contains(rec.Body.String(), tt.wantBody) {
				t.Errorf("GET %s body = %q, want it to contain %q", tt.path, rec.Body.String(), tt.wantBody)
			}
		})
	}
}

// TestHandler_SeriesRoutesUnregisteredWhenDisabled asserts /series is a 404 when
// the generator produced no series content (case 27).
func TestHandler_SeriesRoutesUnregisteredWhenDisabled(t *testing.T) {
	t.Parallel()

	blog := generator.NewEmptyGeneratedBlog()
	blog.Index = []byte("<html>index</html>")
	h := server.Handler(blog)

	for _, path := range []string{"/series", "/series/building-a-blog-in-go"} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))

		if rec.Code != http.StatusNotFound {
			t.Errorf("GET %s status = %d, want %d when series are disabled", path, rec.Code, http.StatusNotFound)
		}
	}
}

// TestHandler_SeriesRoutesUnderBlogRoot asserts the series routes are prefixed
// with the blog root (case 24).
func TestHandler_SeriesRoutesUnderBlogRoot(t *testing.T) {
	t.Parallel()

	h := server.Handler(seriesBlog(), config.WithBlogRoot("/blog/"))

	for _, path := range []string{"/blog/series", "/blog/series/building-a-blog-in-go"} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))

		if rec.Code != http.StatusOK {
			t.Errorf("GET %s status = %d, want %d", path, rec.Code, http.StatusOK)
		}
	}
}

// ---- live reload ------------------------------------------------------------

// seriesSiteDir writes three posts and a series file to a temporary directory
// and returns its path.
func seriesSiteDir(t *testing.T, seriesYAML string) string {
	t.Helper()

	dir := t.TempDir()
	for _, part := range []struct{ file, title, date string }{
		{"part-1.md", "Part One", "2024-01-01"},
		{"part-2.md", "Part Two", "2024-02-01"},
		{"part-3.md", "Part Three", "2024-03-01"},
	} {
		body := "---\ntitle: " + part.title + "\ndate: " + part.date +
			"\ndescription: About " + part.title + "\n---\n\n# " + part.title + "\n"
		if err := os.WriteFile(filepath.Join(dir, part.file), []byte(body), 0o644); err != nil {
			t.Fatalf("write %s: %v", part.file, err)
		}
	}

	writeSeriesFile(t, dir, seriesYAML)
	return dir
}

// writeSeriesFile writes (or overwrites) series.yml in dir.
func writeSeriesFile(t *testing.T, dir, seriesYAML string) {
	t.Helper()

	if err := os.WriteFile(filepath.Join(dir, "series.yml"), []byte(seriesYAML), 0o644); err != nil {
		t.Fatalf("write series.yml: %v", err)
	}
}

// seriesServer builds a server that reads its posts and its series file from
// dir, the way goblog serve --series-file does.
func seriesServer(t *testing.T, dir string) *server.Server {
	t.Helper()

	srv, err := server.New(os.DirFS(dir),
		config.WithPort(0),
		config.WithSeriesFile(os.DirFS(dir), "series.yml").AsServerOption(),
	)
	if err != nil {
		t.Fatalf("server.New() error = %v", err)
	}
	return srv
}

// seriesPageBody requests the series page and returns its body, failing the test
// unless the response is 200.
func seriesPageBody(t *testing.T, srv *server.Server) string {
	t.Helper()

	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/series/parts", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /series/parts status = %d, want %d", rec.Code, http.StatusOK)
	}
	return rec.Body.String()
}

const forwardSeries = "series:\n  - name: Parts\n    posts:\n      - part-1.md\n      - part-2.md\n      - part-3.md\n"

// TestSeries_ReloadReflectsReorderedFile asserts that editing the series file
// and reloading changes the order the next request sees (case 30).
func TestSeries_ReloadReflectsReorderedFile(t *testing.T) {
	t.Parallel()

	dir := seriesSiteDir(t, forwardSeries)
	srv := seriesServer(t, dir)

	before := seriesPageBody(t, srv)
	if strings.Index(before, "Part One") > strings.Index(before, "Part Three") {
		t.Fatal("series page is not in file order before the edit")
	}

	writeSeriesFile(t, dir, "series:\n  - name: Parts\n    posts:\n      - part-3.md\n      - part-2.md\n      - part-1.md\n")

	if err := srv.UpdatePosts(os.DirFS(dir), context.Background()); err != nil {
		t.Fatalf("UpdatePosts() error = %v", err)
	}

	after := seriesPageBody(t, srv)
	if strings.Index(after, "Part Three") > strings.Index(after, "Part One") {
		t.Error("series page does not reflect the reordered series file")
	}
}

// TestSeries_InvalidReloadKeepsLastGoodSite asserts that making the series file
// invalid surfaces an error and leaves the previous site being served (case 31).
func TestSeries_InvalidReloadKeepsLastGoodSite(t *testing.T) {
	t.Parallel()

	dir := seriesSiteDir(t, forwardSeries)
	srv := seriesServer(t, dir)

	before := seriesPageBody(t, srv)

	writeSeriesFile(t, dir, "series:\n  - name: Parts\n    posts:\n      - gone.md\n")

	if err := srv.UpdatePosts(os.DirFS(dir), context.Background()); err == nil {
		t.Fatal("UpdatePosts() error = nil, want an error for the invalid series file")
	}

	if after := seriesPageBody(t, srv); after != before {
		t.Error("series page changed after a failed reload; the last good site should still be served")
	}
}

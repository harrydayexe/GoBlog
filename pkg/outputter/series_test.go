// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package outputter

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/harrydayexe/GoBlog/v2/pkg/config"
	"github.com/harrydayexe/GoBlog/v2/pkg/generator"
)

// blogWithSeries returns a GeneratedBlog carrying one series page and a series
// index, as the generator produces when a series file is configured.
func blogWithSeries() *generator.GeneratedBlog {
	blog := generator.NewEmptyGeneratedBlog()
	blog.Index = []byte("<html>index</html>")
	blog.Series["building-a-blog-in-go"] = []byte("<html>series</html>")
	blog.SeriesIndex = []byte("<html>series index</html>")
	return blog
}

// TestHandleGeneratedBlog_WritesSeries asserts the series directory holds one
// file per series plus its index (case 21).
func TestHandleGeneratedBlog_WritesSeries(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dw := NewDirectoryWriter(dir, config.WithSeriesFile(fstest.MapFS{}, "series.yml"))

	if err := dw.HandleGeneratedBlog(context.Background(), blogWithSeries()); err != nil {
		t.Fatalf("HandleGeneratedBlog() error = %v", err)
	}

	for _, file := range []string{"series/index.html", "series/building-a-blog-in-go.html"} {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(file))); err != nil {
			t.Errorf("%s was not written: %v", file, err)
		}
	}
}

// TestHandleGeneratedBlog_NoSeriesDirectory asserts no series directory appears
// when the generator produced no series content (case 22), and that raw output
// never writes one even if series content is present (case 23).
func TestHandleGeneratedBlog_NoSeriesDirectory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		blog *generator.GeneratedBlog
		opts []config.GeneratorOption
	}{
		{
			name: "series disabled",
			blog: func() *generator.GeneratedBlog {
				blog := generator.NewEmptyGeneratedBlog()
				blog.Index = []byte("<html>index</html>")
				return blog
			}(),
		},
		{
			name: "raw output",
			blog: blogWithSeries(),
			opts: []config.GeneratorOption{config.WithRawOutput()},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			dw := NewDirectoryWriter(dir, tt.opts...)

			if err := dw.HandleGeneratedBlog(context.Background(), tt.blog); err != nil {
				t.Fatalf("HandleGeneratedBlog() error = %v", err)
			}

			if _, err := os.Stat(filepath.Join(dir, "series")); !os.IsNotExist(err) {
				t.Errorf("series directory exists (stat error = %v), want it absent", err)
			}
		})
	}
}

// TestHandleGeneratedBlog_SeriesWithDisableTags asserts series pages are written
// with tags disabled: the two features are independent (case 29).
func TestHandleGeneratedBlog_SeriesWithDisableTags(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dw := NewDirectoryWriter(dir, config.WithDisableTags())

	if err := dw.HandleGeneratedBlog(context.Background(), blogWithSeries()); err != nil {
		t.Fatalf("HandleGeneratedBlog() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "series", "index.html")); err != nil {
		t.Errorf("series/index.html was not written with --disable-tags: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "tags")); !os.IsNotExist(err) {
		t.Errorf("tags directory exists (stat error = %v), want it absent", err)
	}
}

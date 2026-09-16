// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package outputter

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/harrydayexe/GoBlog/v2/pkg/config"
	"github.com/harrydayexe/GoBlog/v2/pkg/generator"
)

func TestHandleGeneratedBlog_CopiesAssets(t *testing.T) {
	t.Parallel()

	assets := fstest.MapFS{
		"pipeline.png":      {Data: []byte("png")},
		"screenshots/a.png": {Data: []byte("a")},
	}

	tests := []struct {
		name string
		raw  bool
	}{
		{name: "templated", raw: false},
		{name: "raw", raw: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			outputDir := t.TempDir()
			opts := []config.GeneratorOption{
				config.WithLogger(slog.New(slog.DiscardHandler)).AsGeneratorOption(),
				config.WithAssetsDir(assets).AsGeneratorOption(),
			}
			if tt.raw {
				opts = append(opts, config.WithRawOutput())
			}
			dw := NewDirectoryWriter(outputDir, opts...)

			// Run twice to confirm existing files are overwritten without error.
			for range 2 {
				if err := dw.HandleGeneratedBlog(context.Background(), generator.NewEmptyGeneratedBlog()); err != nil {
					t.Fatalf("HandleGeneratedBlog() error = %v", err)
				}
			}

			for name, file := range assets {
				got, err := os.ReadFile(filepath.Join(outputDir, "images", filepath.FromSlash(name)))
				if err != nil {
					t.Errorf("expected asset %q to be copied: %v", name, err)
					continue
				}
				if string(got) != string(file.Data) {
					t.Errorf("asset %q: got %q, want %q", name, got, file.Data)
				}
			}
		})
	}
}

func TestHandleGeneratedBlog_NoAssets(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		opts []config.GeneratorOption
	}{
		{name: "option not supplied"},
		{name: "missing directory", opts: []config.GeneratorOption{
			config.WithAssetsDir(os.DirFS(filepath.Join(t.TempDir(), "nope"))).AsGeneratorOption(),
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			outputDir := t.TempDir()
			dw := NewDirectoryWriter(outputDir, tt.opts...)
			if err := dw.HandleGeneratedBlog(context.Background(), generator.NewEmptyGeneratedBlog()); err != nil {
				t.Fatalf("HandleGeneratedBlog() error = %v", err)
			}
			if _, err := os.Stat(filepath.Join(outputDir, "images")); !os.IsNotExist(err) {
				t.Errorf("expected no images directory, stat err = %v", err)
			}
		})
	}
}

func TestHandleGeneratedBlog_AssetSymlinkEscapeSkipped(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	assetsDir := filepath.Join(dir, "images")
	if err := os.Mkdir(assetsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "secret.txt"), []byte("secret"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(assetsDir, "ok.png"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(dir, "secret.txt"), filepath.Join(assetsDir, "link.png")); err != nil {
		t.Fatal(err)
	}
	root, err := os.OpenRoot(assetsDir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { root.Close() })

	outputDir := t.TempDir()
	dw := NewDirectoryWriter(outputDir,
		config.WithLogger(slog.New(slog.DiscardHandler)).AsGeneratorOption(),
		config.WithAssetsDir(root.FS()).AsGeneratorOption(),
	)
	if err := dw.HandleGeneratedBlog(context.Background(), generator.NewEmptyGeneratedBlog()); err != nil {
		t.Fatalf("HandleGeneratedBlog() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(outputDir, "images", "ok.png")); err != nil {
		t.Errorf("expected ok.png to be copied: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(outputDir, "images", "link.png")); !os.IsNotExist(err) {
		t.Errorf("expected escaping symlink to be skipped, lstat err = %v", err)
	}
}

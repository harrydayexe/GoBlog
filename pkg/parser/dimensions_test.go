// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package parser

import (
	"bytes"
	"context"
	"image"
	"image/jpeg"
	"image/png"
	"io/fs"
	"strings"
	"sync"
	"testing"
	"testing/fstest"
)

// encodePNG returns a real PNG file of the given size, so image.DecodeConfig
// has a genuine header to read.
func encodePNG(t *testing.T, width, height int) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, width, height))); err != nil {
		t.Fatalf("failed to encode png: %v", err)
	}
	return buf.Bytes()
}

// encodeJPEG returns a real JPEG file of the given size.
func encodeJPEG(t *testing.T, width, height int) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, image.NewRGBA(image.Rect(0, 0, width, height)), nil); err != nil {
		t.Fatalf("failed to encode jpeg: %v", err)
	}
	return buf.Bytes()
}

// assetsFS returns an assets filesystem holding measurable images alongside
// files that cannot be measured.
func assetsFS(t *testing.T) fstest.MapFS {
	t.Helper()
	return fstest.MapFS{
		"cat.png":              {Data: encodePNG(t, 1200, 800)},
		"dog.png":              {Data: encodePNG(t, 640, 480)},
		"screenshots/a.png":    {Data: encodePNG(t, 300, 150)},
		"photo.jpg":            {Data: encodeJPEG(t, 400, 250)},
		"logo.svg":             {Data: []byte(`<svg width="10" height="10"></svg>`)},
		"broken.png":           {Data: []byte("not really a png")},
		"screenshots/notes.md": {Data: []byte("not an image")},
	}
}

func TestParseFile_ImageDimensions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
		want string
	}{
		{
			name: "markdown image is measured",
			body: "![A cat](cat.png)\n",
			want: `<img src="/images/cat.png" alt="A cat" width="1200" height="800" loading="lazy" decoding="async" />`,
		},
		{
			name: "markdown image with images prefix is measured",
			body: "![A cat](images/cat.png)\n",
			want: `<img src="/images/cat.png" alt="A cat" width="1200" height="800" loading="lazy" decoding="async" />`,
		},
		{
			name: "bare embed has empty alt text",
			body: "![[dog.png]]\n",
			want: `<img src="/images/dog.png" alt="" width="640" height="480" loading="lazy" decoding="async" />`,
		},
		{
			name: "labelled embed keeps its alt text",
			body: "![[dog.png|A dog]]\n",
			want: `<img src="/images/dog.png" alt="A dog" width="640" height="480" loading="lazy" decoding="async" />`,
		},
		{
			name: "subdirectory path is measured",
			body: "![shot](images/screenshots/a.png)\n",
			want: `<img src="/images/screenshots/a.png" alt="shot" width="300" height="150" loading="lazy" decoding="async" />`,
		},
		{
			name: "subdirectory embed is measured",
			body: "![[screenshots/a.png]]\n",
			want: `<img src="/images/screenshots/a.png" alt="" width="300" height="150" loading="lazy" decoding="async" />`,
		},
		{
			name: "jpeg is measured",
			body: "![A photo](photo.jpg)\n",
			want: `<img src="/images/photo.jpg" alt="A photo" width="400" height="250" loading="lazy" decoding="async" />`,
		},
		{
			name: "query string does not prevent measuring",
			body: "![A cat](cat.png?v=2)\n",
			want: `<img src="/images/cat.png?v=2" alt="A cat" width="1200" height="800" loading="lazy" decoding="async" />`,
		},
		{
			name: "missing file has no dimensions",
			body: "![Missing](nope.png)\n",
			want: `<img src="/images/nope.png" alt="Missing" loading="lazy" decoding="async" />`,
		},
		{
			name: "missing file embed has no dimensions",
			body: "![[nope.png]]\n",
			want: `<img src="/images/nope.png" alt="" loading="lazy" decoding="async" />`,
		},
		{
			name: "undecodable format has no dimensions",
			body: "![A logo](logo.svg)\n",
			want: `<img src="/images/logo.svg" alt="A logo" loading="lazy" decoding="async" />`,
		},
		{
			name: "corrupt file has no dimensions",
			body: "![Broken](broken.png)\n",
			want: `<img src="/images/broken.png" alt="Broken" loading="lazy" decoding="async" />`,
		},
		{
			name: "absolute URL keeps loading hints only",
			body: "![Remote](https://example.com/cat.png)\n",
			want: `<img src="https://example.com/cat.png" alt="Remote" loading="lazy" decoding="async" />`,
		},
		{
			name: "root relative path keeps loading hints only",
			body: "![Static](/cat.png)\n",
			want: `<img src="/cat.png" alt="Static" loading="lazy" decoding="async" />`,
		},
		{
			name: "protocol relative URL keeps loading hints only",
			body: "![Remote](//cdn.example.com/cat.png)\n",
			want: `<img src="//cdn.example.com/cat.png" alt="Remote" loading="lazy" decoding="async" />`,
		},
		{
			name: "data URI keeps loading hints only",
			body: "![Inline](data:image/png;base64,AAAA)\n",
			want: `<img src="data:image/png;base64,AAAA" alt="Inline" loading="lazy" decoding="async" />`,
		},
		{
			name: "dot dot path is never read from the assets directory",
			body: "![Escape](../cat.png)\n",
			want: `<img src="../cat.png" alt="Escape" loading="lazy" decoding="async" />`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := parseBodyWithOptions(t, tt.body, WithAssetsDir(assetsFS(t)))
			if !strings.Contains(got, tt.want) {
				t.Errorf("expected output to contain %q, got:\n%s", tt.want, got)
			}
		})
	}
}

func TestParseFile_ImageDimensionsWithoutAssetsDir(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
		want string
	}{
		{
			name: "markdown image",
			body: "![A cat](cat.png)\n",
			want: `<img src="/images/cat.png" alt="A cat" loading="lazy" decoding="async" />`,
		},
		{
			name: "bare embed",
			body: "![[dog.png]]\n",
			want: `<img src="/images/dog.png" alt="" loading="lazy" decoding="async" />`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := parseBodyWithOptions(t, tt.body)
			if !strings.Contains(got, tt.want) {
				t.Errorf("expected output to contain %q, got:\n%s", tt.want, got)
			}
		})
	}
}

// TestParseFile_ImageDimensionsUnderBlogRoot verifies that the assets path a
// measurement reads is derived from the asset path, not the served URL, so a
// blog root does not change which file is measured.
func TestParseFile_ImageDimensionsUnderBlogRoot(t *testing.T) {
	t.Parallel()

	got := parseBodyWithOptions(t, "![A cat](cat.png)\n", WithBlogRoot("/blog/"), WithAssetsDir(assetsFS(t)))
	want := `<img src="/blog/images/cat.png" alt="A cat" width="1200" height="800" loading="lazy" decoding="async" />`
	if !strings.Contains(got, want) {
		t.Errorf("expected output to contain %q, got:\n%s", want, got)
	}
}

// countingFS counts how many times a file is opened.
type countingFS struct {
	fs.FS

	mu    sync.Mutex
	opens map[string]int
}

func (c *countingFS) Open(name string) (fs.File, error) {
	c.mu.Lock()
	if c.opens == nil {
		c.opens = make(map[string]int)
	}
	c.opens[name]++
	c.mu.Unlock()
	return c.FS.Open(name)
}

// TestParseDirectory_MeasurementsAreCached verifies that an image referenced
// from several posts is only read once.
func TestParseDirectory_MeasurementsAreCached(t *testing.T) {
	t.Parallel()

	assets := &countingFS{FS: assetsFS(t)}
	posts := fstest.MapFS{
		"a.md": {Data: []byte(wikilinkFrontmatter + "![A cat](cat.png)\n\n![[cat.png]]\n")},
		"b.md": {Data: []byte(wikilinkFrontmatter + "![A cat](images/cat.png)\n")},
	}

	p := New(WithCodeHighlighting(false), WithAssetsDir(assets))
	if _, err := p.ParseDirectory(context.Background(), posts); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if n := assets.opens["cat.png"]; n != 1 {
		t.Errorf("cat.png was opened %d times, want 1", n)
	}
}

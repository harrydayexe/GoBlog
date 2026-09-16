// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package parser

import (
	"context"
	"strings"
	"testing"
	"testing/fstest"
)

func parseBodyWithRoot(t *testing.T, blogRoot, body string) string {
	t.Helper()
	fsys := fstest.MapFS{
		"post.md": {Data: []byte(wikilinkFrontmatter + body)},
	}
	post, err := New(WithCodeHighlighting(false), WithBlogRoot(blogRoot)).ParseFile(context.Background(), fsys, "post.md")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	return string(post.Content)
}

func TestAssetURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		dest   string
		want   string
		wantOK bool
	}{
		{name: "images prefix", dest: "images/foo.png", want: "/images/foo.png", wantOK: true},
		{name: "bare filename", dest: "foo.png", want: "/images/foo.png", wantOK: true},
		{name: "subdirectory", dest: "images/screenshots/a.png", want: "/images/screenshots/a.png", wantOK: true},
		{name: "subdirectory without prefix", dest: "screenshots/a.png", want: "/images/screenshots/a.png", wantOK: true},
		{name: "dot slash", dest: "./images/foo.png", want: "/images/foo.png", wantOK: true},
		{name: "query and fragment kept", dest: "foo.png?v=2#x", want: "/images/foo.png?v=2#x", wantOK: true},
		{name: "root relative", dest: "/static/foo.png"},
		{name: "protocol relative", dest: "//cdn.example.com/foo.png"},
		{name: "https", dest: "https://example.com/foo.png"},
		{name: "data uri", dest: "data:image/png;base64,AAAA"},
		{name: "dot dot", dest: "../secret.png"},
		{name: "dot dot inside", dest: "images/../../secret.png"},
		{name: "empty", dest: ""},
		{name: "images dir only", dest: "images/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, ok := assetURL("/", tt.dest)
			if ok != tt.wantOK || got != tt.want {
				t.Errorf("assetURL(%q) = (%q, %v), want (%q, %v)", tt.dest, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}

func TestParseFile_ImageRewriting(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
		want map[string]string // blog root -> expected substring
	}{
		{
			name: "markdown images prefix",
			body: "![A diagram](images/pipeline.png)\n",
			want: map[string]string{
				"/":      `<img src="/images/pipeline.png" alt="A diagram"`,
				"/blog/": `<img src="/blog/images/pipeline.png" alt="A diagram"`,
			},
		},
		{
			name: "markdown bare filename",
			body: "![A diagram](pipeline.png)\n",
			want: map[string]string{
				"/":      `<img src="/images/pipeline.png" alt="A diagram"`,
				"/blog/": `<img src="/blog/images/pipeline.png" alt="A diagram"`,
			},
		},
		{
			name: "markdown subdirectory",
			body: "![shot](images/screenshots/a.png)\n",
			want: map[string]string{
				"/":      `<img src="/images/screenshots/a.png"`,
				"/blog/": `<img src="/blog/images/screenshots/a.png"`,
			},
		},
		{
			name: "markdown root relative untouched",
			body: "![x](/static/x.png)\n",
			want: map[string]string{
				"/":      `<img src="/static/x.png"`,
				"/blog/": `<img src="/static/x.png"`,
			},
		},
		{
			name: "markdown absolute URL untouched",
			body: "![x](https://example.com/x.png)\n",
			want: map[string]string{
				"/":      `<img src="https://example.com/x.png"`,
				"/blog/": `<img src="https://example.com/x.png"`,
			},
		},
		{
			name: "markdown dot dot untouched",
			body: "![x](../x.png)\n",
			want: map[string]string{
				"/":      `<img src="../x.png"`,
				"/blog/": `<img src="../x.png"`,
			},
		},
		{
			name: "markdown empty untouched",
			body: "![x]()\n",
			want: map[string]string{
				"/":      `<img src="" alt="x"`,
				"/blog/": `<img src="" alt="x"`,
			},
		},
		{
			name: "wikilink embed",
			body: "![[pipeline.png]]\n",
			want: map[string]string{
				"/":      `<img src="/images/pipeline.png">`,
				"/blog/": `<img src="/blog/images/pipeline.png">`,
			},
		},
		{
			name: "wikilink embed with alt text",
			body: "![[pipeline.png|A diagram]]\n",
			want: map[string]string{
				"/":      `<img src="/images/pipeline.png" alt="A diagram">`,
				"/blog/": `<img src="/blog/images/pipeline.png" alt="A diagram">`,
			},
		},
		{
			name: "wikilink embed subdirectory",
			body: "![[images/screenshots/a.png]]\n",
			want: map[string]string{
				"/":      `<img src="/images/screenshots/a.png">`,
				"/blog/": `<img src="/blog/images/screenshots/a.png">`,
			},
		},
		{
			name: "wikilink embed dot dot untouched",
			body: "![[../x.png]]\n",
			want: map[string]string{
				"/":      `<img src="../x.png">`,
				"/blog/": `<img src="../x.png">`,
			},
		},
		{
			name: "quotes are escaped in src",
			body: "![x](<a\"b.png>)\n",
			want: map[string]string{
				"/": `<img src="/images/a%22b.png"`,
			},
		},
	}

	for _, tt := range tests {
		for root, want := range tt.want {
			t.Run(tt.name+" root "+root, func(t *testing.T) {
				t.Parallel()
				got := parseBodyWithRoot(t, root, tt.body)
				if !strings.Contains(got, want) {
					t.Errorf("expected output to contain %q, got:\n%s", want, got)
				}
			})
		}
	}
}

func TestParseFile_NonImageEmbedIsNotAnImage(t *testing.T) {
	t.Parallel()

	got := parseBodyWithRoot(t, "/", "![[notes.txt]]\n")
	if strings.Contains(got, "<img") || strings.Contains(got, "<a") {
		t.Errorf("expected non-image embed to render as text, got:\n%s", got)
	}
	if !strings.Contains(got, "notes.txt") {
		t.Errorf("expected label text in output, got:\n%s", got)
	}
}

func TestParseFile_HeadingAnchorsUnaffectedByBlogRoot(t *testing.T) {
	t.Parallel()

	got := parseBodyWithRoot(t, "/blog/", "See [[#Future Work]].\n\n## Future Work\n")
	if want := `<a href="#future-work">Future Work</a>`; !strings.Contains(got, want) {
		t.Errorf("expected output to contain %q, got:\n%s", want, got)
	}
}

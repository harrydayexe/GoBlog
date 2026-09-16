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

const wikilinkFrontmatter = `---
title: "Wikilinks"
date: 2026-01-10T10:00:00Z
description: "Heading anchor tests"
---

`

func parseBody(t *testing.T, body string) (string, error) {
	t.Helper()
	fsys := fstest.MapFS{
		"post.md": {Data: []byte(wikilinkFrontmatter + body)},
	}
	post, err := New(WithCodeHighlighting(false)).ParseFile(context.Background(), fsys, "post.md")
	if err != nil {
		return "", err
	}
	return string(post.Content), nil
}

func TestParseFile_HeadingIDs(t *testing.T) {
	t.Parallel()

	got, err := parseBody(t, "## Future Work\n\nSee [below](#future-work).\n")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	for _, want := range []string{
		`<h2 id="future-work">Future Work</h2>`,
		`<a href="#future-work">below</a>`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("expected output to contain %q, got:\n%s", want, got)
		}
	}
}

func TestParseFile_WikilinkHeadingAnchors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
		want []string
	}{
		{
			name: "heading text as label",
			body: "See [[#Future Work]] for details.\n\n## Future Work\n",
			want: []string{
				`<p>See <a href="#future-work">Future Work</a> for details.</p>`,
				`<h2 id="future-work">Future Work</h2>`,
			},
		},
		{
			name: "custom label",
			body: "See [[#Future Work|what comes next]].\n\n## Future Work\n",
			want: []string{`<a href="#future-work">what comes next</a>`},
		},
		{
			name: "link before and after heading",
			body: "[[#Intro]]\n\n# Intro\n\nBack to [[#Intro|top]].\n",
			want: []string{
				`<a href="#intro">Intro</a>`,
				`<a href="#intro">top</a>`,
			},
		},
		{
			name: "punctuation and case are slugified like heading ids",
			body: "## Hello, World!\n\n[[#Hello, World!]]\n",
			want: []string{
				`<h2 id="hello-world">Hello, World!</h2>`,
				`<a href="#hello-world">Hello, World!</a>`,
			},
		},
		{
			name: "label is html escaped",
			body: "## Setup\n\n[[#Setup|<script>alert(1)</script>]]\n",
			want: []string{`<a href="#setup">&lt;script&gt;alert(1)&lt;/script&gt;</a>`},
		},
		{
			name: "heading containing a hash",
			body: "## C# Tips\n\n[[#C# Tips]]\n",
			want: []string{
				`<h2 id="c-tips">C# Tips</h2>`,
				`<a href="#c-tips">C# Tips</a>`,
			},
		},
		{
			name: "label containing a hash is kept",
			body: "## Setup\n\n[[#Setup|#1 step]]\n",
			want: []string{`<a href="#setup">#1 step</a>`},
		},
		{
			name: "regular links are unaffected",
			body: "[text](https://example.com) and [anchor](#setup)\n\n## Setup\n",
			want: []string{
				`<a href="https://example.com">text</a>`,
				`<a href="#setup">anchor</a>`,
			},
		},
		{
			name: "cross-post wikilinks render as plain text",
			body: "See [[other-post]] and [[other-post#Heading|that heading]].\n",
			want: []string{`<p>See other-post and that heading.</p>`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseBody(t, tt.body)
			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
			for _, want := range tt.want {
				if !strings.Contains(got, want) {
					t.Errorf("expected output to contain %q, got:\n%s", want, got)
				}
			}
		})
	}
}

func TestParseFile_WikilinkUnresolvedRendersSilently(t *testing.T) {
	t.Parallel()

	// Mirrors standard [text](#anchor) links, which are not validated.
	got, err := parseBody(t, "See [[#Missing]] and [regular](#also-missing).\n\n## Present\n")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	for _, want := range []string{
		`<a href="#missing">Missing</a>`,
		`<a href="#also-missing">regular</a>`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("expected output to contain %q, got:\n%s", want, got)
		}
	}
}

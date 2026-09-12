// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package parser

import (
	"context"
	"errors"
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
			name: "regular links are unaffected",
			body: "[text](https://example.com) and [[not a heading link]]\n",
			want: []string{
				`<a href="https://example.com">text</a>`,
				`[[not a heading link]]`,
			},
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

func TestParseFile_WikilinkUnresolved(t *testing.T) {
	t.Parallel()

	_, err := parseBody(t, "See [[#Missing]] and [[#Also Missing]].\n\n## Present\n\n[[#Present]]\n")
	if err == nil {
		t.Fatal("expected error for unresolved heading link, got nil")
	}
	if !errors.Is(err, ErrUnresolvedHeadingLink) {
		t.Errorf("expected error to wrap ErrUnresolvedHeadingLink, got: %v", err)
	}
	for _, want := range []string{"[[#Missing]]", "[[#Also Missing]]"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("expected error to mention %q, got: %v", want, err)
		}
	}
	if strings.Contains(err.Error(), "[[#Present]]") {
		t.Errorf("expected resolved link not to be reported, got: %v", err)
	}
}

func TestParseDirectory_WikilinkUnresolved(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{
		"good.md": {Data: []byte(wikilinkFrontmatter + "## A\n\n[[#A]]\n")},
		"bad.md":  {Data: []byte(wikilinkFrontmatter + "[[#Nope]]\n")},
	}

	posts, err := New().ParseDirectory(context.Background(), fsys)
	if len(posts) != 1 {
		t.Errorf("expected 1 valid post, got: %d", len(posts))
	}

	var parseErrs ParseErrors
	if !errors.As(err, &parseErrs) {
		t.Fatalf("expected ParseErrors, got: %T (%v)", err, err)
	}
	if len(parseErrs.Errors) != 1 || parseErrs.Errors[0].Path != "bad.md" {
		t.Fatalf("expected a single error for bad.md, got: %v", parseErrs)
	}
	if !errors.Is(parseErrs.Errors[0], ErrUnresolvedHeadingLink) {
		t.Errorf("expected error to wrap ErrUnresolvedHeadingLink, got: %v", parseErrs.Errors[0])
	}
}

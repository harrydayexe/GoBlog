// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package parser

import (
	"log/slog"
	"net/url"
	"path"
	"slices"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// assetsURLDir is the fixed URL directory, under the blog root, that assets
// are served from.
const assetsURLDir = "images/"

// imageTransformer rewrites the destination of standard markdown images,
// ![alt](path), into root-absolute asset URLs. See assetURL for the rules.
type imageTransformer struct {
	blogRoot string
	logger   *slog.Logger
}

func (t imageTransformer) Transform(doc *ast.Document, _ text.Reader, _ parser.Context) {
	_ = ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		img, ok := node.(*ast.Image)
		if !entering || !ok {
			return ast.WalkContinue, nil
		}
		if dest, ok := assetURL(t.blogRoot, string(img.Destination)); ok {
			img.Destination = []byte(dest)
		} else if hasDotDotSegment(string(img.Destination)) {
			t.logger.Warn("image path contains '..' and was not rewritten", slog.String("path", string(img.Destination)))
		}
		return ast.WalkContinue, nil
	})
}

// assetURL rewrites an author-supplied image path into a URL under
// "{blogRoot}images/". It returns false, meaning the path must be left
// untouched, for:
//   - empty paths
//   - root-relative paths ("/foo.png") and protocol-relative URLs ("//host/x")
//   - absolute URLs with any scheme ("https://...", "data:...")
//   - paths containing a ".." segment, so a rewrite can never escape the
//     assets root
//
// Otherwise a leading "images/" is dropped, so "images/foo.png" and "foo.png"
// both map to "{blogRoot}images/foo.png". Subdirectories, query strings and
// fragments are preserved.
func assetURL(blogRoot, dest string) (string, bool) {
	if dest == "" || strings.HasPrefix(dest, "/") || hasDotDotSegment(dest) {
		return "", false
	}
	if u, err := url.Parse(dest); err != nil || u.Scheme != "" {
		return "", false
	}

	p, suffix := dest, ""
	if i := strings.IndexAny(dest, "?#"); i >= 0 {
		p, suffix = dest[:i], dest[i:]
	}
	p = path.Clean(p)
	if p == "." || p+"/" == assetsURLDir {
		return "", false
	}
	p = strings.TrimPrefix(p, assetsURLDir)

	return normaliseBlogRoot(blogRoot) + assetsURLDir + p + suffix, true
}

// hasDotDotSegment reports whether the path portion of dest contains a ".."
// segment.
func hasDotDotSegment(dest string) bool {
	if i := strings.IndexAny(dest, "?#"); i >= 0 {
		dest = dest[:i]
	}
	return slices.Contains(strings.Split(dest, "/"), "..")
}

// normaliseBlogRoot returns root with exactly one leading and trailing slash,
// or "/" when root is empty.
func normaliseBlogRoot(root string) string {
	trimmed := strings.Trim(root, "/")
	if trimmed == "" {
		return "/"
	}
	return "/" + trimmed + "/"
}

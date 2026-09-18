// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package parser

import (
	"bytes"
	"log/slog"
	"net/url"
	"path"
	"slices"
	"strconv"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"go.abhg.dev/goldmark/wikilink"
)

// assetsURLDir is the fixed URL directory, under the blog root, that assets
// are served from.
const assetsURLDir = "images/"

// imageTransformer rewrites the destination of markdown images, ![alt](path),
// into root-absolute asset URLs (see assetRef) and annotates every image with
// the attributes browsers need to lay the page out before it loads: intrinsic
// width and height, loading="lazy" and decoding="async".
//
// Wikilink image embeds, ![[path]], are converted to markdown image nodes
// first so both syntaxes share this one code path.
type imageTransformer struct {
	blogRoot string
	logger   *slog.Logger
	measurer *imageMeasurer
}

func (t imageTransformer) Transform(doc *ast.Document, reader text.Reader, _ parser.Context) {
	replaceImageEmbeds(doc, reader.Source())

	_ = ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		img, ok := node.(*ast.Image)
		if !entering || !ok {
			return ast.WalkContinue, nil
		}
		if ref, ok := resolveAsset(t.blogRoot, string(img.Destination)); ok {
			img.Destination = []byte(ref.URL)
			if dim, ok := t.measurer.measure(ref.Path); ok {
				img.SetAttributeString("width", []byte(strconv.Itoa(dim.width)))
				img.SetAttributeString("height", []byte(strconv.Itoa(dim.height)))
			}
		} else if hasDotDotSegment(string(img.Destination)) {
			t.logger.Warn("image path contains '..' and was not rewritten", slog.String("path", string(img.Destination)))
		}
		img.SetAttributeString("loading", []byte("lazy"))
		img.SetAttributeString("decoding", []byte("async"))
		return ast.WalkContinue, nil
	})
}

// replaceImageEmbeds swaps every wikilink embed of an image, ![[foo.png]], for
// an equivalent markdown image node. The wikilink extension renders its own
// <img> tag and ignores node attributes, so dimensions and loading hints could
// not otherwise be added to embeds.
//
// A bare ![[foo.png]] is left without alt text, rendering as alt="", because
// the wikilink parser uses the target as the label when none is given and a
// file name is not a description.
func replaceImageEmbeds(doc *ast.Document, src []byte) {
	var embeds []*wikilink.Node
	_ = ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		n, ok := node.(*wikilink.Node)
		if entering && ok && n.Embed && isImageTarget(n.Target) {
			embeds = append(embeds, n)
		}
		return ast.WalkContinue, nil
	})

	for _, n := range embeds {
		link := ast.NewLink()
		link.Destination = n.Target
		img := ast.NewImage(link)
		if label, ok := n.FirstChild().(*ast.Text); ok && n.ChildCount() == 1 &&
			!bytes.Equal(label.Segment.Value(src), n.Target) {
			img.AppendChild(img, ast.NewTextSegment(label.Segment))
		}
		n.Parent().ReplaceChild(n.Parent(), n, img)
	}
}

// assetRef is an author-supplied image path resolved against the assets
// directory: the URL it is served from, and the path it lives at inside the
// assets filesystem.
type assetRef struct {
	URL  string // e.g. "/blog/images/screenshots/a.png?v=2"
	Path string // e.g. "screenshots/a.png"
}

// resolveAsset rewrites an author-supplied image path into a URL under
// "{blogRoot}images/" and the matching path within the assets filesystem. It
// returns false, meaning the path must be left untouched, for:
//   - empty paths
//   - root-relative paths ("/foo.png") and protocol-relative URLs ("//host/x")
//   - absolute URLs with any scheme ("https://...", "data:...")
//   - paths containing a ".." segment, so a rewrite can never escape the
//     assets root
//
// Otherwise a leading "images/" is dropped, so "images/foo.png" and "foo.png"
// both map to "{blogRoot}images/foo.png". Subdirectories are preserved.
// Query strings and fragments are kept in the URL and excluded from the
// filesystem path.
func resolveAsset(blogRoot, dest string) (assetRef, bool) {
	if dest == "" || strings.HasPrefix(dest, "/") || hasDotDotSegment(dest) {
		return assetRef{}, false
	}
	if u, err := url.Parse(dest); err != nil || u.Scheme != "" {
		return assetRef{}, false
	}

	p, suffix := dest, ""
	if i := strings.IndexAny(dest, "?#"); i >= 0 {
		p, suffix = dest[:i], dest[i:]
	}
	p = path.Clean(p)
	if p == "." || p+"/" == assetsURLDir {
		return assetRef{}, false
	}
	p = strings.TrimPrefix(p, assetsURLDir)

	return assetRef{
		URL:  normaliseBlogRoot(blogRoot) + assetsURLDir + p + suffix,
		Path: p,
	}, true
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

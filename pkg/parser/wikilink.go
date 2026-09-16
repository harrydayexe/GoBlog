// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package parser

import (
	"bytes"
	"path/filepath"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	"go.abhg.dev/goldmark/wikilink"
)

// wikilinkExtender adds support for wikilinks:
//   - same-document heading anchors written as [[#Heading Text]] or
//     [[#Heading Text|Custom Label]]
//   - image embeds written as ![[foo.png]] or ![[foo.png|Alt text]]
//
// Parsing and rendering are provided by go.abhg.dev/goldmark/wikilink. Links
// are not validated against the document's headings or the assets directory:
// like a standard [text](#anchor) link, a link to a missing heading or image
// renders without error.
type wikilinkExtender struct {
	blogRoot string
}

// Extend registers the wikilink extension with a resolver for heading anchors
// and image embeds, and a transformer that drops the leading '#' from default
// link labels.
func (e wikilinkExtender) Extend(m goldmark.Markdown) {
	(&wikilink.Extender{Resolver: wikilinkResolver(e)}).Extend(m)
	m.Parser().AddOptions(parser.WithASTTransformers(
		util.Prioritized(headingLabelTransformer{}, 999),
	))
}

// wikilinkResolver resolves [[#Heading Text]] to "#heading-text", and
// ![[foo.png]] to "{blogRoot}images/foo.png" using the same rules as standard
// markdown images (see assetURL).
//
// Wikilinks to other pages, such as [[other-post]] or [[other-post#heading]],
// and embeds of non-image files, such as ![[notes.txt]], are not supported and
// resolve to nil, which renders only their label text.
type wikilinkResolver struct {
	blogRoot string
}

func (r wikilinkResolver) ResolveWikilink(n *wikilink.Node) ([]byte, error) {
	if n.Embed && isImageTarget(n.Target) {
		if dest, ok := assetURL(r.blogRoot, string(n.Target)); ok {
			return []byte(dest), nil
		}
		return n.Target, nil
	}
	heading, ok := linkedHeading(n)
	if !ok {
		return nil, nil
	}
	return append([]byte("#"), headingID(heading)...), nil
}

// isImageTarget reports whether target has an extension that the wikilink
// renderer emits as an <img> tag when embedded. The list must match the
// renderer's, otherwise a resolved non-image embed would render as a link.
func isImageTarget(target []byte) bool {
	switch filepath.Ext(string(target)) {
	case ".apng", ".avif", ".gif", ".jpg", ".jpeg", ".jfif", ".pjpeg", ".pjp", ".png", ".svg", ".webp":
		return true
	default:
		return false
	}
}

// headingLabelTransformer makes [[#Heading Text]] render with the label
// "Heading Text". The wikilink parser uses the raw target, "#Heading Text", as
// the label when no "|label" is given.
type headingLabelTransformer struct{}

func (headingLabelTransformer) Transform(doc *ast.Document, reader text.Reader, _ parser.Context) {
	src := reader.Source()
	_ = ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		n, ok := node.(*wikilink.Node)
		if !entering || !ok || n.ChildCount() != 1 {
			return ast.WalkContinue, nil
		}
		heading, ok := linkedHeading(n)
		label, isText := n.FirstChild().(*ast.Text)
		if ok && isText && bytes.Equal(label.Segment.Value(src), append([]byte("#"), heading...)) {
			label.Segment = label.Segment.WithStart(label.Segment.Start + 1)
		}
		return ast.WalkSkipChildren, nil
	})
}

// linkedHeading returns the heading text targeted by a same-document wikilink
// such as [[#Heading Text]], and false for links to other pages.
func linkedHeading(n *wikilink.Node) ([]byte, bool) {
	switch {
	case len(n.Target) == 0:
		// [[#Heading]]: the parser splits on '#', leaving an empty target.
		return n.Fragment, len(n.Fragment) > 0
	case n.Target[0] == '#':
		// [[#C# Tips]]: the parser splits on the last '#', so rejoin.
		return bytes.Join([][]byte{n.Target[1:], n.Fragment}, []byte("#")), true
	default:
		return nil, false
	}
}

// headingID slugifies heading text exactly as goldmark's WithAutoHeadingID
// does. A fresh context is used so no previously generated ids affect the
// result (goldmark would otherwise append -1, -2, ... suffixes).
func headingID(heading []byte) []byte {
	return parser.NewContext().IDs().Generate(heading, ast.KindHeading)
}

// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package parser

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// ErrUnresolvedHeadingLink is returned (wrapped) by ParseFile when a
// wikilink-style heading anchor such as [[#Future Work]] does not match any
// heading in the same document.
var ErrUnresolvedHeadingLink = errors.New("unresolved heading link")

// wikilinkTargetsKey stores the []wikilinkTarget collected while parsing a
// document, so they can be checked against the document's headings afterwards.
var wikilinkTargetsKey = parser.NewContextKey()

// wikilinkTarget records a single [[#Heading]] link found in a document.
type wikilinkTarget struct {
	heading string // heading text as written by the author
	id      string // slugified heading id the link points at
}

// wikilinkExtender adds support for same-document heading anchors written as
// [[#Heading Text]] or [[#Heading Text|Custom Label]].
type wikilinkExtender struct{}

// Extend registers the wikilink inline parser. The priority is lower than
// goldmark's link parser (200) so that [[#...]] is handled before the '['
// is treated as the start of a regular link.
func (wikilinkExtender) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithInlineParsers(
		util.Prioritized(wikilinkParser{}, 199),
	))
}

// wikilinkParser parses [[#Heading Text]] and [[#Heading Text|Label]] into a
// standard link node targeting the heading's auto-generated id.
type wikilinkParser struct{}

var (
	wikilinkOpen  = []byte("[[#")
	wikilinkClose = []byte("]]")
)

func (wikilinkParser) Trigger() []byte {
	return []byte{'['}
}

func (wikilinkParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, _ := block.PeekLine()
	if !bytes.HasPrefix(line, wikilinkOpen) {
		return nil
	}
	end := bytes.Index(line, wikilinkClose)
	if end < 0 {
		return nil
	}

	heading, label, _ := bytes.Cut(line[len(wikilinkOpen):end], []byte("|"))
	heading = bytes.TrimSpace(heading)
	label = bytes.TrimSpace(label)
	if len(heading) == 0 {
		return nil
	}
	if len(label) == 0 {
		label = heading
	}

	id := headingID(heading)
	targets, _ := pc.Get(wikilinkTargetsKey).([]wikilinkTarget)
	pc.Set(wikilinkTargetsKey, append(targets, wikilinkTarget{
		heading: string(heading),
		id:      string(id),
	}))

	block.Advance(end + len(wikilinkClose))

	link := ast.NewLink()
	link.Destination = append([]byte("#"), id...)
	// Copy the label so the node does not alias the source buffer.
	link.AppendChild(link, ast.NewString(bytes.Clone(label)))
	return link
}

// headingID slugifies heading text exactly as goldmark's WithAutoHeadingID
// does. A fresh context is used so no previously generated ids affect the
// result (goldmark would otherwise append -1, -2, ... suffixes).
func headingID(heading []byte) []byte {
	return parser.NewContext().IDs().Generate(heading, ast.KindHeading)
}

// validateWikilinks returns an error wrapping ErrUnresolvedHeadingLink for
// every [[#Heading]] link in pctx that does not match a heading id in doc.
func validateWikilinks(doc ast.Node, pctx parser.Context) error {
	targets, _ := pctx.Get(wikilinkTargetsKey).([]wikilinkTarget)
	if len(targets) == 0 {
		return nil
	}

	ids := make(map[string]bool)
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering || n.Kind() != ast.KindHeading {
			return ast.WalkContinue, nil
		}
		if id, ok := n.AttributeString("id"); ok {
			if b, ok := id.([]byte); ok {
				ids[string(b)] = true
			}
		}
		return ast.WalkSkipChildren, nil
	})

	var errs []error
	for _, t := range targets {
		if !ids[t.id] {
			errs = append(errs, fmt.Errorf("%w: [[#%s]] (no heading with id %q)", ErrUnresolvedHeadingLink, t.heading, t.id))
		}
	}
	return errors.Join(errs...)
}

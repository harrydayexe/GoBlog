// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package parser reads markdown files with YAML frontmatter and converts them
// to Post objects for the GoBlog system.
//
// The parser uses goldmark for markdown processing and supports:
//   - YAML frontmatter for post metadata
//   - Syntax highlighting for code blocks
//   - Footnotes
//   - Auto-generated heading IDs
//   - Wikilink-style heading anchors
//   - Images with alt text, intrinsic dimensions and lazy loading
//   - HTML sanitization
//
// # Heading Anchor Links
//
// Headings are given auto-generated ids (e.g. "## Future Work" becomes
// id="future-work"). In addition to standard [text](#future-work) links,
// a heading in the same document can be linked with wikilink syntax:
//
//	See [[#Future Work]] for details.            -> <a href="#future-work">Future Work</a>
//	See [[#Future Work|what comes next]].        -> <a href="#future-work">what comes next</a>
//
// Wikilinks are parsed by go.abhg.dev/goldmark/wikilink. As with standard
// links, targets are not validated: a link to a heading that does not exist
// still renders, without error. Cross-post wikilinks such as [[other-post]] or
// [[other-post#heading]] are not supported and render as their label text.
//
// # Images
//
// Images can be written in standard markdown or as wikilink embeds. Relative
// paths are resolved against the assets directory and rewritten to
// root-absolute URLs under the blog root (see WithBlogRoot), so they work from
// a post served at {BlogRoot}posts/{slug}:
//
//	![A diagram](images/pipeline.png)
//	![A diagram](pipeline.png)
//	![[pipeline.png|A diagram]]
//
// All three render as:
//
//	<img src="/images/pipeline.png" alt="A diagram" width="1200" height="800" loading="lazy" decoding="async" />
//
// A bare ![[pipeline.png]] renders the same tag with alt="", the correct
// signal for "no description available"; give it a "|label" when the image
// carries meaning. Subdirectories are preserved ("screenshots/a.png" becomes
// "/images/screenshots/a.png"). Absolute URLs, root-relative paths
// ("/static/x.png") and paths containing ".." are left untouched; they still
// get loading and decoding, but no dimensions, as there is no local file to
// measure. Embeds of non-image files, such as ![[notes.txt]], render as their
// label text.
//
// Width and height are the image's intrinsic pixel size, read from the asset
// file supplied with WithAssetsDir. Both are emitted, or neither: when no
// assets filesystem is configured, or the file is missing, unreadable, or in a
// format whose header cannot be decoded (SVG, AVIF and WebP), the image simply
// renders without them.
//
// As with heading links, image paths are not validated: an image missing from
// the assets directory still renders, without error.
//
// Basic usage:
//
//	import (
//		"context"
//		"os"
//		"github.com/harrydayexe/GoBlog/v2/pkg/parser"
//	)
//
//	// Parse a single file
//	p := parser.New()
//	post, err := p.ParseFile(context.Background(), os.DirFS("/path/to/posts"), "my-post.md")
//	if err != nil {
//		// handle error
//	}
//
//	// Parse all markdown files in a directory
//	posts, err := p.ParseDirectory(context.Background(), os.DirFS("/path/to/posts"))
//	if err != nil {
//		// handle error - may be ParseErrors with partial results
//	}
//
// # Configuration
//
// The parser can be customized using functional options:
//
//	// Disable code highlighting
//	p := parser.New(parser.WithCodeHighlighting(false))
//
//	// Enable footnote support
//	p := parser.New(parser.WithFootnote())
//
//	// Inject a structured logger
//	p := parser.New(parser.WithLogger(myLogger))
//
//	// Resolve image URLs under a subdirectory deployment
//	p := parser.New(parser.WithBlogRoot("/blog/"))
//
//	// Measure images so they render with width and height
//	p := parser.New(parser.WithAssetsDir(os.DirFS("posts/images")))
//
//	// Combine multiple options
//	p := parser.New(
//	    parser.WithCodeHighlighting(true),
//	    parser.WithFootnote(),
//	    parser.WithLogger(myLogger),
//	)
//
// See the Option functions for all available configuration options.
//
// # Syntax Highlighting CSS
//
// The parser renders highlighted code blocks using CSS classes (via
// chroma's html.WithClasses option) rather than inline styles. This means
// the generated HTML will contain class names like .chroma, .k, .s, etc.,
// but will not be visually styled until a matching stylesheet is included
// in your page.
//
// Generate the stylesheet for a given style at startup and embed it in your
// templates:
//
//	import (
//	    chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
//	    "github.com/alecthomas/chroma/v2/styles"
//	    "strings"
//	)
//
//	formatter := chromahtml.New(chromahtml.WithClasses(true))
//	style := styles.Get("monokai")
//	var sb strings.Builder
//	formatter.WriteCSS(&sb, style)
//	chromaCSS := sb.String() // embed in a <style> tag in your template
//
// The class names follow the Pygments short-name convention. A full reference
// is available in the chroma source:
// https://github.com/alecthomas/chroma/blob/master/types.go
//
// A Parser is safe for concurrent use by multiple goroutines after creation.
package parser

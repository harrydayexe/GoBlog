// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package generator

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"strings"

	"github.com/harrydayexe/GoBlog/v2/pkg/models"
	"gopkg.in/yaml.v3"
)

// seriesDocument is the top-level shape of the series YAML file.
//
// Series is a pointer so that a file missing the key entirely can be told apart
// from one declaring an empty list: "series: []" enables the feature with no
// series in it, while a missing key is a malformed file.
type seriesDocument struct {
	Series *[]seriesDefinition `yaml:"series"`
}

// seriesDefinition is one entry of the series file's top-level list.
type seriesDefinition struct {
	Name        string   `yaml:"name"`
	Slug        string   `yaml:"slug"`
	Description string   `yaml:"description"`
	Posts       []string `yaml:"posts"`
}

// resolvedSeries is a validated series: its display metadata plus the posts it
// lists, in the reading order the series file gave them.
type resolvedSeries struct {
	Name        string
	Slug        string
	Description string
	Posts       models.PostList
}

// loadSeries reads, parses and validates the configured series file against the
// parsed posts, returning every series in file order.
//
// Every rule in the file format is a hard error naming the series and the
// offending value, so a typo is reported rather than silently dropping content:
// the YAML must parse and declare a top-level "series" key, each series needs a
// non-empty name and a non-empty posts list, every listed filename must match a
// post, no post may appear twice in a series or in two series, and no two series
// may share a slug. Unknown keys are rejected so that "post:" for "posts:" fails
// loudly.
//
// Posts are matched on their source filename relative to the posts directory
// rather than on their slug: slugs are derived from titles, so a slug reference
// would break whenever a post was retitled.
func (g *Generator) loadSeries(posts models.PostList) ([]resolvedSeries, error) {
	file := g.SeriesFile.Path

	raw, err := fs.ReadFile(g.SeriesFile.FS, file)
	if err != nil {
		return nil, fmt.Errorf("reading series file %q: %w", file, err)
	}

	dec := yaml.NewDecoder(bytes.NewReader(raw))
	dec.KnownFields(true)

	var doc seriesDocument
	// An empty file decodes to io.EOF, which the missing-key check below reports
	// in terms the user can act on.
	if err := dec.Decode(&doc); err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("parsing series file %q: %w", file, err)
	}
	if doc.Series == nil {
		return nil, fmt.Errorf("series file %q: missing top-level \"series\" key", file)
	}

	byFilename := make(map[string]*models.Post, len(posts))
	for _, post := range posts {
		byFilename[post.SourcePath] = post
	}

	definitions := *doc.Series
	resolved := make([]resolvedSeries, 0, len(definitions))
	slugOwners := make(map[string]string, len(definitions)) // slug -> series name
	postOwners := make(map[string]string, len(posts))       // filename -> series name

	for i, def := range definitions {
		name := strings.TrimSpace(def.Name)
		if name == "" {
			return nil, fmt.Errorf("series file %q: series at index %d has no name", file, i)
		}

		slug, err := seriesSlug(file, name, def.Slug)
		if err != nil {
			return nil, err
		}
		if other, taken := slugOwners[slug]; taken {
			return nil, fmt.Errorf("series file %q: series %q and %q both use the slug %q", file, other, name, slug)
		}
		slugOwners[slug] = name

		if len(def.Posts) == 0 {
			return nil, fmt.Errorf("series file %q: series %q lists no posts", file, name)
		}

		seriesPosts := make(models.PostList, 0, len(def.Posts))
		listed := make(map[string]bool, len(def.Posts))
		for _, filename := range def.Posts {
			clean := path.Clean(filename)

			post, ok := byFilename[clean]
			if !ok {
				return nil, fmt.Errorf("series file %q: series %q lists post %q, which does not match any post in the posts directory", file, name, filename)
			}
			if listed[clean] {
				return nil, fmt.Errorf("series file %q: series %q lists post %q more than once", file, name, filename)
			}
			if owner, taken := postOwners[clean]; taken {
				return nil, fmt.Errorf("series file %q: post %q is listed in both series %q and %q; a post may belong to at most one series", file, filename, owner, name)
			}

			listed[clean] = true
			postOwners[clean] = name
			seriesPosts = append(seriesPosts, post)
		}

		resolved = append(resolved, resolvedSeries{
			Name:        name,
			Slug:        slug,
			Description: strings.TrimSpace(def.Description),
			Posts:       seriesPosts,
		})
	}

	return resolved, nil
}

// seriesSlug returns the URL segment for a series: the explicit slug when the
// series declares one, otherwise one derived from its name with the same rules
// post slugs use. Either form is an error when it slugifies to nothing, since a
// series page needs a path.
func seriesSlug(file, name, explicit string) (string, error) {
	if explicit != "" {
		slug := models.Slugify(explicit)
		if slug == "" {
			return "", fmt.Errorf("series file %q: series %q has slug %q, which is empty once slugified", file, name, explicit)
		}
		return slug, nil
	}

	slug := models.Slugify(name)
	if slug == "" {
		return "", fmt.Errorf("series file %q: series %q has a name that is empty once slugified; give the series an explicit slug", file, name)
	}
	return slug, nil
}

// seriesDescription returns the meta description for a series page: the
// author's description when the series file declares one, and a generated
// sentence otherwise, so the page never emits an empty meta description.
func seriesDescription(s resolvedSeries) string {
	if s.Description != "" {
		return s.Description
	}
	return fmt.Sprintf("Posts in the %s series", s.Name)
}

// postSeriesContexts maps each post's source filename to the series context its
// page should render: which series the post is in, where it sits, and the parts
// either side of it. Posts in no series are absent from the map, which leaves
// their PostPageData.Series nil.
func (g *Generator) postSeriesContexts(series []resolvedSeries) map[string]*models.PostSeries {
	contexts := make(map[string]*models.PostSeries)

	for _, s := range series {
		for i, post := range s.Posts {
			ctx := &models.PostSeries{
				Name:        s.Name,
				Slug:        s.Slug,
				Description: s.Description,
				Path:        g.pagePath("series", s.Slug),
				Posts:       s.Posts,
				Position:    i + 1,
				Total:       len(s.Posts),
			}
			if i > 0 {
				ctx.Prev = s.Posts[i-1]
			}
			if i < len(s.Posts)-1 {
				ctx.Next = s.Posts[i+1]
			}
			contexts[post.SourcePath] = ctx
		}
	}

	return contexts
}

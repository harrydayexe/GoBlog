// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package generator

import (
	"strings"
)

// defaultRobotsRules is the rule block emitted when no custom body is
// configured via config.WithRobotsTxt.
const defaultRobotsRules = "User-agent: *\nAllow: /"

// buildRobotsTxt generates the site's robots.txt.
//
// The body is GoBlog's default "allow everything" rule block, or the body
// supplied via config.WithRobotsTxt when set — a custom body replaces the
// rules wholesale rather than being appended to them.
//
// A "Sitemap:" line naming the absolute sitemap URL is appended after the
// rules, so the file stays valid wherever it is deployed. It is emitted only
// when a sitemap was actually produced, which hasSitemap reports: advertising
// a sitemap that is neither written nor served — because
// config.WithDisableSitemap() was applied, or because the blog has no posts —
// would point crawlers at a 404.
func (g *Generator) buildRobotsTxt(hasSitemap bool) []byte {
	rules := defaultRobotsRules
	if g.RobotsTxt.Body != "" {
		rules = g.RobotsTxt.Body
	}

	var b strings.Builder
	b.WriteString(strings.TrimRight(rules, "\n"))
	b.WriteString("\n")

	if hasSitemap {
		b.WriteString("\nSitemap: ")
		b.WriteString(g.canonicalURL(g.sitemapPath()))
		b.WriteString("\n")
	}

	return []byte(b.String())
}

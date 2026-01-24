// Package templates provides embedded default templates for the GoBlog system.
//
// The Default variable contains the embedded filesystem with default HTML templates
// used by the generator when no custom templates are provided.
package templates

import "embed"

//go:embed default/**/*.tmpl
var Default embed.FS

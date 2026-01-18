package templates

import "embed"

//go:embed default/**/*.tmpl
var Default embed.FS

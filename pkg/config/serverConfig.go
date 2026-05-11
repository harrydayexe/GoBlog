package config

import "io/fs"

// ServerConfig is the top-level configuration for the HTTP server.
//
// Server holds server-level options (port, host, middleware).
// Gen holds generator-level options (site title, environment, custom data, etc.).
// TemplateDir overrides the template filesystem; when nil the built-in
// templates are used.
// RendererOpts holds renderer-level options (custom template functions) and
// is forwarded to the internal [github.com/harrydayexe/GoBlog/v2/pkg/generator.NewTemplateRenderer]
// call so users of the server API can register custom FuncMap entries without
// constructing a renderer manually.
type ServerConfig struct {
	Server       []BaseServerOption
	Gen          []GeneratorOption
	TemplateDir  fs.FS
	RendererOpts []RendererOption
}

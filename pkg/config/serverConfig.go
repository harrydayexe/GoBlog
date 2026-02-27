package config

import "io/fs"

type ServerConfig struct {
	Server      []BaseServerOption
	Gen         []GeneratorOption
	TemplateDir fs.FS
}

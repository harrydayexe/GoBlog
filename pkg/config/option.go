package config

import "io/fs"

type Option struct {
	WithRawOutputFunc func(v *RawOutput)
	WithTemplatesFunc func(v *TemplatesDir)
}

type RawOutput struct{ RawOutput bool }

func WithRawOutput() Option {
	return Option{
		WithRawOutputFunc: func(v *RawOutput) {
			v.RawOutput = true
		},
	}
}

type TemplatesDir struct{ TemplatesDir fs.FS }

func WithTemplatesDir(templatesDir fs.FS) Option {
	return Option{
		WithTemplatesFunc: func(v *TemplatesDir) {
			v.TemplatesDir = templatesDir
		},
	}
}

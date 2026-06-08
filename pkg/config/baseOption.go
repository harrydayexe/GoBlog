// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import "log/slog"

// BaseOption represents a configuration option that can be applied to
// many different instances during construction.
//
// Options use the functional options pattern, where each option function
// returns an BaseOption struct containing one or more function pointers that
// modify specific configuration fields.
//
// This type should not be constructed directly by users. Instead, use the
// provided option functions like WithBlogRoot() or WithLogger().
type BaseOption struct {
	WithBlogRootFunc func(v *BlogRoot)
	WithLoggerFunc   func(v *Logger)
}

// Logger is a configuration type that holds a [log/slog.Logger] for structured
// logging.
//
// This type is typically embedded in constructor configuration structs and
// should be set using the [WithLogger] option function. When no logger is
// supplied, constructors fall back to [log/slog.Default] at construction time.
type Logger struct{ Logger *slog.Logger }

// WithLogger returns a BaseOption that sets the structured logger used by the
// component.
//
// Passing nil is permitted and is treated the same as not supplying the option:
// the component falls back to [log/slog.Default] at construction time.
//
// Example usage:
//
//	import "log/slog"
//
//	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
//
//	gen := generator.New(fsys, renderer, config.WithLogger(logger))
//	w, err := watcher.New("posts/", config.WithLogger(logger))
//	writer := outputter.NewDirectoryWriter("output/", config.WithLogger(logger).AsGeneratorOption())
func WithLogger(l *slog.Logger) BaseOption {
	return BaseOption{
		WithLoggerFunc: func(v *Logger) { v.Logger = l },
	}
}

// BlogRoot is a configuration type that holds the blog's root path
//
// This type is typically embedded in generator configuration structs
// and should be set using the WithBlogRoot() option function.
type BlogRoot string

// WithBlogRoot returns an Option that sets the blog's root path.
//
// The blog root is used in generated HTML pages and templates.
//
// Example usage:
//
//	gen := generator.New(fsys, renderer, config.WithBlogRoot("/blog/"))
func WithBlogRoot(root string) BaseOption {
	return BaseOption{
		WithBlogRootFunc: func(v *BlogRoot) {
			*v = BlogRoot(root)
		},
	}
}

// AsOption returns a BaseOption that re-applies this BlogRoot value to another
// component, enabling a resolved blog root to be forwarded to sub-components
// without reaching into its underlying string.
func (o BlogRoot) AsOption() BaseOption {
	return WithBlogRoot(string(o))
}

// AsOption returns a BaseOption that re-applies this Logger value to another
// component, enabling a resolved logger to be forwarded to sub-components
// without reaching into its underlying [log/slog.Logger].
func (o Logger) AsOption() BaseOption {
	return WithLogger(o.Logger)
}

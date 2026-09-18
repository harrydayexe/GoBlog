// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"io/fs"
	"log/slog"
)

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
	WithBlogRootFunc  func(v *BlogRoot)
	WithLoggerFunc    func(v *Logger)
	WithAssetsDirFunc func(v *AssetsDir)
}

// AssetsDir is a configuration type that holds the filesystem containing the
// blog's static assets (images).
//
// Assets are served by the HTTP server at {BlogRoot}images/<path> and copied
// by the outputter into <outputDir>/images/. The generator also forwards the
// filesystem to the parser, which reads each referenced image's header to
// render it with its intrinsic width and height. When FS is nil, or its root
// cannot be stat'd as a directory, asset support is off: no route is
// registered, nothing is copied, and images render without dimensions.
//
// This type is typically embedded in server and outputter configuration
// structs and should be set using the [WithAssetsDir] option function.
type AssetsDir struct{ FS fs.FS }

// WithAssetsDir returns a BaseOption that sets the filesystem images are
// served, copied and measured from.
//
// Prefer a filesystem that cannot escape its root, such as the one returned
// by [os.Root.FS]; [os.DirFS] follows symbolic links that point outside the
// directory. Passing nil disables asset support.
//
// Example usage:
//
//	root, err := os.OpenRoot("posts/images")
//	if err != nil {
//	    return err
//	}
//	defer root.Close()
//
//	writer := outputter.NewDirectoryWriter("output/", config.WithAssetsDir(root.FS()).AsGeneratorOption())
func WithAssetsDir(fsys fs.FS) BaseOption {
	return BaseOption{
		WithAssetsDirFunc: func(v *AssetsDir) { v.FS = fsys },
	}
}

// AsOption returns a BaseOption that re-applies this AssetsDir value to
// another component.
func (o AssetsDir) AsOption() BaseOption {
	return WithAssetsDir(o.FS)
}

// Enabled reports whether FS is set and its root is a readable directory.
func (o AssetsDir) Enabled() bool {
	if o.FS == nil {
		return false
	}
	fi, err := fs.Stat(o.FS, ".")
	return err == nil && fi.IsDir()
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
//	gen := generator.New(fsys, renderer, config.WithLogger(logger).AsGeneratorOption())
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

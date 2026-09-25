// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import "time"

// WatcherOption represents a configuration option for a filesystem Watcher.
// Options use the functional options pattern; each value carries a function
// pointer that modifies a specific watcher setting.
//
// This type should not be constructed directly. Use the provided option
// functions: [WithDebounce], [WithWatchFile], or call
// [BaseOption.AsWatcherOption] on a [BaseOption] value (e.g. from [WithLogger]).
type WatcherOption struct {
	BaseOption

	WithDebounceFunc  func(v *WatcherDebounce)
	WithWatchFileFunc func(v *WatchFiles)
}

// WatcherDebounce holds the event debounce duration for the watcher.
// Events that arrive within this window are coalesced into a single callback
// invocation.
type WatcherDebounce struct{ Debounce time.Duration }

// WithDebounce returns a WatcherOption that sets the event debounce window.
// File-system events that arrive within this duration of each other are
// coalesced into a single onChange invocation. Defaults to 300ms.
//
// Example usage:
//
//	w, err := watcher.New("posts/", config.WithDebounce(500*time.Millisecond))
func WithDebounce(d time.Duration) WatcherOption {
	return WatcherOption{
		WithDebounceFunc: func(v *WatcherDebounce) { v.Debounce = d },
	}
}

// WatchFiles holds individual files that should trigger the watcher's callback
// in addition to the markdown files under the watched directory.
//
// The watcher watches each file's parent directory and matches events against
// the cleaned paths in Paths, so a file replaced by an atomic rename — how most
// editors save — is still noticed, and one that is deleted and recreated does
// not need a new watch descriptor.
//
// This type is typically embedded in the watcher struct and should be set using
// the [WithWatchFile] option function.
type WatchFiles struct{ Paths []string }

// WithWatchFile returns a WatcherOption that adds a single file to the set the
// watcher reacts to, on top of the markdown files under its root directory.
//
// The goblog serve command uses it for the series file (--series-file), which is
// neither markdown nor necessarily inside the posts directory, so that editing,
// creating or deleting it regenerates the blog exactly as editing a post does.
//
// Multiple calls accumulate. The file's parent directory must exist when
// watcher.New is called.
//
// Example usage:
//
//	w, err := watcher.New("posts/", config.WithWatchFile("posts/series.yml"))
func WithWatchFile(path string) WatcherOption {
	return WatcherOption{
		WithWatchFileFunc: func(v *WatchFiles) {
			v.Paths = append(v.Paths, path)
		},
	}
}

// AsWatcherOption returns a WatcherOption that applies this BaseOption to a
// watcher instance, enabling a BaseOption (e.g. from [WithLogger]) to be
// passed to watcher constructors alongside other WatcherOptions.
func (o BaseOption) AsWatcherOption() WatcherOption {
	return WatcherOption{
		BaseOption: o,
	}
}

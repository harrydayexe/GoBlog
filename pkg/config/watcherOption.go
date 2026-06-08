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
// functions: [WithDebounce], or call [BaseOption.AsWatcherOption] on a
// [BaseOption] value (e.g. from [WithLogger]).
type WatcherOption struct {
	BaseOption

	WithDebounceFunc func(v *WatcherDebounce)
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

// AsWatcherOption returns a WatcherOption that applies this BaseOption to a
// watcher instance, enabling a BaseOption (e.g. from [WithLogger]) to be
// passed to watcher constructors alongside other WatcherOptions.
func (o BaseOption) AsWatcherOption() WatcherOption {
	return WatcherOption{
		BaseOption: o,
	}
}

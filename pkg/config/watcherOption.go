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
// functions: WithDebounce, or the inherited [BaseOption] functions such as
// [WithLogger].
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

// WithBaseWatcherOption wraps a [BaseOption] as a [WatcherOption] so it can be
// passed to watcher constructors that accept [WatcherOption] values.
// Use this when you have a BaseOption (e.g. from [WithLogger]) and need to
// supply it alongside other WatcherOptions.
func WithBaseWatcherOption(baseOption BaseOption) WatcherOption {
	return WatcherOption{
		BaseOption: baseOption,
	}
}

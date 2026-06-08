// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package watcher provides recursive filesystem watching with debouncing for
// blog post directories.
//
// It wraps github.com/fsnotify/fsnotify with automatic recursive directory
// watching (fsnotify only watches individual directories non-recursively) and
// a configurable debounce window that coalesces rapid bursts of events — such
// as those produced by editors that write files in multiple steps — into a
// single callback invocation.
//
// # Basic Usage
//
// Create a Watcher rooted at a posts directory and run it alongside a
// pkg/server.Server to achieve live content reload:
//
//	import (
//	    "context"
//	    "log/slog"
//	    "os"
//
//	    "github.com/harrydayexe/GoBlog/v2/pkg/config"
//	    "github.com/harrydayexe/GoBlog/v2/pkg/server"
//	    "github.com/harrydayexe/GoBlog/v2/pkg/watcher"
//	)
//
//	postsPath := "posts/"
//	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
//
//	w, err := watcher.New(postsPath, config.WithBaseWatcherOption(config.WithLogger(logger)))
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//
//	// srv is a *server.Server created with server.New(...)
//	go w.Run(ctx, func(ctx context.Context) {
//	    if err := srv.UpdatePosts(os.DirFS(postsPath), ctx); err != nil {
//	        slog.Warn("failed to reload posts", "error", err)
//	    }
//	})
//
//	// srv.Run blocks until the server shuts down.
//	if err := srv.Run(ctx); err != nil {
//	    log.Fatal(err)
//	}
//
// # Setup Errors
//
// New fails hard on any problem detected at setup time: root path missing or
// not a directory, fsnotify initialisation failure, or any individual
// directory failing to be added to the watch list (e.g. due to exceeding the
// Linux inotify watch limit). These are surfaced immediately so the user is
// aware that watching is not working as requested.
//
// # Runtime Behaviour
//
// Once Run is active, errors are logged at warn level and the loop continues.
// This includes fsnotify errors and failures adding newly created
// subdirectories to the watch set.
//
// Subdirectories created after the Watcher is constructed are automatically
// picked up by Run when the parent directory fires a Create event.
//
// Only changes to files with a .md extension trigger the onChange callback.
// All other file types (images, CSS, YAML, etc.) are silently ignored, as
// are common editor temporary files (dotfiles, *.swp, *~, etc.).
// Deletion of already-watched directories is not currently handled:
// the watcher stops receiving events for deleted directory trees but does not
// error. See the project issue tracker for a planned improvement.
//
// # Concurrency
//
// New and Run are not safe for concurrent use on the same Watcher. Typically
// a single Watcher is created once and Run in a dedicated goroutine.
package watcher

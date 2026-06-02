// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package watcher

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/harrydayexe/GoBlog/v2/pkg/config"
)

// Watcher watches a directory tree for filesystem changes and invokes a
// callback, debounced, whenever changes are detected.
type Watcher struct {
	path string

	config.WatcherDebounce

	fw *fsnotify.Watcher
}

// New creates a Watcher rooted at path. It recursively watches path and all
// subdirectories that exist at creation time. Subdirectories created after
// New returns are picked up automatically inside Run when their parent fires
// a Create event.
//
// New fails immediately if any part of setup fails — including the root path
// being missing, fsnotify initialisation failing, or any subdirectory failing
// to be added to the watch list. Callers should not attempt to fall back
// silently; surface the error to the user.
//
// Call Run to begin receiving events.
func New(path string, opts ...config.WatcherOption) (*Watcher, error) {
	fw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("watcher: failed to create fsnotify watcher: %w", err)
	}

	w := &Watcher{
		path:            path,
		WatcherDebounce: config.WatcherDebounce{Debounce: 300 * time.Millisecond},
		fw:              fw,
	}

	for _, opt := range opts {
		if opt.WithDebounceFunc != nil {
			opt.WithDebounceFunc(&w.WatcherDebounce)
		}
	}

	if err := w.addDirs(path); err != nil {
		fw.Close()
		return nil, err
	}

	return w, nil
}

// Run blocks, watching for filesystem events, until ctx is canceled.
// onChange is invoked (with a child context) whenever file changes are
// detected, coalesced over the configured debounce window. Multiple events
// within the window trigger only one onChange call.
//
// Errors from the fsnotify event stream are logged at warn level but do not
// stop the watch loop. Run returns nil when ctx is canceled.
func (w *Watcher) Run(ctx context.Context, onChange func(context.Context)) error {
	defer w.fw.Close()

	logger := slog.Default()

	timer := time.NewTimer(0)
	<-timer.C // drain so it doesn't fire immediately

	for {
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil

		case event, ok := <-w.fw.Events:
			if !ok {
				return nil
			}
			if isNoise(event.Name) {
				continue
			}
			logger.DebugContext(ctx, "file event", slog.String("path", event.Name), slog.String("op", event.Op.String()))

			// If a new directory appeared, watch it — then skip the
			// debounce reset since a bare directory create is not a
			// content change.
			if event.Has(fsnotify.Create) {
				if fi, err := os.Stat(event.Name); err == nil && fi.IsDir() {
					if err := w.fw.Add(event.Name); err != nil {
						logger.WarnContext(ctx, "watcher: failed to add new directory", slog.String("path", event.Name), slog.Any("error", err))
					} else {
						logger.DebugContext(ctx, "watcher: watching new directory", slog.String("path", event.Name))
					}
					continue
				}
			}

			// Only regenerate for markdown file changes.
			if filepath.Ext(event.Name) != ".md" {
				continue
			}

			// Reset debounce timer.
			timer.Stop()
			timer.Reset(w.Debounce)

		case err, ok := <-w.fw.Errors:
			if !ok {
				return nil
			}
			logger.WarnContext(ctx, "watcher: fsnotify error", slog.Any("error", err))

		case <-timer.C:
			logger.InfoContext(ctx, "watcher: posts changed, regenerating")
			childCtx, cancel := context.WithCancel(ctx)
			defer cancel()
			onChange(childCtx)
		}
	}
}

// addDirs walks path and adds every directory (including path itself) to the
// fsnotify watcher. Returns an error if path is not a directory or if any
// Add call fails.
func (w *Watcher) addDirs(path string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("watcher: cannot stat path %q: %w", path, err)
	}
	if !fi.IsDir() {
		return fmt.Errorf("watcher: path %q is not a directory", path)
	}

	return filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("watcher: error walking %q: %w", p, err)
		}
		if !d.IsDir() {
			return nil
		}
		if err := w.fw.Add(p); err != nil {
			return fmt.Errorf("watcher: failed to watch directory %q: %w", p, err)
		}
		slog.Default().Debug("watcher: watching directory", slog.String("path", p))
		return nil
	})
}

// isNoise reports whether a filesystem event path should be ignored.
// Matches common editor temporary files and OS metadata files.
func isNoise(name string) bool {
	base := filepath.Base(name)
	if strings.HasPrefix(base, ".") {
		return true
	}
	return strings.HasSuffix(base, ".swp") ||
		strings.HasSuffix(base, ".swx") ||
		strings.HasSuffix(base, "~")
}

// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package watcher

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/harrydayexe/GoBlog/v2/pkg/config"
)

// TestRemovedSubdirReleasesWatch verifies that deleting a watched subdirectory
// causes its watch descriptor to be released within one event loop iteration.
func TestRemovedSubdirReleasesWatch(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	sub := filepath.Join(dir, "subdir")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatalf("Mkdir error = %v", err)
	}

	w, err := New(dir, config.WithDebounce(50*time.Millisecond))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	go w.Run(ctx, func(context.Context) {}) //nolint:errcheck

	// Give the watcher time to start listening.
	time.Sleep(50 * time.Millisecond)

	// Confirm the subdir is currently watched.
	if !slicesContains(w.fw.WatchList(), sub) {
		t.Fatalf("expected %q to be in watchedDirs before removal", sub)
	}

	if err := os.Remove(sub); err != nil {
		t.Fatalf("Remove error = %v", err)
	}

	// Poll until the watch descriptor is released (or timeout).
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if !slicesContains(w.fw.WatchList(), sub) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Errorf("%q still present in watchedDirs after removal", sub)
}

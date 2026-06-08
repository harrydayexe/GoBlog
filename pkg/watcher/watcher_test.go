// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package watcher_test

import (
	"context"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/harrydayexe/GoBlog/v2/pkg/config"
	"github.com/harrydayexe/GoBlog/v2/pkg/watcher"
)

const shortDebounce = 50 * time.Millisecond

// waitForCount blocks until count reaches target or the deadline passes.
func waitForCount(t *testing.T, count *atomic.Int64, target int64, deadline time.Duration) {
	t.Helper()
	end := time.Now().Add(deadline)
	for time.Now().Before(end) {
		if count.Load() >= target {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Errorf("onChange called %d time(s), want at least %d", count.Load(), target)
}

// TestNew_MissingPath verifies that New fails when the path does not exist.
func TestNew_MissingPath(t *testing.T) {
	t.Parallel()
	_, err := watcher.New("/nonexistent/path/for/goblog/watcher/test")
	if err == nil {
		t.Error("New() with missing path returned nil error, want error")
	}
}

// TestNew_FilePath verifies that New fails when path is a file, not a directory.
func TestNew_FilePath(t *testing.T) {
	t.Parallel()
	f, err := os.CreateTemp(t.TempDir(), "not-a-dir-*.md")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	_, err = watcher.New(f.Name())
	if err == nil {
		t.Error("New() with file path returned nil error, want error")
	}
}

// TestRun_Cancellation verifies that Run returns promptly when ctx is canceled.
func TestRun_Cancellation(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	w, err := watcher.New(dir, config.WithDebounce(shortDebounce))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	done := make(chan error, 1)
	go func() { done <- w.Run(ctx, func(context.Context) {}) }()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Error("Run() did not return within 2s after ctx cancellation")
	}
}

// TestRun_FileChange verifies that onChange is called when a file is written.
func TestRun_FileChange(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	w, err := watcher.New(dir, config.WithDebounce(shortDebounce))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	var count atomic.Int64
	go w.Run(ctx, func(context.Context) { count.Add(1) }) //nolint:errcheck

	// Give the watcher time to start listening before writing.
	time.Sleep(50 * time.Millisecond)

	if err := os.WriteFile(filepath.Join(dir, "post.md"), []byte("hello"), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	waitForCount(t, &count, 1, 3*time.Second)
}

// TestRun_Debounce verifies that rapid writes coalesce into a single onChange call.
func TestRun_Debounce(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	w, err := watcher.New(dir, config.WithDebounce(200*time.Millisecond))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	var count atomic.Int64
	go w.Run(ctx, func(context.Context) { count.Add(1) }) //nolint:errcheck

	time.Sleep(50 * time.Millisecond)

	// Write three files in quick succession.
	for i := range 3 {
		path := filepath.Join(dir, "post.md")
		_ = i
		if err := os.WriteFile(path, []byte("hello"), 0o644); err != nil {
			t.Fatalf("WriteFile error = %v", err)
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Wait for debounce to settle.
	time.Sleep(400 * time.Millisecond)

	if got := count.Load(); got != 1 {
		t.Errorf("onChange called %d time(s), want 1 (debounced)", got)
	}
}

// TestRun_NewSubdirTracked verifies that files written inside a newly created
// subdirectory also trigger onChange.
func TestRun_NewSubdirTracked(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	w, err := watcher.New(dir, config.WithDebounce(shortDebounce))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	var count atomic.Int64
	go w.Run(ctx, func(context.Context) { count.Add(1) }) //nolint:errcheck

	time.Sleep(50 * time.Millisecond)

	sub := filepath.Join(dir, "subdir")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatalf("Mkdir error = %v", err)
	}
	// Allow the watcher to pick up the new directory.
	time.Sleep(100 * time.Millisecond)

	if err := os.WriteFile(filepath.Join(sub, "post.md"), []byte("hello"), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	waitForCount(t, &count, 1, 3*time.Second)
}

// TestRun_NonMarkdownIgnored verifies that changes to non-.md files do not trigger onChange.
func TestRun_NonMarkdownIgnored(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	w, err := watcher.New(dir, config.WithDebounce(shortDebounce))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	var count atomic.Int64
	go w.Run(ctx, func(context.Context) { count.Add(1) }) //nolint:errcheck

	time.Sleep(50 * time.Millisecond)

	nonMarkdown := []string{"post.html", "style.css", "config.yaml", "image.png"}
	for _, name := range nonMarkdown {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("content"), 0o644); err != nil {
			t.Fatalf("WriteFile %q error = %v", name, err)
		}
	}

	time.Sleep(shortDebounce + 200*time.Millisecond)

	if got := count.Load(); got != 0 {
		t.Errorf("onChange called %d time(s) for non-markdown files, want 0", got)
	}
}

// TestRun_SubdirRemoveThenRecreateTracked verifies that removing a watched
// subdirectory and recreating it still triggers onChange when a file is written
// inside the new directory.
func TestRun_SubdirRemoveThenRecreateTracked(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	sub := filepath.Join(dir, "subdir")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatalf("Mkdir error = %v", err)
	}

	w, err := watcher.New(dir, config.WithDebounce(shortDebounce))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	var count atomic.Int64
	go w.Run(ctx, func(context.Context) { count.Add(1) }) //nolint:errcheck

	time.Sleep(50 * time.Millisecond)

	// Remove the subdirectory.
	if err := os.Remove(sub); err != nil {
		t.Fatalf("Remove error = %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	// Recreate it and write a markdown file inside.
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatalf("Mkdir (recreate) error = %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	if err := os.WriteFile(filepath.Join(sub, "post.md"), []byte("hello"), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	waitForCount(t, &count, 1, 3*time.Second)
}

// TestRun_NoiseIgnored verifies that dot-files and editor swap files are ignored.
func TestRun_NoiseIgnored(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	w, err := watcher.New(dir, config.WithDebounce(shortDebounce))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	var count atomic.Int64
	go w.Run(ctx, func(context.Context) { count.Add(1) }) //nolint:errcheck

	time.Sleep(50 * time.Millisecond)

	noiseFiles := []string{".DS_Store", ".hidden", "post.md.swp", "post.md~", "post.md.swx"}
	for _, name := range noiseFiles {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("ignored"), 0o644); err != nil {
			t.Fatalf("WriteFile %q error = %v", name, err)
		}
	}

	// Wait longer than the debounce to confirm no callback fires.
	time.Sleep(shortDebounce + 200*time.Millisecond)

	if got := count.Load(); got != 0 {
		t.Errorf("onChange called %d time(s) for noise files, want 0", got)
	}
}

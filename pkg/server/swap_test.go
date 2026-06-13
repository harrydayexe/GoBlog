// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package server_test

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"testing/fstest"

	"github.com/harrydayexe/GoBlog/v2/pkg/config"
	"github.com/harrydayexe/GoBlog/v2/pkg/server"
)

// makeVersionFS returns a MapFS whose single post body contains marker so that
// the two versions produce fully distinct rendered bodies.
func makeVersionFS(marker string) fs.FS {
	return fstest.MapFS{
		"test-post.md": &fstest.MapFile{
			Data: []byte(fmt.Sprintf(
				"---\ntitle: Test Post\ndescription: A test post\ndate: 2024-01-01\n---\n\n# %s\n\nContent for %s.",
				marker, marker,
			)),
		},
	}
}

// badFS returns a MapFS whose only file has no frontmatter, causing
// parser.ParseFile to return "no frontmatter found in file" and therefore
// generator.Generate to return a non-nil error.
func badFS() fs.FS {
	return fstest.MapFS{
		"bad.md": &fstest.MapFile{
			Data: []byte("This file has no YAML frontmatter at all."),
		},
	}
}

// canonicalBody builds a throwaway server with the given FS, requests
// GET /posts/test-post, and returns the exact 200 body. It fatals if the
// server cannot be created or the response is not 200.
func canonicalBody(t *testing.T, postsFS fs.FS) string {
	t.Helper()

	cfg := config.ServerConfig{
		Server: []config.BaseServerOption{
			config.WithPort(8080),
		},
		Gen: []config.GeneratorOption{
			config.WithRawOutput(),
		},
	}

	srv, err := server.New(nil, postsFS, cfg)
	if err != nil {
		t.Fatalf("canonicalBody: server.New() error = %v", err)
	}

	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/posts/test-post", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("canonicalBody: GET /posts/test-post returned %d, want 200", rec.Code)
	}
	return rec.Body.String()
}

// TestUpdatePosts_ConcurrentSwapIsAtomic proves, under -race, that concurrent
// readers and writers on the same server only ever see a coherent (fully-old or
// fully-new) response body and never a 503.
//
// Acceptance criteria (from issue #68):
// Given ~N reader goroutines looping ServeHTTP and 1–2 writer goroutines
// looping UpdatePosts over a fixed iteration count, when the test runs under
// go test -race, then every response is 200 (never 503, never a panic) and
// every body is a single coherent version — old or new, never spliced.
func TestUpdatePosts_ConcurrentSwapIsAtomic(t *testing.T) {
	t.Parallel()

	const (
		numReaders = 8
		numWriters = 2
		iterations = 500
	)

	fsA := makeVersionFS("VERSION-A")
	fsB := makeVersionFS("VERSION-B")

	bodyA := canonicalBody(t, fsA)
	bodyB := canonicalBody(t, fsB)

	cfg := config.ServerConfig{
		Server: []config.BaseServerOption{
			config.WithPort(8080),
		},
		Gen: []config.GeneratorOption{
			config.WithRawOutput(),
		},
	}

	srv, err := server.New(nil, fsA, cfg)
	if err != nil {
		t.Fatalf("server.New() error = %v", err)
	}

	ctx := context.Background()

	var wg sync.WaitGroup

	// Reader goroutines: each issues ServeHTTP iterations times and asserts that
	// every response is 200 and its body is exactly bodyA or bodyB.
	for range numReaders {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("reader goroutine panicked: %v", r)
				}
			}()

			for range iterations {
				rec := httptest.NewRecorder()
				srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/posts/test-post", nil))

				if rec.Code != http.StatusOK {
					t.Errorf("reader: got status %d, want 200", rec.Code)
					return
				}

				body := rec.Body.String()
				if body != bodyA && body != bodyB {
					t.Errorf("reader: got unexpected body (len=%d); not VERSION-A or VERSION-B", len(body))
					return
				}
			}
		}()
	}

	// Writer goroutines: each alternates between UpdatePosts(fsA) and
	// UpdatePosts(fsB) for iterations cycles, asserting no error on each.
	for w := range numWriters {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("writer %d goroutine panicked: %v", id, r)
				}
			}()

			for i := range iterations {
				var fs fs.FS
				if i%2 == 0 {
					fs = fsA
				} else {
					fs = fsB
				}
				if err := srv.UpdatePosts(fs, ctx); err != nil {
					t.Errorf("writer %d: UpdatePosts() error = %v", id, err)
					return
				}
			}
		}(w)
	}

	wg.Wait()
}

// TestUpdatePosts_ReloadFailureKeepsOldContent proves that when UpdatePosts
// fails (e.g. due to malformed posts), the previously-serving handler remains
// active and old content continues to be served with a 200 response.
//
// Acceptance criteria (from issue #68):
// Given UpdatePosts is called with an fs.FS that fails generation/validation,
// when the reload fails, then a non-nil error is returned AND a subsequent
// request still serves the previous content with 200.
func TestUpdatePosts_ReloadFailureKeepsOldContent(t *testing.T) {
	t.Parallel()

	fsA := makeVersionFS("VERSION-A")
	fsB := makeVersionFS("VERSION-B")
	bodyA := canonicalBody(t, fsA)
	bodyB := canonicalBody(t, fsB)

	cfg := config.ServerConfig{
		Server: []config.BaseServerOption{
			config.WithPort(8080),
		},
		Gen: []config.GeneratorOption{
			config.WithRawOutput(),
		},
	}

	srv, err := server.New(nil, fsA, cfg)
	if err != nil {
		t.Fatalf("server.New() error = %v", err)
	}

	ctx := context.Background()

	// Baseline: server starts with version A.
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/posts/test-post", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("baseline: got status %d, want 200", rec.Code)
	}
	if got := rec.Body.String(); got != bodyA {
		t.Fatalf("baseline: body = %q, want VERSION-A body", got)
	}

	// Inject a bad FS — the reload must fail.
	err = srv.UpdatePosts(badFS(), ctx)
	if err == nil {
		t.Fatal("UpdatePosts(badFS): expected non-nil error, got nil")
	}

	// After the failure, the old handler must still serve the old content.
	rec2 := httptest.NewRecorder()
	srv.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/posts/test-post", nil))
	if rec2.Code != http.StatusOK {
		t.Errorf("post-failure: got status %d, want 200 (old handler must remain active)", rec2.Code)
	}
	if got := rec2.Body.String(); got != bodyA {
		t.Errorf("post-failure: body = %q, want VERSION-A body (old handler must remain active)", got)
	}

	// Recovery: a subsequent successful update must restore serving.
	if err := srv.UpdatePosts(fsB, ctx); err != nil {
		t.Fatalf("recovery UpdatePosts(fsB): error = %v", err)
	}

	rec3 := httptest.NewRecorder()
	srv.ServeHTTP(rec3, httptest.NewRequest(http.MethodGet, "/posts/test-post", nil))
	if rec3.Code != http.StatusOK {
		t.Errorf("post-recovery: got status %d, want 200", rec3.Code)
	}
	if got := rec3.Body.String(); got != bodyB {
		t.Errorf("post-recovery: body = %q, want VERSION-B body", got)
	}
}

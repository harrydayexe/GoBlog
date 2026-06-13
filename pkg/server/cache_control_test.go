// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package server_test

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/harrydayexe/GoBlog/v2/pkg/config"
	"github.com/harrydayexe/GoBlog/v2/pkg/server"
)

// TestCacheControl_Default verifies that the server sends "Cache-Control: public,
// max-age=3600" by default, with no explicit WithCacheControl option.
func TestCacheControl_Default(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	postsFS := createTestFS(t)

	cfg := config.ServerConfig{
		Gen: []config.GeneratorOption{config.WithRawOutput()},
	}

	srv, err := server.New(logger, postsFS, cfg)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	got := w.Header().Get("Cache-Control")
	want := "public, max-age=3600"
	if got != want {
		t.Errorf("Cache-Control: got %q, want %q", got, want)
	}
}

// TestCacheControl_CustomTTL verifies that WithCacheControl sets the correct
// max-age in whole seconds.
func TestCacheControl_CustomTTL(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	postsFS := createTestFS(t)

	cfg := config.ServerConfig{
		Server: []config.BaseServerOption{
			config.WithCacheControl(30 * time.Minute),
		},
		Gen: []config.GeneratorOption{config.WithRawOutput()},
	}

	srv, err := server.New(logger, postsFS, cfg)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	got := w.Header().Get("Cache-Control")
	want := "public, max-age=1800"
	if got != want {
		t.Errorf("Cache-Control: got %q, want %q", got, want)
	}
}

// TestCacheControl_ZeroDisablesHeader verifies that WithCacheControl(0) results
// in no Cache-Control header being sent.
func TestCacheControl_ZeroDisablesHeader(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	postsFS := createTestFS(t)

	cfg := config.ServerConfig{
		Server: []config.BaseServerOption{
			config.WithCacheControl(0),
		},
		Gen: []config.GeneratorOption{config.WithRawOutput()},
	}

	srv, err := server.New(logger, postsFS, cfg)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if got := w.Header().Get("Cache-Control"); got != "" {
		t.Errorf("expected no Cache-Control header, got %q", got)
	}
}

// TestCacheControl_PersistsAcrossUpdates verifies that the Cache-Control header
// is still present after UpdatePosts() regenerates the handler.
func TestCacheControl_PersistsAcrossUpdates(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	postsFS := createTestFS(t)

	cfg := config.ServerConfig{
		Server: []config.BaseServerOption{
			config.WithCacheControl(2 * time.Hour),
		},
		Gen: []config.GeneratorOption{config.WithRawOutput()},
	}

	srv, err := server.New(logger, postsFS, cfg)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	if err := srv.UpdatePosts(createTestFS(t), context.Background()); err != nil {
		t.Fatalf("UpdatePosts failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	got := w.Header().Get("Cache-Control")
	want := "public, max-age=7200"
	if got != want {
		t.Errorf("after UpdatePosts — Cache-Control: got %q, want %q", got, want)
	}
}

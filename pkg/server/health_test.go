// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package server_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/harrydayexe/GoBlog/v2/pkg/config"
	"github.com/harrydayexe/GoBlog/v2/pkg/server"
	"github.com/harrydayexe/GoWebUtilities/middleware"
)

// TestHealthChecks_Disabled verifies that the three /healthz/* endpoints are
// NOT intercepted when health checks are not enabled; they fall through to the
// content mux and return 404.
func TestHealthChecks_Disabled(t *testing.T) {
	t.Parallel()

	postsFS := createTestFS(t)
	cfg := config.ServerConfig{
		Server: []config.BaseServerOption{config.WithPort(8080)},
	}
	srv, err := server.New(nil, postsFS, cfg)
	if err != nil {
		t.Fatalf("server.New: %v", err)
	}

	for _, path := range []string{"/healthz/live", "/healthz/ready", "/healthz/startup"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()
		srv.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("health checks disabled: GET %s → %d, want 404", path, w.Code)
		}
	}
}

// TestHealthChecks_Live_AlwaysOK verifies that /healthz/live returns 200 OK
// even while the server is still in the starting state (before Run is called).
func TestHealthChecks_Live_AlwaysOK(t *testing.T) {
	t.Parallel()

	postsFS := createTestFS(t)
	cfg := config.ServerConfig{
		Server: []config.BaseServerOption{
			config.WithPort(8080),
			config.WithHealthChecks(),
		},
	}
	// New returns immediately with state=starting when health checks are on.
	srv, err := server.New(nil, postsFS, cfg)
	if err != nil {
		t.Fatalf("server.New: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/healthz/live", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("/healthz/live: got %d, want 200", w.Code)
	}
	if got := strings.TrimSpace(w.Body.String()); got != "ok" {
		t.Errorf("/healthz/live body: got %q, want \"ok\"", got)
	}
}

// TestHealthChecks_Ready_WhileStarting verifies that /healthz/ready and
// /healthz/startup return 503 before content has been loaded.
func TestHealthChecks_Ready_WhileStarting(t *testing.T) {
	t.Parallel()

	postsFS := createTestFS(t)
	cfg := config.ServerConfig{
		Server: []config.BaseServerOption{
			config.WithPort(8080),
			config.WithHealthChecks(),
		},
	}
	// State is "starting" immediately after New (before Run).
	srv, err := server.New(nil, postsFS, cfg)
	if err != nil {
		t.Fatalf("server.New: %v", err)
	}

	for _, path := range []string{"/healthz/ready", "/healthz/startup"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()
		srv.ServeHTTP(w, req)

		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("GET %s while starting: got %d, want 503", path, w.Code)
		}
		body := strings.TrimSpace(w.Body.String())
		if body == "" {
			t.Errorf("GET %s while starting: body is empty, want a reason string", path)
		}
	}
}

// TestHealthChecks_NonGetMethod verifies that all three endpoints return
// 405 Method Not Allowed for non-GET methods.
func TestHealthChecks_NonGetMethod(t *testing.T) {
	t.Parallel()

	postsFS := createTestFS(t)
	cfg := config.ServerConfig{
		Server: []config.BaseServerOption{
			config.WithPort(8080),
			config.WithHealthChecks(),
		},
	}
	srv, err := server.New(nil, postsFS, cfg)
	if err != nil {
		t.Fatalf("server.New: %v", err)
	}

	methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete}
	paths := []string{"/healthz/live", "/healthz/ready", "/healthz/startup"}

	for _, method := range methods {
		for _, path := range paths {
			req := httptest.NewRequest(method, path, nil)
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("%s %s: got %d, want 405", method, path, w.Code)
			}
		}
	}
}

// TestHealthChecks_BypassesMiddleware verifies that /healthz/live returns
// 200 OK even when a middleware that rejects all requests is configured.
// This confirms that health endpoints are served before the middleware stack.
func TestHealthChecks_BypassesMiddleware(t *testing.T) {
	t.Parallel()

	rejectAll := middleware.Middleware(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "forbidden", http.StatusForbidden)
		})
	})

	postsFS := createTestFS(t)
	cfg := config.ServerConfig{
		Server: []config.BaseServerOption{
			config.WithPort(8080),
			config.WithMiddleware(rejectAll),
			config.WithHealthChecks(),
		},
	}
	// With health checks disabled the sync path is used (middleware is applied
	// but health routes are not intercepted). Enable health checks to get the
	// pre-middleware interception.
	srv, err := server.New(nil, postsFS, cfg)
	if err != nil {
		t.Fatalf("server.New: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/healthz/live", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("/healthz/live with rejecting middleware: got %d, want 200", w.Code)
	}
}

// TestHealthChecks_ReadyAfterInit verifies that /healthz/ready returns 200 OK
// after Run has completed content initialisation. This requires a real HTTP
// listener so we poll until the probe responds 200 or a timeout is reached.
func TestHealthChecks_ReadyAfterInit(t *testing.T) {
	t.Parallel()

	postsFS := createTestFS(t)

	// Grab a free port.
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen: %v", err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()

	cfg := config.ServerConfig{
		Server: []config.BaseServerOption{
			config.WithPort(port),
			config.WithHost("127.0.0.1"),
			config.WithHealthChecks(),
		},
	}
	srv, err := server.New(nil, postsFS, cfg)
	if err != nil {
		t.Fatalf("server.New: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	runErr := make(chan error, 1)
	go func() { runErr <- srv.Run(ctx) }()

	addr := fmt.Sprintf("http://127.0.0.1:%d", port)

	// Poll /healthz/ready until 200 or timeout.
	deadline := time.Now().Add(8 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get(addr + "/healthz/ready") //nolint:noctx
		if err == nil {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				if got := strings.TrimSpace(string(body)); got != "ok" {
					t.Errorf("/healthz/ready body: got %q, want \"ok\"", got)
				}
				cancel() // graceful shutdown
				return
			}
		}
		time.Sleep(25 * time.Millisecond)
	}

	t.Error("/healthz/ready did not return 200 within the deadline")
	cancel()
}

// TestHealthChecks_FailedInit verifies that /healthz/ready returns 503 with a
// non-empty reason when content initialisation fails (e.g. unreadable FS).
func TestHealthChecks_FailedInit(t *testing.T) {
	t.Parallel()

	// errFS always returns an error on Open, causing generator.Generate to fail.
	brokenFS := fstest.MapFS{
		"broken.md": &fstest.MapFile{
			// Frontmatter is deliberately malformed to trigger a parse error.
			Data: []byte("---\nbad yaml: [unclosed\n---\n# hello\n"),
		},
	}

	// Grab a free port.
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen: %v", err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()

	cfg := config.ServerConfig{
		Server: []config.BaseServerOption{
			config.WithPort(port),
			config.WithHost("127.0.0.1"),
			config.WithHealthChecks(),
		},
	}
	srv, err := server.New(nil, brokenFS, cfg)
	if err != nil {
		t.Fatalf("server.New: %v (want nil when health checks enabled)", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	go func() { srv.Run(ctx) }() //nolint:errcheck

	addr := fmt.Sprintf("http://127.0.0.1:%d", port)

	// Poll until we get either a 503 (failed) or timeout.
	deadline := time.Now().Add(8 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get(addr + "/healthz/ready") //nolint:noctx
		if err != nil {
			time.Sleep(25 * time.Millisecond)
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		switch resp.StatusCode {
		case http.StatusServiceUnavailable:
			// Good — check that a reason is included.
			bodyStr := strings.TrimSpace(string(body))
			if bodyStr == "" || bodyStr == "starting" {
				// Still starting; keep polling.
				time.Sleep(25 * time.Millisecond)
				continue
			}
			// Got a failure reason. Test passes.
			cancel()
			return
		case http.StatusOK:
			t.Error("/healthz/ready returned 200 even with a broken FS")
			cancel()
			return
		}
		time.Sleep(25 * time.Millisecond)
	}

	t.Error("/healthz/ready did not return 503 with a failure reason within the deadline")
	cancel()
}

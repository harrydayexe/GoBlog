// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package integration_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/harrydayexe/GoBlog/v2/pkg/config"
	"github.com/harrydayexe/GoBlog/v2/pkg/server"
)

// TestRun_BindError verifies that Server.Run surfaces a bind error when the
// configured port is already occupied, rather than silently swallowing the
// listenErr (pkg/server/server.go, the ListenAndServe goroutine).
func TestRun_BindError(t *testing.T) {
	// Occupy a port to force the bind conflict.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen: %v", err)
	}
	defer listener.Close()
	port := listener.Addr().(*net.TCPAddr).Port

	dir := t.TempDir()
	writePost(t, dir, "post.md", minimalPost("Hello World"))

	cfg := config.ServerConfig{
		Server: []config.BaseServerOption{
			config.WithPort(port),
			config.WithHost("127.0.0.1"),
		},
	}
	srv, err := server.New(nil, os.DirFS(dir), cfg)
	if err != nil {
		t.Fatalf("server.New: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Run(ctx); err == nil {
		t.Fatal("expected a bind error from Run, got nil")
	}
}

// TestRun_GracefulShutdown verifies that cancelling the context causes Run to
// return nil and complete well within the 10 s configured shutdown window.
func TestRun_GracefulShutdown(t *testing.T) {
	dir := t.TempDir()
	writePost(t, dir, "post.md", minimalPost("Hello World"))

	// Grab a free port by binding, recording it, then releasing it.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	cfg := config.ServerConfig{
		Server: []config.BaseServerOption{
			config.WithPort(port),
			config.WithHost("127.0.0.1"),
		},
	}
	srv, err := server.New(nil, os.DirFS(dir), cfg)
	if err != nil {
		t.Fatalf("server.New: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() { done <- srv.Run(ctx) }()

	// Wait until the server is ready to serve requests.
	addr := fmt.Sprintf("http://127.0.0.1:%d/", port)
	eventually(t, 5*time.Second, 50*time.Millisecond, func() bool {
		//nolint:gosec // test-controlled URL
		resp, err := http.Get(addr)
		if err != nil {
			return false
		}
		resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	})

	// Cancel the context and measure how long shutdown takes.
	start := time.Now()
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run returned unexpected error: %v", err)
		}
		if elapsed := time.Since(start); elapsed > 5*time.Second {
			t.Errorf("shutdown took %s; want < 5 s", elapsed)
		}
	case <-time.After(15 * time.Second):
		t.Fatal("server did not shut down within 15 s")
	}
}

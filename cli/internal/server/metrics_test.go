// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package server

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/harrydayexe/GoBlog/v2/cli/internal/cliflags"
	"github.com/harrydayexe/GoBlog/v2/pkg/config"
	pkgserver "github.com/harrydayexe/GoBlog/v2/pkg/server"
	"github.com/urfave/cli/v3"
)

// startMetricsServer binds an admin listener on an ephemeral localhost port and
// serves it for the duration of the test.
func startMetricsServer(t *testing.T) *metricsServer {
	t.Helper()

	metrics, err := newMetricsServer("127.0.0.1", 0)
	if err != nil {
		t.Fatalf("newMetricsServer() error = %v", err)
	}
	go func() {
		if err := metrics.Serve(); err != nil {
			t.Errorf("metrics.Serve() error = %v", err)
		}
	}()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := metrics.Shutdown(ctx); err != nil {
			t.Errorf("metrics.Shutdown() error = %v", err)
		}
	})

	return metrics
}

// scrape performs a GET against the admin listener and returns the status and
// body, as Prometheus would.
func scrape(t *testing.T, metrics *metricsServer, path string) (int, string) {
	t.Helper()

	resp, err := http.Get("http://" + metrics.Addr() + path) //nolint:gosec // address is test-controlled
	if err != nil {
		t.Fatalf("GET %s error = %v", path, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("reading %s body error = %v", path, err)
	}
	return resp.StatusCode, string(body)
}

// occupyPort binds a localhost port and keeps it bound for the test, so a
// second bind attempt on it is guaranteed to fail.
func occupyPort(t *testing.T) int {
	t.Helper()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() error = %v", err)
	}
	t.Cleanup(func() { _ = l.Close() })

	return mustPort(t, l.Addr().String())
}

// runServeCLIContext runs the serve subcommand through a root command shaped
// like the real goblog binary, cancelling on ctx.
//
// These tests do not run in parallel: urfave/cli stores parsed values on the
// flag definitions, which the package-level ServeCommand shares.
func runServeCLIContext(t *testing.T, ctx context.Context, args ...string) error {
	t.Helper()

	cmd := &cli.Command{
		Name:                   "goblog",
		UseShortOptionHandling: true,
		Commands:               []*cli.Command{&ServeCommand},
		Flags:                  cliflags.Shared(),
	}
	return cmd.Run(ctx, append([]string{"goblog", "serve"}, args...))
}

// TestMetrics_ServesSemanticConventionNames verifies that traffic through the
// instrumented blog handler shows up on the admin listener under the OTel HTTP
// server semantic convention names Prometheus rules expect.
func TestMetrics_ServesSemanticConventionNames(t *testing.T) {
	t.Parallel()

	metrics := startMetricsServer(t)

	srv, err := pkgserver.New(testFS(),
		config.WithPort(0),
		config.WithMeterProvider(metrics.MeterProvider()),
	)
	if err != nil {
		t.Fatalf("server.New() error = %v", err)
	}

	for range 3 {
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("GET / status = %d, want %d", rec.Code, http.StatusOK)
		}
	}

	status, body := scrape(t, metrics, metricsPath)
	if status != http.StatusOK {
		t.Fatalf("GET %s status = %d, want %d", metricsPath, status, http.StatusOK)
	}

	for _, want := range []string{
		"http_server_request_duration_seconds_count",
		"http_server_request_duration_seconds_bucket",
		"http_server_response_body_size_bytes_count",
		// The index route is registered as the exact-match mux pattern "/{$}".
		`http_route="/{$}"`,
		`http_request_method="GET"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("scrape does not contain %q:\n%s", want, body)
		}
	}

	// The three requests above must be reflected in the counter.
	if want := `http_server_request_duration_seconds_count{`; !strings.Contains(body, want) {
		t.Fatalf("scrape has no request duration count:\n%s", body)
	}
	if !strings.Contains(body, "} 3") {
		t.Errorf("scrape does not report the 3 requests made:\n%s", body)
	}
}

// TestMetrics_AdminListenerServesOnlyMetrics verifies that the admin listener
// is not a second copy of the blog: only /metrics answers on it.
func TestMetrics_AdminListenerServesOnlyMetrics(t *testing.T) {
	t.Parallel()

	metrics := startMetricsServer(t)

	for _, path := range []string{"/", "/posts/test-post", "/metrics/", "/healthz/live"} {
		if status, _ := scrape(t, metrics, path); status != http.StatusNotFound {
			t.Errorf("GET %s on the admin listener status = %d, want %d", path, status, http.StatusNotFound)
		}
	}
}

// TestMetrics_NotOnBlogListener verifies that /metrics is never reachable on
// the blog's public listener, with or without a blog root.
func TestMetrics_NotOnBlogListener(t *testing.T) {
	t.Parallel()

	metrics := startMetricsServer(t)

	srv, err := pkgserver.New(testFS(),
		config.WithPort(0),
		config.WithBlogRoot("/blog/").AsServerOption(),
		config.WithMeterProvider(metrics.MeterProvider()),
	)
	if err != nil {
		t.Fatalf("server.New() error = %v", err)
	}

	for _, path := range []string{"/metrics", "/blog/metrics"} {
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusNotFound {
			t.Errorf("GET %s on the blog listener status = %d, want %d", path, rec.Code, http.StatusNotFound)
		}
	}
}

// TestMetrics_BindFailureIsFatal verifies that a clash on the admin port fails
// the serve command outright, naming the address, rather than serving a blog
// with metrics silently missing.
func TestMetrics_BindFailureIsFatal(t *testing.T) {
	port := occupyPort(t)

	err := runServeCLIContext(t, context.Background(),
		"--"+MetricsFlagName,
		"--"+MetricsHostFlagName, "127.0.0.1",
		"--"+MetricsPortFlagName, strconv.Itoa(port),
		"--"+PortFlagName, "0",
		t.TempDir(),
	)

	if err == nil {
		t.Fatal("serve with an occupied metrics port returned nil, want a bind error")
	}
	if want := fmt.Sprintf("127.0.0.1:%d", port); !strings.Contains(err.Error(), want) {
		t.Errorf("error %q does not name the address %q", err, want)
	}
}

// TestMetrics_NoAdminPortWhenDisabled verifies that without --metrics no admin
// listener is bound at all: the serve command starts cleanly even when the
// metrics port is already taken.
func TestMetrics_NoAdminPortWhenDisabled(t *testing.T) {
	port := occupyPort(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := runServeCLIContext(t, ctx,
		"--"+MetricsHostFlagName, "127.0.0.1",
		"--"+MetricsPortFlagName, strconv.Itoa(port),
		"--"+PortFlagName, "0",
		t.TempDir(),
	)

	if err != nil {
		t.Errorf("serve without --metrics error = %v, want nil (no admin listener should be bound)", err)
	}
}

// TestMetrics_ShutdownReleasesListener verifies that both listeners are torn
// down when the server stops: the admin port is free to bind again once
// runServe returns.
func TestMetrics_ShutdownReleasesListener(t *testing.T) {
	t.Parallel()

	metrics, err := newMetricsServer("127.0.0.1", 0)
	if err != nil {
		t.Fatalf("newMetricsServer() error = %v", err)
	}
	addr := metrics.Addr()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := runServe(ctx, t.TempDir(), testFS(), false, metrics,
		config.WithPort(0),
		config.WithMeterProvider(metrics.MeterProvider()),
	); err != nil {
		t.Fatalf("runServe() error = %v", err)
	}

	l, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatalf("admin port %s still bound after shutdown: %v", addr, err)
	}
	_ = l.Close()
}

// TestMetrics_GracefulShutdownWithBothListeners verifies that cancelling the
// context while both listeners are live stops them both: the blog port and the
// admin port are free to bind again once runServe returns.
func TestMetrics_GracefulShutdownWithBothListeners(t *testing.T) {
	t.Parallel()

	blogAddr := net.JoinHostPort("127.0.0.1", strconv.Itoa(freePort(t)))

	metrics, err := newMetricsServer("127.0.0.1", 0)
	if err != nil {
		t.Fatalf("newMetricsServer() error = %v", err)
	}
	metricsAddr := metrics.Addr()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- runServe(ctx, t.TempDir(), testFS(), false, metrics,
			config.WithHost("127.0.0.1"),
			config.WithPort(mustPort(t, blogAddr)),
			config.WithMeterProvider(metrics.MeterProvider()),
		)
	}()

	// Both listeners must be answering before the shutdown is meaningful.
	for _, url := range []string{"http://" + blogAddr + "/", "http://" + metricsAddr + metricsPath} {
		waitForStatus(t, url, http.StatusOK)
	}

	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("runServe() error = %v", err)
		}
	case <-time.After(15 * time.Second):
		t.Fatal("runServe() did not return within 15s of cancellation")
	}

	for _, addr := range []string{blogAddr, metricsAddr} {
		l, err := net.Listen("tcp", addr)
		if err != nil {
			t.Errorf("port %s still bound after shutdown: %v", addr, err)
			continue
		}
		_ = l.Close()
	}
}

// TestMetrics_ListenerReleasedWhenServerFails verifies that a failure to build
// the blog server does not leave the admin listener bound behind it.
func TestMetrics_ListenerReleasedWhenServerFails(t *testing.T) {
	t.Parallel()

	metrics, err := newMetricsServer("127.0.0.1", 0)
	if err != nil {
		t.Fatalf("newMetricsServer() error = %v", err)
	}
	addr := metrics.Addr()

	// An empty template filesystem fails the renderer, so server.New errors.
	err = runServe(context.Background(), t.TempDir(), testFS(), false, metrics,
		config.WithPort(0),
		config.WithTemplateDir(fstest.MapFS{}),
		config.WithMeterProvider(metrics.MeterProvider()),
	)
	if err == nil {
		t.Fatal("runServe() with an empty template dir returned nil, want an error")
	}

	l, listenErr := net.Listen("tcp", addr)
	if listenErr != nil {
		t.Fatalf("admin port %s still bound after runServe failed: %v", addr, listenErr)
	}
	_ = l.Close()
}

// freePort returns a port that was free a moment ago, for cases where the
// listener has to be created by the code under test rather than the test.
func freePort(t *testing.T) int {
	t.Helper()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() error = %v", err)
	}
	port := mustPort(t, l.Addr().String())
	if err := l.Close(); err != nil {
		t.Fatalf("closing probe listener error = %v", err)
	}
	return port
}

// mustPort extracts the port number from a "host:port" address.
func mustPort(t *testing.T, addr string) int {
	t.Helper()

	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("net.SplitHostPort(%q) error = %v", addr, err)
	}
	n, err := strconv.Atoi(port)
	if err != nil {
		t.Fatalf("strconv.Atoi(%q) error = %v", port, err)
	}
	return n
}

// waitForStatus polls url until it answers with want, failing the test if it
// does not do so within 10 seconds.
func waitForStatus(t *testing.T, url string, want int) {
	t.Helper()

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get(url) //nolint:gosec // address is test-controlled
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == want {
				return
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("GET %s did not return %d within 10s", url, want)
}

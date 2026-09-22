// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package integration_test

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// startMetricsContainer starts a goblog container with both the blog port and
// the admin port published, and returns the "host:port" address of each.
//
// entrypoint overrides the image ENTRYPOINT when non-nil, which is how a test
// runs the binary without the image's default --metrics.
func startMetricsContainer(t *testing.T, ctx context.Context, postsDir string, entrypoint []string) (testcontainers.Container, string, string) {
	t.Helper()

	req := testcontainers.ContainerRequest{
		Image:        imageTag,
		Entrypoint:   entrypoint,
		Cmd:          []string{"/posts"},
		ExposedPorts: []string{"8080/tcp", "9090/tcp"},
		WaitingFor:   wait.ForListeningPort("8080/tcp").WithStartupTimeout(60 * time.Second),
		Mounts: testcontainers.ContainerMounts{
			{
				Source: testcontainers.GenericBindMountSource{HostPath: postsDir},
				Target: testcontainers.ContainerMountTarget("/posts"),
			},
		},
	}

	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("startMetricsContainer: %v", err)
	}

	host, err := c.Host(ctx)
	if err != nil {
		_ = c.Terminate(ctx)
		t.Fatalf("container host: %v", err)
	}

	blogPort, err := c.MappedPort(ctx, "8080")
	if err != nil {
		_ = c.Terminate(ctx)
		t.Fatalf("container port 8080: %v", err)
	}
	metricsPort, err := c.MappedPort(ctx, "9090")
	if err != nil {
		_ = c.Terminate(ctx)
		t.Fatalf("container port 9090: %v", err)
	}

	return c, fmt.Sprintf("%s:%s", host, blogPort.Port()), fmt.Sprintf("%s:%s", host, metricsPort.Port())
}

// metricValue returns the value of the first sample of metric whose label set
// contains every want substring, or -1 when no such sample is present.
func metricValue(scrape, metric string, want ...string) float64 {
	for line := range strings.Lines(scrape) {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, metric+"{") {
			continue
		}
		matched := true
		for _, w := range want {
			if !strings.Contains(line, w) {
				matched = false
				break
			}
		}
		if !matched {
			continue
		}
		value, err := strconv.ParseFloat(line[strings.LastIndex(line, " ")+1:], 64)
		if err != nil {
			continue
		}
		return value
	}
	return -1
}

// TestServe_Metrics boots the Docker image, drives traffic at the blog port and
// asserts that the admin port reports it under the OpenTelemetry HTTP server
// semantic convention names, while the blog port never serves /metrics.
func TestServe_Metrics(t *testing.T) {
	skipIfNoDocker(t)
	ctx := context.Background()

	dir := t.TempDir()
	writePost(t, dir, "post.md", minimalPost("Metrics Post"))

	c, blogAddr, metricsAddr := startMetricsContainer(t, ctx, dir, nil)
	defer func() { _ = c.Terminate(ctx) }()

	eventually(t, 15*time.Second, 500*time.Millisecond, func() bool {
		status, _ := httpGet(t, fmt.Sprintf("http://%s/", blogAddr))
		return status == http.StatusOK
	})

	// Generate a known amount of traffic on one route.
	const requests = 5
	for range requests {
		if status, _ := httpGet(t, fmt.Sprintf("http://%s/posts/metrics-post", blogAddr)); status != http.StatusOK {
			t.Fatalf("GET /posts/metrics-post status = %d, want %d", status, http.StatusOK)
		}
	}

	var scrape string
	eventually(t, 15*time.Second, 500*time.Millisecond, func() bool {
		status, body := httpGet(t, fmt.Sprintf("http://%s/metrics", metricsAddr))
		scrape = body
		return status == http.StatusOK
	})

	const counter = "http_server_request_duration_seconds_count"
	if !strings.Contains(scrape, counter) {
		t.Fatalf("scrape does not contain %s:\n%s", counter, scrape)
	}

	got := metricValue(scrape, counter, `http_route="/posts/{postName}"`, `http_response_status_code="200"`)
	if got < requests {
		t.Errorf("%s for the post route = %v, want at least %d\n%s", counter, got, requests, scrape)
	}

	// The same data must never be reachable on the blog's public port.
	if status, _ := httpGet(t, fmt.Sprintf("http://%s/metrics", blogAddr)); status != http.StatusNotFound {
		t.Errorf("GET /metrics on the blog port status = %d, want %d", status, http.StatusNotFound)
	}
}

// TestServe_MetricsDisabled overrides the image ENTRYPOINT to drop --metrics
// and asserts that nothing is listening on the admin port inside the container.
func TestServe_MetricsDisabled(t *testing.T) {
	skipIfNoDocker(t)
	ctx := context.Background()

	dir := t.TempDir()
	writePost(t, dir, "post.md", minimalPost("No Metrics Post"))

	c, blogAddr, _ := startMetricsContainer(t, ctx, dir, []string{"./goblog", "serve", "--health-checks"})
	defer func() { _ = c.Terminate(ctx) }()

	eventually(t, 15*time.Second, 500*time.Millisecond, func() bool {
		status, _ := httpGet(t, fmt.Sprintf("http://%s/", blogAddr))
		return status == http.StatusOK
	})

	// Probing from inside the container avoids relying on the published port:
	// with no listener there is nothing for Docker to forward to.
	code, _, err := c.Exec(ctx, []string{"wget", "--spider", "-q", "-T", "2", "http://localhost:9090/metrics"})
	if err != nil {
		t.Fatalf("exec wget: %v", err)
	}
	if code == 0 {
		t.Error("something answered on the admin port without --metrics, want no listener")
	}
}

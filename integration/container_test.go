// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package integration_test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// startContainer starts a goblog container using the pre-built test image.
//
// When postsDir is non-nil the directory is bind-mounted into the container
// at /posts. cmd overrides the image's default CMD — the ENTRYPOINT
// (./goblog serve) is preserved, so cmd is the argument list that follows it.
// Passing nil cmd uses the Dockerfile default (CMD ["/posts"]).
//
// Returns the running container and the "host:port" address of the mapped
// port 8080.
func startContainer(t *testing.T, ctx context.Context, postsDir *string, cmd []string) (testcontainers.Container, string) {
	t.Helper()

	req := testcontainers.ContainerRequest{
		Image:        imageTag,
		Cmd:          cmd,
		ExposedPorts: []string{"8080/tcp"},
		WaitingFor:   wait.ForListeningPort("8080/tcp").WithStartupTimeout(60 * time.Second),
	}

	if postsDir != nil {
		req.Mounts = testcontainers.ContainerMounts{
			{
				Source: testcontainers.GenericBindMountSource{HostPath: *postsDir},
				Target: testcontainers.ContainerMountTarget("/posts"),
			},
		}
	}

	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("startContainer: %v", err)
	}

	host, err := c.Host(ctx)
	if err != nil {
		_ = c.Terminate(ctx)
		t.Fatalf("container host: %v", err)
	}
	port, err := c.MappedPort(ctx, "8080")
	if err != nil {
		_ = c.Terminate(ctx)
		t.Fatalf("container port: %v", err)
	}

	return c, fmt.Sprintf("%s:%s", host, port.Port())
}

// TestServe_Smoke boots the Docker image against the built-in empty /posts
// directory and asserts that the index page is served with HTTP 200.
// This exercises the Docker distribution channel end-to-end.
func TestServe_Smoke(t *testing.T) {
	skipIfNoDocker(t)
	ctx := context.Background()

	c, addr := startContainer(t, ctx, nil, nil)
	defer func() { _ = c.Terminate(ctx) }()

	eventually(t, 5*time.Second, 200*time.Millisecond, func() bool {
		status, _ := httpGet(t, fmt.Sprintf("http://%s/", addr))
		return status == http.StatusOK
	})
}

// TestServe_LiveReload verifies end-to-end live reload: after writing a new
// markdown file to the bind-mounted posts directory the served index page
// reflects the change, confirming that the watcher → server pipeline works
// through the full stack.
func TestServe_LiveReload(t *testing.T) {
	skipIfNoDocker(t)
	ctx := context.Background()

	dir := t.TempDir()
	writePost(t, dir, "initial.md", minimalPost("Initial Post"))

	// -w enables the file watcher; /posts is bind-mounted from dir.
	c, addr := startContainer(t, ctx, &dir, []string{"-w", "/posts"})
	defer func() { _ = c.Terminate(ctx) }()

	// Wait for the initial post to appear in the served index.
	eventually(t, 10*time.Second, 500*time.Millisecond, func() bool {
		_, body := httpGet(t, fmt.Sprintf("http://%s/", addr))
		return strings.Contains(body, "Initial Post")
	})

	// Write a new post on the host and assert the live server reflects it.
	writePost(t, dir, "new-post.md", minimalPost("New Post After Reload"))

	eventually(t, 15*time.Second, 500*time.Millisecond, func() bool {
		_, body := httpGet(t, fmt.Sprintf("http://%s/", addr))
		return strings.Contains(body, "New Post After Reload")
	})
}

// TestServe_CacheControl boots the server with default flags and asserts that
// every response includes a non-empty Cache-Control header, confirming the
// default 1-hour TTL is applied end-to-end through the Docker distribution path.
func TestServe_CacheControl(t *testing.T) {
	skipIfNoDocker(t)
	ctx := context.Background()

	c, addr := startContainer(t, ctx, nil, nil)
	defer func() { _ = c.Terminate(ctx) }()

	var cacheHeader string
	eventually(t, 10*time.Second, 200*time.Millisecond, func() bool {
		status, h := httpGetHeader(t, fmt.Sprintf("http://%s/", addr), "Cache-Control")
		cacheHeader = h
		return status == http.StatusOK
	})

	if cacheHeader == "" {
		t.Error("expected a non-empty Cache-Control header, got none")
	}
}

// TestServe_BlogRootFlag boots the server with -p /blog/ and verifies that
// the served HTML contains /blog/-prefixed links, confirming that the CLI flag
// is correctly wired through to the running process.
func TestServe_BlogRootFlag(t *testing.T) {
	skipIfNoDocker(t)
	ctx := context.Background()

	dir := t.TempDir()
	writePost(t, dir, "post.md", minimalPost("Blog Root Post"))

	// -p /blog/ sets the blog root; /posts is the positional posts-dir argument.
	c, addr := startContainer(t, ctx, &dir, []string{"-p", "/blog/", "/posts"})
	defer func() { _ = c.Terminate(ctx) }()

	// With -p /blog/ the index is served at /blog/, not /.
	var body string
	eventually(t, 10*time.Second, 500*time.Millisecond, func() bool {
		status, b := httpGet(t, fmt.Sprintf("http://%s/blog/", addr))
		body = b
		return status == http.StatusOK
	})

	if !strings.Contains(body, "/blog/") {
		t.Errorf("expected response body to contain /blog/ links\nbody:\n%s", body)
	}
}

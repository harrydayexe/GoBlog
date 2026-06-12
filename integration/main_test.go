// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package integration_test contains black-box integration tests for GoBlog.
//
// Tests in this package require Docker to be running for the container-based
// tests. The in-process lifecycle tests (TestRun_*) run without Docker.
//
// Run with:
//
//	cd integration && go test -v ./...
package integration_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
)

// imageTag is the Docker image reference built once in TestMain and reused by
// all container tests. Empty string means Docker was unavailable.
var imageTag string

// dockerSkip is set to true when Docker is not reachable on this host.
var dockerSkip bool

// TestMain builds the goblog Docker image once before any tests run so the
// heavy Dockerfile build is paid only once per test binary invocation.
// Container tests are skipped gracefully when Docker is unavailable.
func TestMain(m *testing.M) {
	os.Exit(run(m))
}

// run is the real body of TestMain. It is extracted so that deferred
// cleanup (notably cancel()) fires before os.Exit is called — os.Exit
// bypasses deferred functions in the calling frame.
func run(m *testing.M) int {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// First check whether Docker is reachable at all. A Health ping avoids
	// the ambiguity of treating a Dockerfile build failure as "Docker down".
	if !dockerHealthy(ctx) {
		fmt.Fprintln(os.Stderr, "integration: Docker unavailable, container tests will be skipped")
		dockerSkip = true
		return m.Run()
	}

	tag, err := buildTestImage(ctx)
	if err != nil {
		// Docker is up, so this is a genuine image-build failure (e.g. compile
		// error, missing COPY target). Fail loudly rather than silently skipping.
		fmt.Fprintf(os.Stderr, "integration: image build failed: %v\n", err)
		return 1
	}
	imageTag = tag
	return m.Run()
}

// dockerHealthy reports whether a Docker daemon is reachable by performing a
// lightweight Health ping, without building or starting any container.
func dockerHealthy(ctx context.Context) bool {
	provider, err := testcontainers.NewDockerProvider()
	if err != nil {
		return false
	}
	defer provider.Close()
	return provider.Health(ctx) == nil
}

// buildTestImage builds the goblog Docker image from the repository Dockerfile.
// KeepImage: true ensures the built image persists between test runs so Docker's
// layer cache is used on subsequent invocations.
func buildTestImage(ctx context.Context) (string, error) {
	const (
		repo = "goblog-integration-test"
		tag  = "latest"
	)

	// Creating the container with Started: false builds the image without
	// starting a container, warming the Docker cache for all subsequent tests.
	// Repo and Tag are separate fields in testcontainers-go v0.37+; combining
	// them into Tag alone produces an invalid reference.
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			FromDockerfile: testcontainers.FromDockerfile{
				Context:    "..",
				Dockerfile: "Dockerfile",
				KeepImage:  true,
				Repo:       repo,
				Tag:        tag,
			},
		},
		Started: false,
	})
	if err != nil {
		return "", err
	}
	if c != nil {
		_ = c.Terminate(ctx)
	}
	return repo + ":" + tag, nil
}

// skipIfNoDocker skips the calling test when Docker is not available.
func skipIfNoDocker(t *testing.T) {
	t.Helper()
	if dockerSkip {
		t.Skip("Docker not available on this host")
	}
}

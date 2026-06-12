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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	tag, err := buildTestImage(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "integration: Docker unavailable, container tests will be skipped: %v\n", err)
		dockerSkip = true
	} else {
		imageTag = tag
	}

	os.Exit(m.Run())
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

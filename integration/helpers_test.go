// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package integration_test

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// minimalPost returns a valid markdown blog post body with the given title.
func minimalPost(title string) string {
	return fmt.Sprintf(`---
title: %q
date: 2026-01-01T00:00:00Z
description: "A test post for integration testing"
---

# %s

This is a test post for integration testing.
`, title, title)
}

// writePost writes a markdown file with the given name and body to dir.
func writePost(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0600); err != nil {
		t.Fatalf("writePost %s: %v", name, err)
	}
}

// eventually polls fn every interval until it returns true or timeout elapses.
// The test is failed if fn does not return true within the timeout.
func eventually(t *testing.T, timeout, interval time.Duration, fn func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if fn() {
			return
		}
		time.Sleep(interval)
	}
	t.Fatalf("eventually: condition not met within %s", timeout)
}

// httpGet performs a GET request and returns the status code and response body.
// Network errors are returned as (0, "") so callers can handle them uniformly
// inside eventually loops.
func httpGet(t *testing.T, url string) (int, string) {
	t.Helper()
	//nolint:gosec // test helper — URL is always test-controlled
	resp, err := http.Get(url)
	if err != nil {
		return 0, ""
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, ""
	}
	return resp.StatusCode, string(body)
}

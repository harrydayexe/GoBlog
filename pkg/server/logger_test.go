// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package server

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/harrydayexe/GoBlog/v2/pkg/config"
)

func makeInternalTestFS(t *testing.T) fstest.MapFS {
	t.Helper()
	return fstest.MapFS{
		"test-post.md": &fstest.MapFile{
			Data: []byte(strings.TrimSpace(`
---
title: Test Post
description: A test post
date: 2024-01-01
---

# Test Content
`)),
		},
	}
}

// TestServer_LoggerPropagatedToGenerator verifies that the server's resolved
// logger is forwarded into the internal generator it constructs.
func TestServer_LoggerPropagatedToGenerator(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	injected := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	postsFS := makeInternalTestFS(t)
	cfg := config.ServerConfig{
		Server: []config.BaseServerOption{
			{BaseOption: config.WithLogger(injected)},
		},
		Gen: []config.GeneratorOption{config.WithRawOutput()},
	}

	srv, err := New(nil, postsFS, cfg)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	if srv.generator.Logger.Logger != injected {
		t.Error("expected server to propagate its logger into the internal generator")
	}
}

// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package generator

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/harrydayexe/GoBlog/v2/cli/internal/cliflags"
	inerrors "github.com/harrydayexe/GoBlog/v2/cli/internal/errors"
	"github.com/urfave/cli/v3"
)

// runGenerateCLI runs the generate subcommand through a root command shaped
// like the real goblog binary, so the shared flags are registered exactly as
// they are in production.
//
// These tests do not run in parallel: urfave/cli stores parsed values on the
// flag definitions, which the package-level GeneratorCommand shares.
func runGenerateCLI(t *testing.T, args ...string) error {
	t.Helper()

	cmd := &cli.Command{
		Name:                   "goblog",
		UseShortOptionHandling: true,
		Commands:               []*cli.Command{&GeneratorCommand},
		Flags:                  cliflags.Shared(),
	}
	return cmd.Run(context.Background(), append([]string{"goblog", "generate"}, args...))
}

// assertFatalFlagError asserts that err is a fatal FlagError whose message
// mentions every want substring.
func assertFatalFlagError(t *testing.T, err error, want ...string) {
	t.Helper()

	if err == nil {
		t.Fatal("expected an error, got nil")
	}

	var flagErr *inerrors.FlagError
	if !errors.As(err, &flagErr) {
		t.Fatalf("error %v is not a *errors.FlagError, so the CLI would exit 0", err)
	}
	if !flagErr.Type.IsFatalError() {
		t.Errorf("FlagError type = %v, want a fatal error so the CLI exits non-zero", flagErr.Type)
	}
	for _, w := range want {
		if !strings.Contains(flagErr.Error(), w) {
			t.Errorf("error %q does not mention %q", flagErr.Error(), w)
		}
	}
}

// generateDirs returns a posts directory holding one valid post and an empty
// output directory.
func generateDirs(t *testing.T) (postsDir, outputDir string) {
	t.Helper()

	tempDir := t.TempDir()
	postsDir = filepath.Join(tempDir, "posts")
	outputDir = filepath.Join(tempDir, "output")
	for _, dir := range []string{postsDir, outputDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("MkdirAll %s: %v", dir, err)
		}
	}

	post := `---
title: Flag Test
date: 2024-01-15
description: A test post
---

# Flag Test
`
	if err := os.WriteFile(filepath.Join(postsDir, "flag-test.md"), []byte(post), 0644); err != nil {
		t.Fatalf("write post: %v", err)
	}
	return postsDir, outputDir
}

// assertNothingGenerated asserts the output directory is still empty, i.e. the
// command failed before doing any generation work.
func assertNothingGenerated(t *testing.T, outputDir string) {
	t.Helper()

	entries, err := os.ReadDir(outputDir)
	if err != nil {
		t.Fatalf("ReadDir %s: %v", outputDir, err)
	}
	if len(entries) != 0 {
		t.Errorf("output directory is not empty (%d entries); generation ran despite the flag error", len(entries))
	}
}

// writeSeriesFile writes a series file next to the posts directory, listing the
// single post generateDirs creates, and returns its path.
func writeSeriesFile(t *testing.T, dir, body string) string {
	t.Helper()

	path := filepath.Join(dir, "series.yml")
	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		t.Fatalf("write series file: %v", err)
	}
	return path
}

// TestGenerateCLI_SeriesFileMissing asserts that a --series-file naming an
// unreadable path is a hard error rather than a silent fallback to series being
// disabled (case 15).
func TestGenerateCLI_SeriesFileMissing(t *testing.T) {
	postsDir, outputDir := generateDirs(t)
	missing := filepath.Join(t.TempDir(), "does-not-exist.yml")

	err := runGenerateCLI(t, "--"+cliflags.SeriesFileFlagName, missing, postsDir, outputDir)

	assertFatalFlagError(t, err, cliflags.SeriesFileFlagName, missing)
	assertNothingGenerated(t, outputDir)
}

// TestGenerateCLI_SeriesFileWritesPages asserts that a valid --series-file
// produces the series directory (case 21).
func TestGenerateCLI_SeriesFileWritesPages(t *testing.T) {
	postsDir, outputDir := generateDirs(t)
	seriesFile := writeSeriesFile(t, t.TempDir(), "series:\n  - name: Flags\n    posts:\n      - flag-test.md\n")

	err := runGenerateCLI(t, "--"+cliflags.SeriesFileFlagName, seriesFile, postsDir, outputDir)
	if err != nil {
		t.Fatalf("generate error = %v, want nil", err)
	}

	for _, file := range []string{"series/index.html", "series/flags.html"} {
		if _, err := os.Stat(filepath.Join(outputDir, filepath.FromSlash(file))); err != nil {
			t.Errorf("%s was not written: %v", file, err)
		}
	}
}

// TestGenerateCLI_NoSeriesFile asserts no series directory is written when the
// flag is absent (case 22).
func TestGenerateCLI_NoSeriesFile(t *testing.T) {
	postsDir, outputDir := generateDirs(t)

	err := runGenerateCLI(t, postsDir, outputDir)
	if err != nil {
		t.Fatalf("generate error = %v, want nil", err)
	}

	if _, err := os.Stat(filepath.Join(outputDir, "series")); !os.IsNotExist(err) {
		t.Errorf("series directory exists (stat error = %v) without --%s", err, cliflags.SeriesFileFlagName)
	}
}

// TestGenerateCLI_SeriesFileWithRawOutput asserts raw output writes no series
// directory even with a valid series file (case 23).
func TestGenerateCLI_SeriesFileWithRawOutput(t *testing.T) {
	postsDir, outputDir := generateDirs(t)
	seriesFile := writeSeriesFile(t, t.TempDir(), "series:\n  - name: Flags\n    posts:\n      - flag-test.md\n")

	err := runGenerateCLI(t,
		"--"+cliflags.SeriesFileFlagName, seriesFile,
		"--"+RawOutputFlagName,
		postsDir, outputDir,
	)
	if err != nil {
		t.Fatalf("generate error = %v, want nil", err)
	}

	if _, err := os.Stat(filepath.Join(outputDir, "series")); !os.IsNotExist(err) {
		t.Errorf("series directory exists (stat error = %v) in raw output mode", err)
	}
}

// TestGenerateCLI_InvalidSeriesFileFails asserts that a series file naming a
// post that does not exist fails generation, with the series and filename named
// (case 10, through the CLI).
func TestGenerateCLI_InvalidSeriesFileFails(t *testing.T) {
	postsDir, outputDir := generateDirs(t)
	seriesFile := writeSeriesFile(t, t.TempDir(), "series:\n  - name: Ghosts\n    posts:\n      - nope.md\n")

	err := runGenerateCLI(t, "--"+cliflags.SeriesFileFlagName, seriesFile, postsDir, outputDir)
	if err == nil {
		t.Fatal("generate error = nil, want an error for the invalid series file")
	}
	for _, want := range []string{"Ghosts", "nope.md"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not mention %q", err, want)
		}
	}
}

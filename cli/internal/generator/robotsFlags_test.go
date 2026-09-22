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

// TestGenerateCLI_RobotsFileWithDisableRobots asserts that the contradictory
// flag pair is rejected, with both flags named, before any generation happens.
func TestGenerateCLI_RobotsFileWithDisableRobots(t *testing.T) {
	postsDir, outputDir := generateDirs(t)

	robotsFile := filepath.Join(t.TempDir(), "robots.txt")
	if err := os.WriteFile(robotsFile, []byte("User-agent: *\nDisallow: /\n"), 0644); err != nil {
		t.Fatalf("write robots file: %v", err)
	}

	err := runGenerateCLI(t,
		"--"+cliflags.RobotsFileFlagName, robotsFile,
		"--"+cliflags.DisableRobotsFlagName,
		postsDir, outputDir,
	)

	assertFatalFlagError(t, err, cliflags.RobotsFileFlagName, cliflags.DisableRobotsFlagName)
	assertNothingGenerated(t, outputDir)
}

// TestGenerateCLI_RobotsFileMissing asserts that an unreadable --robots-file is
// a hard error rather than a silent fallback to the default rules.
func TestGenerateCLI_RobotsFileMissing(t *testing.T) {
	postsDir, outputDir := generateDirs(t)
	missing := filepath.Join(t.TempDir(), "does-not-exist.txt")

	err := runGenerateCLI(t, "--"+cliflags.RobotsFileFlagName, missing, postsDir, outputDir)

	assertFatalFlagError(t, err, cliflags.RobotsFileFlagName, missing)
	assertNothingGenerated(t, outputDir)
}

// TestGenerateCLI_RobotsFileApplied asserts that a readable --robots-file is
// written through to the generated robots.txt, replacing the default rules.
func TestGenerateCLI_RobotsFileApplied(t *testing.T) {
	postsDir, outputDir := generateDirs(t)

	robotsFile := filepath.Join(t.TempDir(), "robots.txt")
	if err := os.WriteFile(robotsFile, []byte("User-agent: BadBot\nDisallow: /\n"), 0644); err != nil {
		t.Fatalf("write robots file: %v", err)
	}

	err := runGenerateCLI(t,
		"--"+cliflags.BaseURLFlagName, "https://example.com",
		"--"+cliflags.RobotsFileFlagName, robotsFile,
		postsDir, outputDir,
	)
	if err != nil {
		t.Fatalf("generate error = %v, want nil", err)
	}

	got, err := os.ReadFile(filepath.Join(outputDir, "robots.txt"))
	if err != nil {
		t.Fatalf("read generated robots.txt: %v", err)
	}
	want := "User-agent: BadBot\nDisallow: /\n\nSitemap: https://example.com/sitemap.xml\n"
	if string(got) != want {
		t.Errorf("robots.txt =\n%q\nwant\n%q", got, want)
	}

	// The sitemap must be written alongside it, with .html paths matching the
	// files on disk.
	sitemap, err := os.ReadFile(filepath.Join(outputDir, "sitemap.xml"))
	if err != nil {
		t.Fatalf("read generated sitemap.xml: %v", err)
	}
	if want := "https://example.com/posts/flag-test.html"; !strings.Contains(string(sitemap), want) {
		t.Errorf("sitemap.xml does not contain %q:\n%s", want, sitemap)
	}
}

// TestGenerateCLI_DisableFlags asserts the two disable flags suppress their
// respective files.
func TestGenerateCLI_DisableFlags(t *testing.T) {
	postsDir, outputDir := generateDirs(t)

	err := runGenerateCLI(t,
		"--"+cliflags.BaseURLFlagName, "https://example.com",
		"--"+cliflags.DisableSitemapFlagName,
		"--"+cliflags.DisableRobotsFlagName,
		postsDir, outputDir,
	)
	if err != nil {
		t.Fatalf("generate error = %v, want nil", err)
	}

	for _, file := range []string{"sitemap.xml", "robots.txt"} {
		if _, err := os.Stat(filepath.Join(outputDir, file)); !os.IsNotExist(err) {
			t.Errorf("%s was written despite its disable flag", file)
		}
	}
}

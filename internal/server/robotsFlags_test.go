// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package server

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/harrydayexe/GoBlog/v2/internal/cliflags"
	inerrors "github.com/harrydayexe/GoBlog/v2/internal/errors"
	"github.com/urfave/cli/v3"
)

// runServeCLI runs the serve subcommand through a root command shaped like the
// real goblog binary, so the shared flags are registered exactly as they are in
// production.
//
// Every case here fails during flag validation, before the server is created,
// so Run returns rather than blocking on a listener.
//
// These tests do not run in parallel: urfave/cli stores parsed values on the
// flag definitions, which the package-level ServeCommand shares.
func runServeCLI(t *testing.T, args ...string) error {
	t.Helper()

	cmd := &cli.Command{
		Name:                   "goblog",
		UseShortOptionHandling: true,
		Commands:               []*cli.Command{&ServeCommand},
		Flags:                  cliflags.Shared(),
	}
	return cmd.Run(context.Background(), append([]string{"goblog", "serve"}, args...))
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

// TestServeCLI_RobotsFileWithDisableRobots asserts that the contradictory flag
// pair is rejected, with both flags named, before the server starts.
func TestServeCLI_RobotsFileWithDisableRobots(t *testing.T) {
	dir := t.TempDir()
	robotsFile := filepath.Join(dir, "robots.txt")
	if err := os.WriteFile(robotsFile, []byte("User-agent: *\nDisallow: /\n"), 0644); err != nil {
		t.Fatalf("write robots file: %v", err)
	}

	err := runServeCLI(t,
		"--"+cliflags.RobotsFileFlagName, robotsFile,
		"--"+cliflags.DisableRobotsFlagName,
		dir,
	)

	assertFatalFlagError(t, err, cliflags.RobotsFileFlagName, cliflags.DisableRobotsFlagName)
}

// TestServeCLI_RobotsFileMissing asserts that an unreadable --robots-file is a
// hard error rather than a silent fallback to the default rules.
func TestServeCLI_RobotsFileMissing(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "does-not-exist.txt")

	err := runServeCLI(t, "--"+cliflags.RobotsFileFlagName, missing, dir)

	assertFatalFlagError(t, err, cliflags.RobotsFileFlagName, missing)
}

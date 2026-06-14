// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestNewRootCommand_Properties(t *testing.T) {
	t.Parallel()

	cmd := newRootCommand()

	if cmd.Name != "goblog" {
		t.Errorf("root command Name = %q, want %q", cmd.Name, "goblog")
	}
	if !cmd.EnableShellCompletion {
		t.Error("root command EnableShellCompletion = false, want true")
	}
}

func TestCompletion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		shell  string
		marker string // string that must appear in the generated script
	}{
		{shell: "bash", marker: "complete"},
		{shell: "zsh", marker: "compdef"},
	}

	for _, tc := range tests {
		t.Run(tc.shell, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			cmd := newRootCommand()
			cmd.Writer = &buf

			err := cmd.Run(context.Background(), []string{"goblog", "completion", tc.shell})
			if err != nil {
				t.Fatalf("completion %s returned error: %v", tc.shell, err)
			}
			got := buf.String()
			if got == "" {
				t.Fatalf("completion %s produced empty output", tc.shell)
			}
			if !strings.Contains(got, tc.marker) {
				t.Errorf("completion %s output missing %q marker; got:\n%s", tc.shell, tc.marker, got)
			}
		})
	}
}

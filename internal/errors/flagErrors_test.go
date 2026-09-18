// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package errors

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// TestFlagError_Error tests the Error method for all error types.
func TestFlagError_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		err        *FlagError
		wantPrefix string
	}{
		{
			name:       "TypeHint formats with hint prefix",
			err:        &FlagError{Type: TypeHint, Msg: "consider setting --base-url"},
			wantPrefix: "hint:",
		},
		{
			name:       "TypeError formats with error prefix",
			err:        &FlagError{Type: TypeError, Msg: "--a cannot be used with --b"},
			wantPrefix: "error:",
		},
		{
			name:       "default case formats with just message",
			err:        &FlagError{Type: ErrorType(999), Msg: "some message"},
			wantPrefix: "some message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.err.Error()
			if !strings.HasPrefix(got, tt.wantPrefix) {
				t.Errorf("Error() = %q, want prefix %q", got, tt.wantPrefix)
			}
			if !strings.Contains(got, tt.err.Msg) {
				t.Errorf("Error() = %q, want to contain %q", got, tt.err.Msg)
			}
			if got := tt.err.HandlerString(); !strings.Contains(got, tt.err.Msg) {
				t.Errorf("HandlerString() = %q, want to contain %q", got, tt.err.Msg)
			}
		})
	}
}

// TestNewConflictingFlagsError asserts the error is fatal and names both flags,
// so the user can see which pair to resolve.
func TestNewConflictingFlagsError(t *testing.T) {
	t.Parallel()

	err := NewConflictingFlagsError("robots-file", "disable-robots", "contradictory intent")

	var flagErr *FlagError
	if !errors.As(err, &flagErr) {
		t.Fatalf("NewConflictingFlagsError() returned %T, want *FlagError", err)
	}
	if !flagErr.Type.IsFatalError() {
		t.Error("expected a fatal error so the CLI exits non-zero")
	}
	for _, want := range []string{"--robots-file", "--disable-robots", "contradictory intent"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("Error() = %q, want to contain %q", err.Error(), want)
		}
	}
}

// TestNewFlagFileError asserts the error is fatal and names both the flag and
// the offending path.
func TestNewFlagFileError(t *testing.T) {
	t.Parallel()

	err := NewFlagFileError("robots-file", "/tmp/nope.txt", fmt.Errorf("no such file or directory"))

	var flagErr *FlagError
	if !errors.As(err, &flagErr) {
		t.Fatalf("NewFlagFileError() returned %T, want *FlagError", err)
	}
	if !flagErr.Type.IsFatalError() {
		t.Error("expected a fatal error so the CLI exits non-zero")
	}
	for _, want := range []string{"--robots-file", "/tmp/nope.txt", "no such file or directory"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("Error() = %q, want to contain %q", err.Error(), want)
		}
	}
}

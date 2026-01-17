package errors

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// TestErrorType_IsFatalError tests the IsFatalError method for all error types.
func TestErrorType_IsFatalError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		errorType  ErrorType
		wantFatal  bool
	}{
		{
			name:      "TypeHint is not fatal",
			errorType: TypeHint,
			wantFatal: false,
		},
		{
			name:      "TypeError is fatal",
			errorType: TypeError,
			wantFatal: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.errorType.IsFatalError()
			if got != tt.wantFatal {
				t.Errorf("IsFatalError() = %v, want %v", got, tt.wantFatal)
			}
		})
	}
}

// TestInputDirectoryError_Error tests the Error method for all error types.
func TestInputDirectoryError_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		err       *InputDirectoryError
		wantPrefix string
		wantContains string
	}{
		{
			name: "TypeHint formats with hint prefix",
			err: &InputDirectoryError{
				Type: TypeHint,
				Msg:  "please specify a path",
			},
			wantPrefix: "hint:",
			wantContains: "please specify a path",
		},
		{
			name: "TypeError formats with error prefix",
			err: &InputDirectoryError{
				Type: TypeError,
				Msg:  "path is not a directory",
			},
			wantPrefix: "error:",
			wantContains: "path is not a directory",
		},
		{
			name: "default case formats with just message",
			err: &InputDirectoryError{
				Type: ErrorType(999), // Invalid type
				Msg:  "some message",
			},
			wantPrefix: "some message",
			wantContains: "some message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.err.Error()
			if !strings.HasPrefix(got, tt.wantPrefix) {
				t.Errorf("Error() = %q, want prefix %q", got, tt.wantPrefix)
			}
			if !strings.Contains(got, tt.wantContains) {
				t.Errorf("Error() = %q, want to contain %q", got, tt.wantContains)
			}
		})
	}
}

// TestInputDirectoryError_HandlerString tests the HandlerString method with color output.
func TestInputDirectoryError_HandlerString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		err          *InputDirectoryError
		wantContains string
	}{
		{
			name: "TypeHint produces yellow colored output",
			err: &InputDirectoryError{
				Type: TypeHint,
				Msg:  "please specify a path",
			},
			wantContains: "hint:",
		},
		{
			name: "TypeError produces red colored output",
			err: &InputDirectoryError{
				Type: TypeError,
				Msg:  "path is not a directory",
			},
			wantContains: "error:",
		},
		{
			name: "default case returns plain error message",
			err: &InputDirectoryError{
				Type: ErrorType(999),
				Msg:  "some message",
			},
			wantContains: "some message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.err.HandlerString()

			// Verify the string contains the expected message
			if !strings.Contains(got, tt.wantContains) {
				t.Errorf("HandlerString() = %q, want to contain %q", got, tt.wantContains)
			}

			// For TypeHint and TypeError, the color library may add ANSI codes
			// (depends on environment - TTY vs non-TTY)
			// In either case, verify the message content is present
			plainError := tt.err.Error()
			if tt.err.Type != TypeHint && tt.err.Type != TypeError {
				// Default case should match Error() exactly
				if got != plainError {
					t.Errorf("HandlerString() = %q, want %q", got, plainError)
				}
			}
			// For colored cases, just verify message is present (already checked in wantContains)
		})
	}
}

// TestNewDirectoryInaccessibleError tests the factory function for directory access errors.
func TestNewDirectoryInaccessibleError(t *testing.T) {
	t.Parallel()

	underlyingErr := fmt.Errorf("permission denied")
	err := NewDirectoryInaccessibleError(underlyingErr)

	// Verify it's an InputDirectoryError
	var inputErr *InputDirectoryError
	if !errors.As(err, &inputErr) {
		t.Fatalf("NewDirectoryInaccessibleError() returned %T, want *InputDirectoryError", err)
	}

	// Verify error type
	if inputErr.Type != TypeError {
		t.Errorf("Type = %v, want %v", inputErr.Type, TypeError)
	}

	// Verify it's a fatal error
	if !inputErr.Type.IsFatalError() {
		t.Error("expected fatal error")
	}

	// Verify message content
	errMsg := err.Error()
	if !strings.Contains(errMsg, "cannot access directory") {
		t.Errorf("Error() = %q, want to contain 'cannot access directory'", errMsg)
	}
	if !strings.Contains(errMsg, "permission denied") {
		t.Errorf("Error() = %q, want to contain underlying error message 'permission denied'", errMsg)
	}

	// Verify HandlerString returns the error message (may or may not have color codes depending on environment)
	handlerStr := inputErr.HandlerString()
	if len(handlerStr) == 0 {
		t.Error("HandlerString() returned empty string")
	}
}

// TestNewNotADirectoryError tests the factory function for not-a-directory errors.
func TestNewNotADirectoryError(t *testing.T) {
	t.Parallel()

	testPath := "/test/path/file.txt"
	err := NewNotADirectoryError(testPath)

	// Verify it's an InputDirectoryError
	var inputErr *InputDirectoryError
	if !errors.As(err, &inputErr) {
		t.Fatalf("NewNotADirectoryError() returned %T, want *InputDirectoryError", err)
	}

	// Verify error type
	if inputErr.Type != TypeError {
		t.Errorf("Type = %v, want %v", inputErr.Type, TypeError)
	}

	// Verify it's a fatal error
	if !inputErr.Type.IsFatalError() {
		t.Error("expected fatal error")
	}

	// Verify message content
	errMsg := err.Error()
	if !strings.Contains(errMsg, "not a directory") {
		t.Errorf("Error() = %q, want to contain 'not a directory'", errMsg)
	}
	if !strings.Contains(errMsg, testPath) {
		t.Errorf("Error() = %q, want to contain path %q", errMsg, testPath)
	}

	// Verify HandlerString returns the error message (may or may not have color codes depending on environment)
	handlerStr := inputErr.HandlerString()
	if len(handlerStr) == 0 {
		t.Error("HandlerString() returned empty string")
	}
}

// TestNewPathNotSpecifiedError tests the factory function for path-not-specified errors.
func TestNewPathNotSpecifiedError(t *testing.T) {
	t.Parallel()

	err := NewPathNotSpecifiedError()

	// Verify it's an InputDirectoryError
	var inputErr *InputDirectoryError
	if !errors.As(err, &inputErr) {
		t.Fatalf("NewPathNotSpecifiedError() returned %T, want *InputDirectoryError", err)
	}

	// Verify error type (should be TypeHint, not TypeError)
	if inputErr.Type != TypeHint {
		t.Errorf("Type = %v, want %v", inputErr.Type, TypeHint)
	}

	// Verify it's NOT a fatal error (it's a hint)
	if inputErr.Type.IsFatalError() {
		t.Error("expected non-fatal error (hint)")
	}

	// Verify message content
	errMsg := err.Error()
	if !strings.Contains(errMsg, "hint:") {
		t.Errorf("Error() = %q, want to contain 'hint:'", errMsg)
	}
	if !strings.Contains(errMsg, "please specify a path") {
		t.Errorf("Error() = %q, want to contain 'please specify a path'", errMsg)
	}

	// Verify HandlerString returns the error message (may or may not have color codes depending on environment)
	handlerStr := inputErr.HandlerString()
	if len(handlerStr) == 0 {
		t.Error("HandlerString() returned empty string")
	}
}

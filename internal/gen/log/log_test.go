package log

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

// TestLogger_Info tests the Info logging method
func TestLogger_Info(t *testing.T) {
	tests := []struct {
		name           string
		key            string
		verbose        bool
		format         string
		args           []any
		expectedOutput string
	}{
		{
			name:           "simple info message",
			key:            "TEST",
			verbose:        false,
			format:         "test message",
			args:           nil,
			expectedOutput: "INFO (TEST)",
		},
		{
			name:           "info message with arguments",
			key:            "APP",
			verbose:        false,
			format:         "loaded %d files",
			args:           []any{5},
			expectedOutput: "loaded 5 files",
		},
		{
			name:           "info message with multiple arguments",
			key:            "PARSE",
			verbose:        false,
			format:         "processed %s in %dms",
			args:           []any{"file.md", 123},
			expectedOutput: "processed file.md in 123ms",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			logger := NewTestLogger(tt.key, tt.verbose, &stdout, &stderr)

			logger.Info(tt.format, tt.args...)

			output := stdout.String()
			if !strings.Contains(output, tt.expectedOutput) {
				t.Errorf("expected output to contain %q, got %q", tt.expectedOutput, output)
			}

			// Verify nothing was written to stderr
			if stderr.Len() > 0 {
				t.Errorf("expected no stderr output, got %q", stderr.String())
			}
		})
	}
}

// TestLogger_Debug tests the Debug logging method
func TestLogger_Debug(t *testing.T) {
	tests := []struct {
		name           string
		key            string
		verbose        bool
		format         string
		args           []any
		expectOutput   bool
		expectedOutput string
	}{
		{
			name:           "debug message with verbose enabled",
			key:            "TEST",
			verbose:        true,
			format:         "debug info",
			args:           nil,
			expectOutput:   true,
			expectedOutput: "DEBUG (TEST)",
		},
		{
			name:         "debug message with verbose disabled",
			key:          "TEST",
			verbose:      false,
			format:       "debug info",
			args:         nil,
			expectOutput: false,
		},
		{
			name:           "debug message with args and verbose enabled",
			key:            "APP",
			verbose:        true,
			format:         "variable value: %v",
			args:           []any{42},
			expectOutput:   true,
			expectedOutput: "variable value: 42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			logger := NewTestLogger(tt.key, tt.verbose, &stdout, &stderr)

			logger.Debug(tt.format, tt.args...)

			output := stdout.String()
			if tt.expectOutput {
				if !strings.Contains(output, tt.expectedOutput) {
					t.Errorf("expected output to contain %q, got %q", tt.expectedOutput, output)
				}
			} else {
				if output != "" {
					t.Errorf("expected no output when verbose is disabled, got %q", output)
				}
			}

			// Verify nothing was written to stderr
			if stderr.Len() > 0 {
				t.Errorf("expected no stderr output, got %q", stderr.String())
			}
		})
	}
}

// TestLogger_Warn tests the Warn logging method
func TestLogger_Warn(t *testing.T) {
	tests := []struct {
		name           string
		key            string
		format         string
		args           []any
		expectedOutput string
	}{
		{
			name:           "simple warning",
			key:            "TEST",
			format:         "warning message",
			args:           nil,
			expectedOutput: "WARN (TEST)",
		},
		{
			name:           "warning with arguments",
			key:            "PARSER",
			format:         "skipping invalid file: %s",
			args:           []any{"test.txt"},
			expectedOutput: "skipping invalid file: test.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			logger := NewTestLogger(tt.key, false, &stdout, &stderr)

			logger.Warn(tt.format, tt.args...)

			output := stdout.String()
			if !strings.Contains(output, tt.expectedOutput) {
				t.Errorf("expected output to contain %q, got %q", tt.expectedOutput, output)
			}

			// Verify nothing was written to stderr
			if stderr.Len() > 0 {
				t.Errorf("expected no stderr output, got %q", stderr.String())
			}
		})
	}
}

// TestLogger_Error tests the Error logging method
func TestLogger_Error(t *testing.T) {
	tests := []struct {
		name           string
		key            string
		err            error
		expectedOutput string
	}{
		{
			name:           "simple error",
			key:            "TEST",
			err:            errors.New("test error"),
			expectedOutput: "test error",
		},
		{
			name:           "formatted error",
			key:            "CONFIG",
			err:            errors.New("failed to load config: file not found"),
			expectedOutput: "failed to load config: file not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			logger := NewTestLogger(tt.key, false, &stdout, &stderr)

			logger.Error(tt.err)

			// Error should write to stderr
			output := stderr.String()
			if !strings.Contains(output, tt.expectedOutput) {
				t.Errorf("expected stderr to contain %q, got %q", tt.expectedOutput, output)
			}
			if !strings.Contains(output, "ERROR ("+tt.key+")") {
				t.Errorf("expected stderr to contain ERROR (%s), got %q", tt.key, output)
			}

			// Verify nothing was written to stdout
			if stdout.Len() > 0 {
				t.Errorf("expected no stdout output, got %q", stdout.String())
			}
		})
	}
}

// TestNewCLILogger tests the CLI logger constructor
func TestNewCLILogger(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		verbose bool
	}{
		{
			name:    "verbose logger",
			key:     "TEST",
			verbose: true,
		},
		{
			name:    "non-verbose logger",
			key:     "APP",
			verbose: false,
		},
		{
			name:    "empty key",
			key:     "",
			verbose: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewCLILogger(tt.key, tt.verbose)

			if logger == nil {
				t.Fatal("expected logger to be non-nil")
			}

			if logger.key != tt.key {
				t.Errorf("expected key %q, got %q", tt.key, logger.key)
			}

			if logger.verbose != tt.verbose {
				t.Errorf("expected verbose %v, got %v", tt.verbose, logger.verbose)
			}

			if logger.width != "15" {
				t.Errorf("expected width to be 15, got %q", logger.width)
			}

			if logger.stdout == nil {
				t.Error("expected stdout to be set")
			}

			if logger.stderr == nil {
				t.Error("expected stderr to be set")
			}
		})
	}
}

// TestNewTestLogger tests the test logger constructor
func TestNewTestLogger(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	logger := NewTestLogger("TEST", true, &stdout, &stderr)

	if logger == nil {
		t.Fatal("expected logger to be non-nil")
	}

	if logger.key != "TEST" {
		t.Errorf("expected key TEST, got %q", logger.key)
	}

	if !logger.verbose {
		t.Error("expected verbose to be true")
	}

	if logger.stdout != &stdout {
		t.Error("expected stdout to match provided writer")
	}

	if logger.stderr != &stderr {
		t.Error("expected stderr to match provided writer")
	}
}

// TestLogger_Interface verifies CLILogger implements Logger interface
func TestLogger_Interface(t *testing.T) {
	var _ Logger = (*CLILogger)(nil)
}

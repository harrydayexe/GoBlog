package logger

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/fatih/color"
)

// Helper function to create a test slog.Record
func createRecord(level slog.Level, message string, attrs ...slog.Attr) slog.Record {
	return slog.NewRecord(time.Now(), level, message, 0)
}

// Helper function to create a record with attributes
func createRecordWithAttrs(level slog.Level, message string, attrs ...slog.Attr) slog.Record {
	r := slog.NewRecord(time.Now(), level, message, 0)
	r.AddAttrs(attrs...)
	return r
}

// TestNewDefaultCLIHandler tests basic handler creation
func TestNewDefaultCLIHandler(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	handler := NewDefaultCLIHandler(buf)

	if handler == nil {
		t.Fatal("expected handler to be created, got nil")
	}

	if handler.opts == nil {
		t.Error("expected opts to be initialized")
	}

	if handler.l == nil {
		t.Error("expected log.Logger to be initialized")
	}

	if handler.Handler == nil {
		t.Error("expected slog.Handler to be initialized")
	}
}

// TestNewDefaultCLIHandlerWithVerbosity tests handler creation with various log levels
func TestNewDefaultCLIHandlerWithVerbosity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		level slog.Level
	}{
		{"debug level", slog.LevelDebug},
		{"info level", slog.LevelInfo},
		{"warn level", slog.LevelWarn},
		{"error level", slog.LevelError},
		{"custom level below info", slog.LevelInfo - 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			handler := NewDefaultCLIHandlerWithVerbosity(buf, tt.level)

			if handler == nil {
				t.Fatal("expected handler to be created, got nil")
			}

			if handler.opts == nil {
				t.Fatal("expected opts to be initialized")
			}

			if handler.opts.Level.Level() != tt.level {
				t.Errorf("expected level %v, got %v", tt.level, handler.opts.Level.Level())
			}
		})
	}
}

// TestNewCLIHandlerWithOptions tests handler creation with custom options
func TestNewCLIHandlerWithOptions(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	customLevel := slog.LevelWarn
	opts := slog.HandlerOptions{
		Level: customLevel,
	}

	handler := NewCLIHandlerWithOptions(buf, opts)

	if handler == nil {
		t.Fatal("expected handler to be created, got nil")
	}

	if handler.Handler == nil {
		t.Error("expected slog.Handler to be initialized")
	}

	if handler.l == nil {
		t.Error("expected log.Logger to be initialized")
	}
}

// TestHandle_InfoPrefixVisibility tests INFO prefix visibility at different log levels
func TestHandle_InfoPrefixVisibility(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		handlerLevel   slog.Level
		recordLevel    slog.Level
		message        string
		expectInfoText bool // expect "INFO:" in output
	}{
		{
			name:           "debug mode shows INFO prefix",
			handlerLevel:   slog.LevelDebug,
			recordLevel:    slog.LevelInfo,
			message:        "test info message",
			expectInfoText: true,
		},
		{
			name:           "level below info shows INFO prefix",
			handlerLevel:   slog.LevelInfo - 1,
			recordLevel:    slog.LevelInfo,
			message:        "test info message",
			expectInfoText: true,
		},
		{
			name:           "info level hides INFO prefix",
			handlerLevel:   slog.LevelInfo,
			recordLevel:    slog.LevelInfo,
			message:        "test info message",
			expectInfoText: false,
		},
		{
			name:           "warn level hides INFO prefix",
			handlerLevel:   slog.LevelWarn,
			recordLevel:    slog.LevelInfo,
			message:        "test info message",
			expectInfoText: false,
		},
		{
			name:           "error level hides INFO prefix",
			handlerLevel:   slog.LevelError,
			recordLevel:    slog.LevelInfo,
			message:        "test info message",
			expectInfoText: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			handler := NewDefaultCLIHandlerWithVerbosity(buf, tt.handlerLevel)
			record := createRecord(tt.recordLevel, tt.message)

			err := handler.Handle(context.Background(), record)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			output := buf.String()

			// Check for the message itself
			if !strings.Contains(output, tt.message) {
				t.Errorf("expected message %q in output %q", tt.message, output)
			}

			// Check for INFO: prefix
			hasInfoPrefix := strings.Contains(output, "INFO:")
			if tt.expectInfoText && !hasInfoPrefix {
				t.Errorf("expected INFO: prefix in output, got: %q", output)
			}
			if !tt.expectInfoText && hasInfoPrefix {
				t.Errorf("expected no INFO: prefix in output, got: %q", output)
			}
		})
	}
}

// TestHandle_AllLevelPrefixes tests that all levels show their prefixes appropriately
func TestHandle_AllLevelPrefixes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		handlerLevel   slog.Level
		recordLevel    slog.Level
		message        string
		expectedPrefix string
	}{
		{
			name:           "debug level shows DEBUG prefix",
			handlerLevel:   slog.LevelDebug,
			recordLevel:    slog.LevelDebug,
			message:        "debug message",
			expectedPrefix: "DEBUG:",
		},
		{
			name:           "warn level shows WARN prefix",
			handlerLevel:   slog.LevelDebug,
			recordLevel:    slog.LevelWarn,
			message:        "warn message",
			expectedPrefix: "WARN:",
		},
		{
			name:           "error level shows ERROR prefix",
			handlerLevel:   slog.LevelDebug,
			recordLevel:    slog.LevelError,
			message:        "error message",
			expectedPrefix: "ERROR:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			handler := NewDefaultCLIHandlerWithVerbosity(buf, tt.handlerLevel)
			record := createRecord(tt.recordLevel, tt.message)

			err := handler.Handle(context.Background(), record)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			output := buf.String()

			if !strings.Contains(output, tt.expectedPrefix) {
				t.Errorf("expected prefix %q in output %q", tt.expectedPrefix, output)
			}

			if !strings.Contains(output, tt.message) {
				t.Errorf("expected message %q in output %q", tt.message, output)
			}
		})
	}
}

// TestHandle_AttributeVisibility tests attribute visibility at different log levels
func TestHandle_AttributeVisibility(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		handlerLevel     slog.Level
		recordLevel      slog.Level
		message          string
		attrs            []slog.Attr
		expectAttributes bool
	}{
		{
			name:         "debug level shows attributes",
			handlerLevel: slog.LevelDebug,
			recordLevel:  slog.LevelInfo,
			message:      "test message",
			attrs: []slog.Attr{
				slog.String("key1", "value1"),
				slog.Int("key2", 42),
			},
			expectAttributes: true,
		},
		{
			name:         "level below info shows attributes",
			handlerLevel: slog.LevelInfo - 1,
			recordLevel:  slog.LevelInfo,
			message:      "test message",
			attrs: []slog.Attr{
				slog.String("key1", "value1"),
			},
			expectAttributes: true,
		},
		{
			name:         "info level hides attributes",
			handlerLevel: slog.LevelInfo,
			recordLevel:  slog.LevelInfo,
			message:      "test message",
			attrs: []slog.Attr{
				slog.String("key1", "value1"),
			},
			expectAttributes: false,
		},
		{
			name:         "warn level hides attributes",
			handlerLevel: slog.LevelWarn,
			recordLevel:  slog.LevelWarn,
			message:      "test message",
			attrs: []slog.Attr{
				slog.String("key1", "value1"),
			},
			expectAttributes: false,
		},
		{
			name:         "error level hides attributes",
			handlerLevel: slog.LevelError,
			recordLevel:  slog.LevelError,
			message:      "test message",
			attrs: []slog.Attr{
				slog.String("key1", "value1"),
			},
			expectAttributes: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			handler := NewDefaultCLIHandlerWithVerbosity(buf, tt.handlerLevel)
			record := createRecordWithAttrs(tt.recordLevel, tt.message, tt.attrs...)

			err := handler.Handle(context.Background(), record)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			output := buf.String()

			hasAttributes := strings.Contains(output, "Attributes:")
			if tt.expectAttributes && !hasAttributes {
				t.Errorf("expected attributes in output, got: %q", output)
			}
			if !tt.expectAttributes && hasAttributes {
				t.Errorf("expected no attributes in output, got: %q", output)
			}

			// Verify attribute values are present when expected
			if tt.expectAttributes {
				for _, attr := range tt.attrs {
					if !strings.Contains(output, attr.Key) {
						t.Errorf("expected attribute key %q in output %q", attr.Key, output)
					}
				}
			}
		})
	}
}

// TestProcessAttributes_SingleAttribute tests processing a single attribute
func TestProcessAttributes_SingleAttribute(t *testing.T) {
	t.Parallel()

	record := createRecordWithAttrs(slog.LevelDebug, "test", slog.String("key", "value"))
	log := newLog()

	processAttributes(record, &log)

	if !strings.Contains(log.Attributes, "Attributes:") {
		t.Errorf("expected 'Attributes:' in output, got: %q", log.Attributes)
	}

	if !strings.Contains(log.Attributes, "key: value") {
		t.Errorf("expected 'key: value' in output, got: %q", log.Attributes)
	}

	if !strings.Contains(log.Attributes, " - ") {
		t.Errorf("expected ' - ' prefix in output, got: %q", log.Attributes)
	}
}

// TestProcessAttributes_MultipleAttributes tests processing multiple attributes
func TestProcessAttributes_MultipleAttributes(t *testing.T) {
	t.Parallel()

	record := createRecordWithAttrs(
		slog.LevelDebug,
		"test",
		slog.String("key1", "value1"),
		slog.Int("key2", 42),
		slog.Bool("key3", true),
	)
	log := newLog()

	processAttributes(record, &log)

	if !strings.Contains(log.Attributes, "Attributes:") {
		t.Errorf("expected 'Attributes:' in output, got: %q", log.Attributes)
	}

	// Check for each key-value pair
	expectedPairs := []string{
		"key1: value1",
		"key2: 42",
		"key3: true",
	}

	for _, pair := range expectedPairs {
		if !strings.Contains(log.Attributes, pair) {
			t.Errorf("expected %q in output, got: %q", pair, log.Attributes)
		}
	}

	// Count newlines to verify formatting
	newlineCount := strings.Count(log.Attributes, "\n")
	if newlineCount != 3 { // One newline per attribute
		t.Errorf("expected 3 newlines (one per attribute), got %d", newlineCount)
	}
}

// TestProcessAttributes_NoAttributes tests processing with no attributes
func TestProcessAttributes_NoAttributes(t *testing.T) {
	t.Parallel()

	record := createRecord(slog.LevelDebug, "test")
	log := newLog()

	processAttributes(record, &log)

	if log.Attributes != "" {
		t.Errorf("expected empty attributes, got: %q", log.Attributes)
	}
}

// TestProcessAttributes_VariousValueTypes tests various attribute value types
func TestProcessAttributes_VariousValueTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		attr     slog.Attr
		expected string
	}{
		{
			name:     "string value",
			attr:     slog.String("str", "hello"),
			expected: "str: hello",
		},
		{
			name:     "int value",
			attr:     slog.Int("num", 123),
			expected: "num: 123",
		},
		{
			name:     "bool true",
			attr:     slog.Bool("flag", true),
			expected: "flag: true",
		},
		{
			name:     "bool false",
			attr:     slog.Bool("flag", false),
			expected: "flag: false",
		},
		{
			name:     "int64 value",
			attr:     slog.Int64("big", 9223372036854775807),
			expected: "big: 9223372036854775807",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			record := createRecordWithAttrs(slog.LevelDebug, "test", tt.attr)
			log := newLog()

			processAttributes(record, &log)

			if !strings.Contains(log.Attributes, tt.expected) {
				t.Errorf("expected %q in output, got: %q", tt.expected, log.Attributes)
			}
		})
	}
}

// TestSetLevel_ColorFunctions tests color function assignment for each level
func TestSetLevel_ColorFunctions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		level         slog.Level
		showInfo      bool
		expectedFunc  func(string, ...any) string
		expectedLevel string
	}{
		{
			name:          "debug level - blue",
			level:         slog.LevelDebug,
			showInfo:      true,
			expectedFunc:  color.BlueString,
			expectedLevel: "DEBUG: ",
		},
		{
			name:          "info level with showInfo - white",
			level:         slog.LevelInfo,
			showInfo:      true,
			expectedFunc:  color.WhiteString,
			expectedLevel: "INFO: ",
		},
		{
			name:          "info level without showInfo - white, no prefix",
			level:         slog.LevelInfo,
			showInfo:      false,
			expectedFunc:  color.WhiteString,
			expectedLevel: "",
		},
		{
			name:          "warn level - yellow",
			level:         slog.LevelWarn,
			showInfo:      false,
			expectedFunc:  color.YellowString,
			expectedLevel: "WARN: ",
		},
		{
			name:          "error level - red",
			level:         slog.LevelError,
			showInfo:      false,
			expectedFunc:  color.RedString,
			expectedLevel: "ERROR: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			record := createRecord(tt.level, "test message")
			log := newLog()

			setLevel(record, &log, tt.showInfo)

			// Check level string
			if log.Level != tt.expectedLevel {
				t.Errorf("expected level %q, got %q", tt.expectedLevel, log.Level)
			}

			// Test color function by comparing output
			testStr := "test"
			expectedOutput := tt.expectedFunc(testStr)
			actualOutput := log.ColorFunc(testStr)

			if expectedOutput != actualOutput {
				t.Errorf("color function mismatch: expected %q, got %q", expectedOutput, actualOutput)
			}
		})
	}
}

// TestLog_String_WithAttributes tests Log.String() formatting with attributes
func TestLog_String_WithAttributes(t *testing.T) {
	t.Parallel()

	log := Log{
		ColorFunc:  color.WhiteString,
		Level:      "INFO: ",
		Message:    "test message",
		Attributes: "Attributes:\n - key: value",
	}

	output := log.String()

	// Check that output contains all parts in correct order
	expectedParts := []string{
		"INFO: ",
		"test message",
		"Attributes:",
		" - key: value",
	}

	for _, part := range expectedParts {
		if !strings.Contains(output, part) {
			t.Errorf("expected %q in output, got: %q", part, output)
		}
	}

	// Verify newline between message and attributes
	if !strings.Contains(output, "\n") {
		t.Error("expected newline between message and attributes")
	}
}

// TestLog_String_WithoutAttributes tests Log.String() formatting without attributes
func TestLog_String_WithoutAttributes(t *testing.T) {
	t.Parallel()

	log := Log{
		ColorFunc:  color.WhiteString,
		Level:      "WARN: ",
		Message:    "warning message",
		Attributes: "",
	}

	output := log.String()

	// Check that output contains level and message
	if !strings.Contains(output, "WARN: ") {
		t.Errorf("expected 'WARN: ' in output, got: %q", output)
	}

	if !strings.Contains(output, "warning message") {
		t.Errorf("expected 'warning message' in output, got: %q", output)
	}

	// Verify no trailing newline before empty attributes
	if strings.Contains(output, "warning message\n\n") {
		t.Errorf("unexpected double newline in output: %q", output)
	}
}

// TestLog_String_EmptyMessage tests Log.String() with empty message
func TestLog_String_EmptyMessage(t *testing.T) {
	t.Parallel()

	log := Log{
		ColorFunc:  color.WhiteString,
		Level:      "INFO: ",
		Message:    "",
		Attributes: "",
	}

	output := log.String()

	// Should still contain the level
	if !strings.Contains(output, "INFO: ") {
		t.Errorf("expected 'INFO: ' in output, got: %q", output)
	}
}

// TestNewLog_DefaultValues tests newLog() function
func TestNewLog_DefaultValues(t *testing.T) {
	t.Parallel()

	log := newLog()

	if log.Level != "" {
		t.Errorf("expected empty Level, got %q", log.Level)
	}

	if log.Message != "" {
		t.Errorf("expected empty Message, got %q", log.Message)
	}

	if log.Attributes != "" {
		t.Errorf("expected empty Attributes, got %q", log.Attributes)
	}

	// Test default color function
	testStr := "test"
	expectedOutput := color.WhiteString(testStr)
	actualOutput := log.ColorFunc(testStr)

	if expectedOutput != actualOutput {
		t.Errorf("expected default color function to be WhiteString")
	}
}

// TestHandle_CompleteFlow tests end-to-end scenarios
func TestHandle_CompleteFlow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		handlerLevel     slog.Level
		recordLevel      slog.Level
		message          string
		attrs            []slog.Attr
		expectPrefix     bool
		expectAttributes bool
	}{
		{
			name:         "debug mode - all visible",
			handlerLevel: slog.LevelDebug,
			recordLevel:  slog.LevelInfo,
			message:      "complete test",
			attrs: []slog.Attr{
				slog.String("file", "test.go"),
				slog.Int("line", 42),
			},
			expectPrefix:     true,
			expectAttributes: true,
		},
		{
			name:         "info mode - clean output",
			handlerLevel: slog.LevelInfo,
			recordLevel:  slog.LevelInfo,
			message:      "clean message",
			attrs: []slog.Attr{
				slog.String("hidden", "value"),
			},
			expectPrefix:     false,
			expectAttributes: false,
		},
		{
			name:         "warn mode - prefix visible, no attributes",
			handlerLevel: slog.LevelWarn,
			recordLevel:  slog.LevelWarn,
			message:      "warning message",
			attrs: []slog.Attr{
				slog.String("hidden", "value"),
			},
			expectPrefix:     true,
			expectAttributes: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			handler := NewDefaultCLIHandlerWithVerbosity(buf, tt.handlerLevel)
			record := createRecordWithAttrs(tt.recordLevel, tt.message, tt.attrs...)

			err := handler.Handle(context.Background(), record)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			output := buf.String()

			// Always expect the message
			if !strings.Contains(output, tt.message) {
				t.Errorf("expected message %q in output %q", tt.message, output)
			}

			// Check prefix expectation (for non-Info levels)
			if tt.recordLevel != slog.LevelInfo && tt.expectPrefix {
				levelStr := tt.recordLevel.String()
				if !strings.Contains(output, levelStr) {
					t.Errorf("expected level prefix %q in output %q", levelStr, output)
				}
			}

			// Check INFO prefix expectation
			if tt.recordLevel == slog.LevelInfo {
				hasInfoPrefix := strings.Contains(output, "INFO:")
				if tt.expectPrefix && !hasInfoPrefix {
					t.Errorf("expected INFO: prefix in output %q", output)
				}
				if !tt.expectPrefix && hasInfoPrefix {
					t.Errorf("unexpected INFO: prefix in output %q", output)
				}
			}

			// Check attributes expectation
			hasAttributes := strings.Contains(output, "Attributes:")
			if tt.expectAttributes && !hasAttributes {
				t.Errorf("expected attributes in output %q", output)
			}
			if !tt.expectAttributes && hasAttributes {
				t.Errorf("unexpected attributes in output %q", output)
			}
		})
	}
}

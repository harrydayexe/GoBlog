package utilities

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	inerrors "github.com/harrydayexe/GoBlog/v2/internal/errors"
)

// TestGetDirectoryFromInput_ValidAbsolutePath tests with a valid absolute path.
func TestGetDirectoryFromInput_ValidAbsolutePath(t *testing.T) {
	t.Parallel()

	// Create a temporary directory
	tempDir := t.TempDir()

	// Get the path
	result, err := GetDirectoryFromInput(tempDir, false)
	if err != nil {
		t.Fatalf("GetDirectoryFromInput() error = %v, want nil", err)
	}

	// Verify it's a valid path
	if result == "" {
		t.Fatal("GetDirectoryFromInput() returned empty string")
	}

	// Verify we can read the directory using os.DirFS
	testFS := os.DirFS(result)
	_, err = fs.ReadDir(testFS, ".")
	if err != nil {
		t.Errorf("fs.ReadDir() error = %v, want nil", err)
	}
}

// TestGetDirectoryFromInput_ValidRelativePath tests with a valid relative path.
func TestGetDirectoryFromInput_ValidRelativePath(t *testing.T) {
	t.Parallel()

	// Create a temporary directory and navigate to it
	tempDir := t.TempDir()
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() error = %v", err)
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("os.Chdir() error = %v", err)
	}

	// Create a subdirectory
	subDir := "testdir"
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("os.Mkdir() error = %v", err)
	}

	// Get the path using relative path
	result, err := GetDirectoryFromInput(subDir, false)
	if err != nil {
		t.Fatalf("GetDirectoryFromInput() error = %v, want nil", err)
	}

	// Verify it's a valid path
	if result == "" {
		t.Fatal("GetDirectoryFromInput() returned empty string")
	}
}

// TestGetDirectoryFromInput_EmptyPath tests with an empty path.
func TestGetDirectoryFromInput_EmptyPath(t *testing.T) {
	t.Parallel()

	_, err := GetDirectoryFromInput("", false)
	if err == nil {
		t.Fatal("GetDirectoryFromInput(\"\") expected error, got nil")
	}

	// Verify it's a PathNotSpecifiedError (TypeHint)
	var inputErr *inerrors.InputDirectoryError
	if !errors.As(err, &inputErr) {
		t.Fatalf("expected *InputDirectoryError, got %T", err)
	}

	if inputErr.Type != inerrors.TypeHint {
		t.Errorf("Type = %v, want %v", inputErr.Type, inerrors.TypeHint)
	}

	if !strings.Contains(err.Error(), "please specify a path") {
		t.Errorf("Error() = %q, want to contain 'please specify a path'", err.Error())
	}
}

// TestGetDirectoryFromInput_NonExistentPath tests with a non-existent path.
func TestGetDirectoryFromInput_NonExistentPath(t *testing.T) {
	t.Parallel()

	nonExistentPath := filepath.Join(t.TempDir(), "does-not-exist")

	_, err := GetDirectoryFromInput(nonExistentPath, false)
	if err == nil {
		t.Fatal("GetDirectoryFromInput() expected error for non-existent path, got nil")
	}

	// Verify it's a DirectoryInaccessibleError (TypeError)
	var inputErr *inerrors.InputDirectoryError
	if !errors.As(err, &inputErr) {
		t.Fatalf("expected *InputDirectoryError, got %T", err)
	}

	if inputErr.Type != inerrors.TypeError {
		t.Errorf("Type = %v, want %v", inputErr.Type, inerrors.TypeError)
	}

	if !strings.Contains(err.Error(), "cannot access directory") {
		t.Errorf("Error() = %q, want to contain 'cannot access directory'", err.Error())
	}
}

// TestGetDirectoryFromInput_FileNotDirectory tests with a file path instead of directory.
func TestGetDirectoryFromInput_FileNotDirectory(t *testing.T) {
	t.Parallel()

	// Create a temporary file
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "file.txt")
	if err := os.WriteFile(tempFile, []byte("test"), 0644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	_, err := GetDirectoryFromInput(tempFile, false)
	if err == nil {
		t.Fatal("GetDirectoryFromInput() expected error for file path, got nil")
	}

	// Verify it's a NotADirectoryError (TypeError)
	var inputErr *inerrors.InputDirectoryError
	if !errors.As(err, &inputErr) {
		t.Fatalf("expected *InputDirectoryError, got %T", err)
	}

	if inputErr.Type != inerrors.TypeError {
		t.Errorf("Type = %v, want %v", inputErr.Type, inerrors.TypeError)
	}

	if !strings.Contains(err.Error(), "not a directory") {
		t.Errorf("Error() = %q, want to contain 'not a directory'", err.Error())
	}

	if !strings.Contains(err.Error(), tempFile) {
		t.Errorf("Error() = %q, want to contain path %q", err.Error(), tempFile)
	}
}

// TestGetDirectoryFromInput_CurrentDirectory tests with current directory ".".
func TestGetDirectoryFromInput_CurrentDirectory(t *testing.T) {
	t.Parallel()

	result, err := GetDirectoryFromInput(".", false)
	if err != nil {
		t.Fatalf("GetDirectoryFromInput(\".\") error = %v, want nil", err)
	}

	if result == "" {
		t.Fatal("GetDirectoryFromInput(\".\") returned empty string")
	}

	// Verify we can read the directory using os.DirFS
	testFS := os.DirFS(result)
	_, err = fs.ReadDir(testFS, ".")
	if err != nil {
		t.Errorf("fs.ReadDir() error = %v, want nil", err)
	}
}

// TestGetDirectoryFromInput_ParentDirectory tests with parent directory "..".
func TestGetDirectoryFromInput_ParentDirectory(t *testing.T) {
	t.Parallel()

	result, err := GetDirectoryFromInput("..", false)
	if err != nil {
		t.Fatalf("GetDirectoryFromInput(\"..\") error = %v, want nil", err)
	}

	if result == "" {
		t.Fatal("GetDirectoryFromInput(\"..\") returned empty string")
	}

	// Verify we can read the directory using os.DirFS
	testFS := os.DirFS(result)
	_, err = fs.ReadDir(testFS, ".")
	if err != nil {
		t.Errorf("fs.ReadDir() error = %v, want nil", err)
	}
}

// TestCliErrorHandler_TypeError tests error detection and stderr routing for TypeError.
// NOTE: Cannot test os.Exit(1) in unit tests - only testing error detection and output routing.
func TestCliErrorHandler_TypeError(t *testing.T) {
	t.Parallel()

	err := inerrors.NewNotADirectoryError("/fake/path")

	// Verify error type detection
	var inputErr *inerrors.InputDirectoryError
	if !errors.As(err, &inputErr) {
		t.Fatal("expected InputDirectoryError")
	}

	if !inputErr.Type.IsFatalError() {
		t.Error("expected fatal error")
	}

	// Verify HandlerString format
	handlerStr := inputErr.HandlerString()
	if !strings.Contains(handlerStr, "not a directory") {
		t.Errorf("HandlerString() = %q, want to contain 'not a directory'", handlerStr)
	}

	// NOTE: We cannot test CliErrorHandler directly due to os.Exit(1)
	// The function would terminate the test process
	// Testing error detection and message formatting instead
}

// TestCliErrorHandler_TypeHint tests error detection and stdout routing for TypeHint.
func TestCliErrorHandler_TypeHint(t *testing.T) {
	// Cannot use t.Parallel() here because we're capturing stdout

	err := inerrors.NewPathNotSpecifiedError()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Call the handler (won't exit for TypeHint)
	CliErrorHandler(err)

	// Restore stdout and read output
	w.Close()
	os.Stdout = oldStdout

	var buf strings.Builder
	// Read from pipe
	done := make(chan bool)
	go func() {
		var readBuf [1024]byte
		for {
			n, err := r.Read(readBuf[:])
			if n > 0 {
				buf.Write(readBuf[:n])
			}
			if err != nil {
				break
			}
		}
		done <- true
	}()
	<-done

	output := buf.String()

	// Verify output contains hint message
	if !strings.Contains(output, "hint:") {
		t.Errorf("output = %q, want to contain 'hint:'", output)
	}

	if !strings.Contains(output, "please specify a path") {
		t.Errorf("output = %q, want to contain 'please specify a path'", output)
	}
}

// TestCliErrorHandler_NonInputDirectoryError verifies behavior with non-InputDirectoryError.
func TestCliErrorHandler_NonInputDirectoryError(t *testing.T) {
	// Cannot use t.Parallel() here because we're capturing stdout/stderr

	// Create a regular error (not InputDirectoryError)
	err := errors.New("some random error")

	// Capture both stdout and stderr
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr

	// Call the handler (should do nothing for non-InputDirectoryError)
	CliErrorHandler(err)

	// Restore stdout/stderr
	wOut.Close()
	wErr.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	// Read stdout
	var bufOut strings.Builder
	doneOut := make(chan bool)
	go func() {
		var readBuf [1024]byte
		for {
			n, err := rOut.Read(readBuf[:])
			if n > 0 {
				bufOut.Write(readBuf[:n])
			}
			if err != nil {
				break
			}
		}
		doneOut <- true
	}()
	<-doneOut

	// Read stderr
	var bufErr strings.Builder
	doneErr := make(chan bool)
	go func() {
		var readBuf [1024]byte
		for {
			n, err := rErr.Read(readBuf[:])
			if n > 0 {
				bufErr.Write(readBuf[:n])
			}
			if err != nil {
				break
			}
		}
		doneErr <- true
	}()
	<-doneErr

	// Verify no output (handler does nothing for non-InputDirectoryError)
	if bufOut.Len() > 0 {
		t.Errorf("expected no stdout output, got: %q", bufOut.String())
	}
	if bufErr.Len() > 0 {
		t.Errorf("expected no stderr output, got: %q", bufErr.String())
	}
}

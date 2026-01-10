package parser

import (
	"fmt"
	"strings"
)

// FileError represents an error that occurred while parsing a specific file.
type FileError struct {
	Path string // Path to the file that caused the error
	Err  error  // The underlying error
}

// Error implements the error interface.
func (fe FileError) Error() string {
	return fmt.Sprintf("%s: %v", fe.Path, fe.Err)
}

// Unwrap returns the underlying error for error wrapping support.
func (fe FileError) Unwrap() error {
	return fe.Err
}

// ParseErrors aggregates multiple parsing errors encountered during
// directory-wide parsing operations.
//
// ParseErrors is returned when one or more files fail to parse, but other
// files may have been successfully parsed. Check the Errors slice to see
// all individual failures.
type ParseErrors struct {
	Errors []FileError
}

// Error implements the error interface, returning a formatted multi-line
// error message listing all file errors.
func (pe ParseErrors) Error() string {
	if len(pe.Errors) == 0 {
		return "no parsing errors"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("failed to parse %d file(s):\n", len(pe.Errors)))
	for _, err := range pe.Errors {
		sb.WriteString(fmt.Sprintf("  - %s\n", err.Error()))
	}
	return sb.String()
}

// HasErrors returns true if there are any errors in the collection.
func (pe ParseErrors) HasErrors() bool {
	return len(pe.Errors) > 0
}

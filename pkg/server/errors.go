package server

import (
	"errors"
	"fmt"
)

// Error codes for BlogError
const (
	ErrCodeInvalidConfig    = "INVALID_CONFIG"
	ErrCodePostNotFound     = "POST_NOT_FOUND"
	ErrCodeIndexCorrupt     = "INDEX_CORRUPT"
	ErrCodeCacheFailure     = "CACHE_FAILURE"
	ErrCodeSearchFailure    = "SEARCH_FAILURE"
	ErrCodeContentLoad      = "CONTENT_LOAD_FAILURE"
	ErrCodeFileWatch        = "FILE_WATCH_FAILURE"
	ErrCodeServerNotStarted = "SERVER_NOT_STARTED"
)

// BlogError represents a structured error with context
type BlogError struct {
	Code    string // Machine-readable error code
	Message string // Human-readable message
	Err     error  // Underlying error
}

// Error implements the error interface
func (e *BlogError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *BlogError) Unwrap() error {
	return e.Err
}

// Is checks if the error matches the target
func (e *BlogError) Is(target error) bool {
	t, ok := target.(*BlogError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// NewBlogError creates a new BlogError
func NewBlogError(code, message string, err error) *BlogError {
	return &BlogError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Predefined errors for common cases
var (
	// ErrInvalidContentPath is returned when content path is invalid
	ErrInvalidContentPath = errors.New("invalid content path")

	// ErrInvalidCacheSize is returned when cache size is invalid
	ErrInvalidCacheSize = errors.New("cache size must be at least 1MB")

	// ErrInvalidPostsPerPage is returned when posts per page is invalid
	ErrInvalidPostsPerPage = errors.New("posts per page must be at least 1")

	// ErrServerNotStarted is returned when server operations are attempted before starting
	ErrServerNotStarted = errors.New("server not started")

	// ErrPostNotFound is returned when a post cannot be found
	ErrPostNotFound = &BlogError{
		Code:    ErrCodePostNotFound,
		Message: "post not found",
	}

	// ErrIndexCorrupt is returned when the search index is corrupted
	ErrIndexCorrupt = &BlogError{
		Code:    ErrCodeIndexCorrupt,
		Message: "search index is corrupted",
	}

	// ErrCacheFailure is returned when cache operations fail
	ErrCacheFailure = &BlogError{
		Code:    ErrCodeCacheFailure,
		Message: "cache operation failed",
	}
)

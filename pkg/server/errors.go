package server

import "errors"

var (
	// ErrInvalidContentPath is returned when content path is invalid
	ErrInvalidContentPath = errors.New("invalid content path")

	// ErrInvalidCacheSize is returned when cache size is invalid
	ErrInvalidCacheSize = errors.New("cache size must be at least 1MB")

	// ErrInvalidPostsPerPage is returned when posts per page is invalid
	ErrInvalidPostsPerPage = errors.New("posts per page must be at least 1")

	// ErrServerNotStarted is returned when server operations are attempted before starting
	ErrServerNotStarted = errors.New("server not started")
)

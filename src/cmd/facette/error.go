package main

import "errors"

var (
	// ErrInvalidFilter represents an invalid filter pattern error.
	ErrInvalidFilter = errors.New("invalid filter pattern")
	// ErrInvalidJSON represents an invalid JSON data error.
	ErrInvalidJSON = errors.New("invalid JSON data")
	// ErrInvalidParameter represents an invalid request parameter error.
	ErrInvalidParameter = errors.New("invalid request parameter")
	// ErrInvalidTimerange represents an invalid time range error.
	ErrInvalidTimerange = errors.New("invalid time range")
	// ErrReadOnly represents a read-only instance error.
	ErrReadOnly = errors.New("read-only instance")
	// ErrUnhandledError represents a service unhandled error.
	ErrUnhandledError = errors.New("an unhandled error has occurred")
	// ErrUnknownEndpoint represents an unknown endpoint error.
	ErrUnknownEndpoint = errors.New("unknown endpoint")
)

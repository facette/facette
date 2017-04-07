package main

import "errors"

var (
	// ErrInvalidJSON represents an invalid JSON data error.
	ErrInvalidJSON = errors.New("invalid JSON data")
	// ErrInvalidParameter represents an invalid request parameter error.
	ErrInvalidParameter = errors.New("invalid request parameter")
	// ErrInvalidTimerange represents an invalid time range error.
	ErrInvalidTimerange = errors.New("invalid time range")
	// ErrUnhandledError represents a service unhandled error.
	ErrUnhandledError = errors.New("an unhandled error has occurred")
	// ErrUnknownEndpoint represents an unknown endpoint error.
	ErrUnknownEndpoint = errors.New("unknown endpoint")
	// ErrUnsupportedType represents an unsupported content type error.
	ErrUnsupportedType = errors.New("unsupported content type")
)

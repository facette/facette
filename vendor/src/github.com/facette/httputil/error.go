package httputil

import "errors"

var (
	// ErrInvalidContentType represents an unsupported content type error.
	ErrInvalidContentType = errors.New("invalid content type")
	// ErrInvalidInterface represents an invalid interface error.
	ErrInvalidInterface = errors.New("invalid interface")
)

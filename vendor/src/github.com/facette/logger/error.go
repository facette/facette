package logger

import "errors"

var (
	// ErrInvalidFacility represents an invalid syslog facility error.
	ErrInvalidFacility = errors.New("invalid syslog facility")
	// ErrInvalidLevel represents an invalid logging level error.
	ErrInvalidLevel = errors.New("invalid logging level")
	// ErrUnsupportedBackend represents an unsupported backend error.
	ErrUnsupportedBackend = errors.New("unsupported backend")
)

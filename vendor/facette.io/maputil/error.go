package maputil

import "errors"

var (
	// ErrInvalidType represents an invalid configuration value type error.
	ErrInvalidType = errors.New("invalid value type")
	// ErrUnscannableValue represents an unscannable value error.
	ErrUnscannableValue = errors.New("unscannable value")
)

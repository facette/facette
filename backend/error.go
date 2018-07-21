package backend

import "errors"

var (
	// ErrEmptyGroup represents an empty group error.
	ErrEmptyGroup = errors.New("empty group")
	// ErrInvalidAlias represents an invalid alias error.
	ErrInvalidAlias = errors.New("invalid alias")
	// ErrInvalidID represents an invalid identifier error.
	ErrInvalidID = errors.New("invalid identifier")
	// ErrInvalidInterval represents an invalid interval error.
	ErrInvalidInterval = errors.New("invalid interval")
	// ErrInvalidName represents an invalid name error.
	ErrInvalidName = errors.New("invalid name")
	// ErrInvalidPattern represents an invalid pattern error.
	ErrInvalidPattern = errors.New("invalid pattern")
	// ErrInvalidPriority represents an invalid priority error.
	ErrInvalidPriority = errors.New("invalid priority")
	// ErrUnresolvableItem represents an unresolvable item error.
	ErrUnresolvableItem = errors.New("unresolvable item")
	// ErrUnscannableValue represents an unscannable value error.
	ErrUnscannableValue = errors.New("unscannable value")
)

package backend

import "errors"

var (
	// ErrInvalidAlias represents an invalid alias error.
	ErrInvalidAlias = errors.New("invalid alias")
	// ErrInvalidID represents an invalid identifier error.
	ErrInvalidID = errors.New("invalid identifier")
	// ErrInvalidInterval represents an invalid interval error.
	ErrInvalidInterval = errors.New("invalid interval")
	// ErrInvalidName represents an invalid name error.
	ErrInvalidName = errors.New("invalid name")
	// ErrInvalidPriority represents an invalid priority error.
	ErrInvalidPriority = errors.New("invalid priority")
	// ErrUnresolvableItem represents an unresolvable item error.
	ErrUnresolvableItem = errors.New("unresolvable item")
)

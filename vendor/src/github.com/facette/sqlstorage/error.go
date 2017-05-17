package sqlstorage

import "errors"

var (
	// ErrEmptyPassword represents an empty password error.
	ErrEmptyPassword = errors.New("empty password")
	// ErrItemConflict represents an item conflict error.
	ErrItemConflict = errors.New("item conflict")
	// ErrItemNotFound represents an item not found error.
	ErrItemNotFound = errors.New("item not found")
	// ErrMissingField represents a missing mandatory field error.
	ErrMissingField = errors.New("missing mandatory field")
	// ErrUnknownColumn represents an unknown column error.
	ErrUnknownColumn = errors.New("unknown column")
	// ErrUnknownReference represents an unknown reference error.
	ErrUnknownReference = errors.New("unknown reference")
	// ErrUnsupportedDriver represents an unsupported database driver error.
	ErrUnsupportedDriver = errors.New("unsupported driver")
)

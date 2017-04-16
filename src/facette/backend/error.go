package backend

import "errors"

var (
	// ErrEmptyGraph represents an invalid graph error.
	ErrEmptyGraph = errors.New("empty graph")
	// ErrEmptyGroup represents an invalid group error.
	ErrEmptyGroup = errors.New("empty group")
	// ErrIncompatibleAttributes represents an incompatible attributes error.
	ErrIncompatibleAttributes = errors.New("incompatible attributes")
	// ErrInvalidAlias represents an invalid alias error.
	ErrInvalidAlias = errors.New("invalid alias")
	// ErrInvalidID represents an invalid identifier error.
	ErrInvalidID = errors.New("invalid identifier")
	// ErrInvalidName represents an invalid name error.
	ErrInvalidName = errors.New("invalid name")
	// ErrInvalidParent represents an invalid parent error.
	ErrInvalidParent = errors.New("invalid parent")
	// ErrInvalidSlice represents an invalid slice error.
	ErrInvalidSlice = errors.New("invalid slice")
	// ErrItemNotExist represents a non-existent item error.
	ErrItemNotExist = errors.New("item not found")
	// ErrMissingBackendConfig represents a missing backend configuration error.
	ErrMissingBackendConfig = errors.New("missing backend configuration")
	// ErrMultipleBackendConfig represents a multiple backend configurations error.
	ErrMultipleBackendConfig = errors.New("too many backend configurations")
	// ErrInvalidBackendConfig represents an invalid backend configuration error.
	ErrInvalidBackendConfig = errors.New("invalid backend configuration")
	// ErrResourceConflict represents a backend resource conflict error.
	ErrResourceConflict = errors.New("a resource conflict occurred")
	// ErrResourceMissingDependency represents a backend resource missing dependency error.
	ErrResourceMissingDependency = errors.New("missing resource dependency")
	// ErrResourceMissingData represents a backend resource missing data error.
	ErrResourceMissingData = errors.New("missing resource data")
	// ErrUnknownColumn represents an unknown column error.
	ErrUnknownColumn = errors.New("unknown column")
)

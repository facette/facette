package catalog

import "errors"

var (
	// ErrEmptyOrigin represents an empty catalog origin error.
	ErrEmptyOrigin = errors.New("empty origin")
	// ErrEmptySource represents an empty catalog source error.
	ErrEmptySource = errors.New("empty source")
	// ErrEmptyMetric represents an empty catalog metric error.
	ErrEmptyMetric = errors.New("empty metric")
	// ErrUnknownOrigin represents an unknown catalog origin error.
	ErrUnknownOrigin = errors.New("unknown origin")
	// ErrUnknownSource represents an unknown catalog source error.
	ErrUnknownSource = errors.New("unknown source")
	// ErrUnknownMetric represents an unknown catalog metric error.
	ErrUnknownMetric = errors.New("unknown metric")
)

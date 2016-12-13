package catalog

import "errors"

var (
	// ErrUnknownOrigin represents an unknown catalog origin error.
	ErrUnknownOrigin = errors.New("unknown origin")
	// ErrUnknownSource represents an unknown catalog source error.
	ErrUnknownSource = errors.New("unknown source")
	// ErrUnknownMetric represents an unknown catalog metric error.
	ErrUnknownMetric = errors.New("unknown metric")
)

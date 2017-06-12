package series

import "errors"

var (
	// ErrInvalidSample represents an invalid series sample value error.
	ErrInvalidSample = errors.New("invalid sample value")
	// ErrEmptySeries represents a empty series list error.
	ErrEmptySeries = errors.New("no series provided")
	// ErrUnnormalizedSeries represents an unnormalized series list error.
	ErrUnnormalizedSeries = errors.New("unnormalized series")
)

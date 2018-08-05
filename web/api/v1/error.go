package v1

import "errors"

var (
	errInvalidFilter    = errors.New("invalid filter pattern")
	errInvalidJSON      = errors.New("invalid JSON data")
	errInvalidParameter = errors.New("invalid request parameter")
	errInvalidTimerange = errors.New("invalid time range")
	errReadOnly         = errors.New("read-only instance")
	errUnhandledError   = errors.New("an unhandled error has occurred")
	errUnknownEndpoint  = errors.New("unknown endpoint")
)

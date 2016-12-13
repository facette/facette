package connector

import (
	"errors"
	"fmt"
)

var (
	// ErrMissingMetricPattern represents a missing metric pattern keyword error.
	ErrMissingMetricPattern = errors.New("missing \"metric\" pattern keyword")
	// ErrMissingSourcePattern represents a missing source pattern keyword error.
	ErrMissingSourcePattern = errors.New("missing \"source\" pattern keyword")
	// ErrUnsupportedConnector represents an unsupported connector handler error.
	ErrUnsupportedConnector = errors.New("unsupported connector handler")
	// ErrUnknownSource represents an unknown source error.
	ErrUnknownSource = errors.New("unknown source")
	// ErrUnknownMetric represents an unknown metric error.
	ErrUnknownMetric = errors.New("unknown metric")
)

// ErrMissingConnectorSetting creates a new missing connector setting error.
func ErrMissingConnectorSetting(key string) error {
	return fmt.Errorf("missing mandatory %q connector setting", key)
}

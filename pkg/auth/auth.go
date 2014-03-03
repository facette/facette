// Package auth implements the authentication handlers.
package auth

import (
	"fmt"
)

var (
	// Handlers represents the list of available auth handlers.
	Handlers = make(map[string]func(map[string]string, int) interface{})
)

// Handler represents the main interface of a authentication handler.
type Handler interface {
	Authenticate(login, password string) bool
	Refresh() error
}

// NewAuth creates a new authentication handler instance.
func NewAuth(config map[string]string, debugLevel int) (Handler, error) {
	if _, ok := config["type"]; !ok || config["type"] == "none" {
		return nil, nil
	} else if _, ok := Handlers[config["type"]]; !ok {
		return nil, fmt.Errorf("unknown `%s' authentication handler", config["type"])
	}

	return Handlers[config["type"]](config, debugLevel).(Handler), nil
}

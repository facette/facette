package auth

import (
	"fmt"
)

var (
	// AuthHandlers represents the list of available auth handlers.
	AuthHandlers = make(map[string]func(map[string]string, int) AuthHandler)
)

// AuthHandler represents the main interface of auth handlers.
type AuthHandler interface {
	Authenticate(login, password string) bool
	Update() error
}

// NewAuth creates a new AuthHandler instance.
func NewAuth(config map[string]string, debugLevel int) (AuthHandler, error) {
	if _, ok := config["type"]; !ok {
		return nil, fmt.Errorf("missing authentication handler type")
	} else if _, ok := AuthHandlers[config["type"]]; !ok {
		return nil, fmt.Errorf("unknown `%s' authentication handler", config["type"])
	}

	return AuthHandlers[config["type"]](config, debugLevel), nil
}

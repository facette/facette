package auth

import (
	"facette/common"
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
func NewAuth(config *common.Config, debugLevel int) (AuthHandler, error) {
	if _, ok := config.Auth["type"]; !ok {
		return nil, fmt.Errorf("missing authentication handler type")
	} else if _, ok := AuthHandlers[config.Auth["type"]]; !ok {
		return nil, fmt.Errorf("unknown `%s' authentication handler", config.Auth["type"])
	}

	return AuthHandlers[config.Auth["type"]](config.Auth, debugLevel), nil
}

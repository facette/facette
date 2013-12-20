package auth

import (
	"facette/common"
	"facette/utils"
	"log"
)

/* Auth */

// Auth represents the main authentication method structure.
type Auth struct {
	Config     *common.Config
	Users      map[string]string
	debugLevel int
}

// Update updates the authentication base content.
func (auth *Auth) Update() error {
	var (
		err error
	)

	if auth.debugLevel > 0 {
		log.Printf("DEBUG: load authentication data from `%s' file...\n", auth.Config.AuthFile)
	}

	// Empty users map
	auth.Users = make(map[string]string)

	if _, err = utils.JSONLoad(auth.Config.AuthFile, &auth.Users); err != nil {
		log.Println("ERROR: " + err.Error())
		return err
	}

	return nil
}

// NewAuth creates a new instance of Auth.
func NewAuth(config *common.Config, debugLevel int) *Auth {
	// Create new Auth instance
	return &Auth{Config: config, debugLevel: debugLevel}
}

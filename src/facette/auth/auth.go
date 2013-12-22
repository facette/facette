package auth

import (
	"crypto/sha256"
	"encoding/base64"
	"facette/common"
	"facette/utils"
	"hash"
	"log"
)

/* Auth */

// Auth represents the main authentication method structure.
type Auth struct {
	Config     *common.Config
	Users      map[string]string
	debugLevel int
}

// Authenticate tries to match user login name with its password.
func (auth *Auth) Authenticate(login, password string) bool {
	if _, ok := auth.Users[login]; !ok {
		return false
	}

	return auth.Hash(password) == auth.Users[login]
}

// Hash generate the password hash.
func (auth *Auth) Hash(password string) string {
	var (
		hash hash.Hash
	)

	// Get password hash
	hash = sha256.New()
	hash.Write([]byte(password))

	return base64.StdEncoding.EncodeToString(hash.Sum(nil))
}

// Update updates the authentication base content.
func (auth *Auth) Update() error {
	var (
		err error
	)

	if auth.debugLevel > 0 {
		log.Printf("DEBUG: loading authentication data from `%s' file...\n", auth.Config.AuthFile)
	}

	// Empty users map
	auth.Users = make(map[string]string)

	_, err = utils.JSONLoad(auth.Config.AuthFile, &auth.Users)
	return err
}

// NewAuth creates a new instance of Auth.
func NewAuth(config *common.Config, debugLevel int) *Auth {
	// Create new Auth instance
	return &Auth{Config: config, debugLevel: debugLevel}
}

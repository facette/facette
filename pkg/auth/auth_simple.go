package auth

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"hash"
	"log"

	"github.com/facette/facette/pkg/utils"
)

// AuthSimpleHandler represents the main simple authentication method structure.
type AuthSimpleHandler struct {
	Config     map[string]string
	Users      map[string]string
	debugLevel int
}

// Authenticate tries to match user login name with its password.
func (handler *AuthSimpleHandler) Authenticate(login, password string) bool {
	if _, ok := handler.Users[login]; !ok {
		return false
	}

	return handler.Hash(password) == handler.Users[login]
}

// Hash generates the password hash.
func (handler *AuthSimpleHandler) Hash(password string) string {
	var (
		hash hash.Hash
	)

	// Get password hash
	hash = sha256.New()
	hash.Write([]byte(password))

	return base64.StdEncoding.EncodeToString(hash.Sum(nil))
}

// Update updates the authentication base content.
func (handler *AuthSimpleHandler) Update() error {
	var (
		err error
	)

	if _, ok := handler.Config["path"]; !ok {
		return fmt.Errorf("missing `path' authentication backend setting")
	}

	if handler.debugLevel > 0 {
		log.Printf("DEBUG: loading authentication data from `%s' file...\n", handler.Config["path"])
	}

	// Empty users map
	handler.Users = make(map[string]string)

	_, err = utils.JSONLoad(handler.Config["path"], &handler.Users)
	return err
}

func init() {
	AuthHandlers["simple"] = func(config map[string]string, debugLevel int) AuthHandler {
		return &AuthSimpleHandler{Config: config, debugLevel: debugLevel}
	}
}

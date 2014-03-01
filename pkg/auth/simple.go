package auth

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"

	"github.com/facette/facette/pkg/utils"
)

// SimpleHandler represents the main structue of the simple authentication handler.
type SimpleHandler struct {
	Config     map[string]string
	Users      map[string]string
	debugLevel int
}

// Authenticate tries to match user login name with its password.
func (handler *SimpleHandler) Authenticate(login, password string) bool {
	if _, ok := handler.Users[login]; !ok {
		return false
	}

	return handler.Hash(password) == handler.Users[login]
}

// Hash generates the password hash.
func (handler *SimpleHandler) Hash(password string) string {
	hash := sha256.New()
	hash.Write([]byte(password))

	return base64.StdEncoding.EncodeToString(hash.Sum(nil))
}

// Refresh updates the content of the authentication base.
func (handler *SimpleHandler) Refresh() error {
	if _, ok := handler.Config["path"]; !ok {
		return fmt.Errorf("missing `path' authentication handler setting")
	}

	if handler.debugLevel > 0 {
		log.Printf("DEBUG: loading authentication data from `%s' file...\n", handler.Config["path"])
	}

	handler.Users = make(map[string]string)

	_, err := utils.JSONLoad(handler.Config["path"], &handler.Users)
	return err
}

func init() {
	Handlers["simple"] = func(config map[string]string, debugLevel int) interface{} {
		return &SimpleHandler{Config: config, debugLevel: debugLevel}
	}
}

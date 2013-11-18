package server

import (
	"crypto/sha256"
	"encoding/base64"
	"hash"
	"net/http"
	"strings"
)

func (server *Server) handleAuth(writer http.ResponseWriter, request *http.Request) bool {
	var (
		authorization string
		chunks        []string
		data          []byte
		err           error
		hash          hash.Hash
	)

	authorization = request.Header.Get("Authorization")

	if strings.HasPrefix(authorization, "Basic ") {
		if data, err = base64.StdEncoding.DecodeString(authorization[6:]); err != nil {
			return false
		}

		if chunks = strings.Split(string(data), ":"); len(chunks) != 2 {
			return false
		}

		// Get password hash
		hash = sha256.New()
		hash.Write([]byte(chunks[1]))

		chunks[1] = base64.StdEncoding.EncodeToString(hash.Sum(nil))

		// Check for credentials match
		for login, password := range server.Auth.Users {
			if login != chunks[0] {
				continue
			} else if password != chunks[1] {
				break
			}

			return true
		}
	}

	writer.Header().Add("WWW-Authenticate", "Basic realm=\"Authorization Required\"")
	server.handleResponse(writer, http.StatusUnauthorized)

	return false
}

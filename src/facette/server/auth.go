package server

import (
	"encoding/base64"
	"net/http"
	"strings"
)

func (server *Server) handleAuth(writer http.ResponseWriter, request *http.Request) bool {
	var (
		authorization string
		chunks        []string
		data          []byte
		err           error
	)

	authorization = request.Header.Get("Authorization")

	if strings.HasPrefix(authorization, "Basic ") {
		if data, err = base64.StdEncoding.DecodeString(authorization[6:]); err != nil {
			return false
		}

		if chunks = strings.Split(string(data), ":"); len(chunks) != 2 {
			return false
		}

		if server.Auth.Authenticate(chunks[0], chunks[1]) {
			return true
		}
	}

	writer.Header().Add("WWW-Authenticate", "Basic realm=\"Authorization Required\"")
	server.handleResponse(writer, http.StatusUnauthorized)

	return false
}

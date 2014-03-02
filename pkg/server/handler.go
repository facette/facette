package server

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
)

func (server *Server) handleAuth(writer http.ResponseWriter, request *http.Request) bool {
	authorization := request.Header.Get("Authorization")

	if strings.HasPrefix(authorization, "Basic ") {
		data, err := base64.StdEncoding.DecodeString(authorization[6:])
		if err != nil {
			return false
		}

		chunks := strings.Split(string(data), ":")
		if len(chunks) != 2 {
			return false
		}

		if server.AuthHandler.Authenticate(chunks[0], chunks[1]) {
			return true
		}
	}

	writer.Header().Add("WWW-Authenticate", "Basic realm=\"Authorization Required\"")

	return false
}

type serverResponse struct {
	Message string `json:"message"`
}

func (server *Server) handleResponse(writer http.ResponseWriter, data interface{}, status int) {
	var err error

	output := make([]byte, 0)

	if data != nil {
		output, err = json.Marshal(data)
		if err != nil {
			server.handleResponse(writer, nil, http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	}

	writer.WriteHeader(status)

	if len(output) > 0 {
		writer.Write(output)
		writer.Write([]byte("\n"))
	}
}

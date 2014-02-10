package utils

import (
	"net/http"
	"strings"
)

// RequestGetContentType returns the HTTP request `Content-Type' header value.
func RequestGetContentType(request *http.Request) string {
	contentType := request.Header.Get("Content-Type")

	index := strings.Index(contentType, ";")
	if index != -1 {
		return contentType[:index]
	}

	return contentType
}

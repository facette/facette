package utils

import (
	"net/http"
	"strings"
)

// RequestGetContentType returns the HTTP request `Content-Type' header value.
func RequestGetContentType(request *http.Request) string {
	var (
		contentType string
		index       int
	)

	contentType = request.Header.Get("Content-Type")
	index = strings.Index(contentType, ";")

	if index != -1 {
		return contentType[:index]
	}

	return contentType
}

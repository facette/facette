package utils

import (
	"net/http"
	"strings"
)

// HTTPGetContentType returns the HTTP request `Content-Type' header value.
func HTTPGetContentType(input interface{}) string {
	var (
		header http.Header
	)

	switch input.(type) {
	case *http.Request:
		header = input.(*http.Request).Header
	case *http.Response:
		header = input.(*http.Response).Header
	default:
		return ""
	}

	contentType := header.Get("Content-Type")

	index := strings.Index(contentType, ";")
	if index != -1 {
		return contentType[:index]
	}

	return contentType
}

package httputil

import (
	"mime"
	"net/http"
)

// GetContentType returns the HTTP 'Content-Type' header value.
func GetContentType(v interface{}) (string, error) {
	var header http.Header

	switch v.(type) {
	case *http.Request:
		header = v.(*http.Request).Header
	case *http.Response:
		header = v.(*http.Response).Header
	default:
		return "", ErrInvalidInterface
	}

	ct, _, err := mime.ParseMediaType(header.Get("Content-Type"))
	if err != nil {
		return "", err
	}

	return ct, nil
}

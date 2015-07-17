package utils

import (
	"crypto/tls"
	"net"
	"net/http"
	"strings"
	"time"
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
	case http.ResponseWriter:
		header = input.(http.ResponseWriter).Header()
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

// HTTPGetURLBase returns the HTTP URL base based on available header values.
func HTTPGetURLBase(request *http.Request) string {
	base := request.Header.Get("X-Forwarded-Proto")

	if base == "" {
		base = "http"
	}

	base += "://" + request.Host

	return base
}

// NewHTTPClient returns a new HTTP client instance.
func NewHTTPClient(timeout int, insecureTLS bool) *http.Client {
	t := &http.Transport{
		Dial: (&net.Dialer{
			DualStack: true,
			Timeout:   time.Duration(timeout) * time.Second,
		}).Dial,
	}

	if insecureTLS {
		t.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	return &http.Client{Transport: t}
}

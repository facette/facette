package httputil

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

// NewClient creates a new HTTP client instance.
func NewClient(timeout time.Duration, dualStack, skipVerify bool) *http.Client {
	t := &http.Transport{
		Dial: (&net.Dialer{
			DualStack: dualStack,
			Timeout:   timeout,
		}).Dial,
	}

	if skipVerify {
		t.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	return &http.Client{Transport: t}
}

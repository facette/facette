package main

import (
	"io"
	"net/http"
	"time"

	"facette.io/facette/version"
	api "facette.io/facette/web/api/v1"
	"facette.io/httputil"
	"github.com/pkg/errors"
)

func apiRequest(method, endpoint string, headers map[string]string, r io.Reader, result interface{}) error {
	var (
		req *http.Request
		err error
	)

	req, err = http.NewRequest(method, cmd.Address+api.Prefix+endpoint, r)

	if err != nil {
		return errors.Wrap(err, "failed to create request")
	}

	req.Header.Add("User-Agent", "facettectl/"+version.Version)
	if r != nil {
		req.Header.Add("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Add(k, v)
	}

	// Execute request and check for error
	hc := httputil.NewClient(time.Duration(cmd.Timeout)*time.Second, true, false)

	resp, err := hc.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to execute request")
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		var msg api.Message
		if err := httputil.BindJSON(resp, &msg); err != nil {
			return errors.Wrap(err, "failed to unmarshal error JSON")
		}
		return errors.Errorf("failed to fetch data: %s", msg.Message)
	}

	// Only handle result data if receiver is not nil
	if result != nil {
		if err := httputil.BindJSON(resp, &result); err != nil {
			return errors.Wrap(err, "failed to unmarshal JSON")
		}
	}

	return nil
}

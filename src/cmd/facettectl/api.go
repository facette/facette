package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/facette/httputil"
	"github.com/pkg/errors"
)

const apiPrefix = "/api/v1"

type apiInfo struct {
	Version    string   `json:"version"`
	BuildDate  string   `json:"build_date"`
	BuildHash  string   `json:"build_hash"`
	Compiler   string   `json:"compiler"`
	Drivers    []string `json:"drivers"`
	Connectors []string `json:"connectors"`
}

type apiError struct {
	Message string `json:"message"`
}

func apiRequest(method, enpoint string, headers map[string]string, v, result interface{}) error {
	var (
		req *http.Request
		err error
	)

	if v != nil {
		var r io.Reader

		// Create reader from marshaled struct if input value if not already an io.Reader
		r, ok := v.(io.Reader)
		if !ok {
			data, err := json.Marshal(v)
			if err != nil {
				return errors.Wrap(err, "failed to marshal JSON")
			}

			r = bytes.NewReader(data)
		}

		req, err = http.NewRequest(method, cmd.Address+apiPrefix+enpoint, r)
	} else {
		req, err = http.NewRequest(method, cmd.Address+apiPrefix+enpoint, nil)
	}

	if err != nil {
		return errors.Wrap(err, "failed to create request")
	}

	req.Header.Add("User-Agent", "facettectl/"+version)
	if v != nil {
		req.Header.Add("Content-Type", "application/json")
	}
	if headers != nil {
		for k, v := range headers {
			req.Header.Add(k, v)
		}
	}

	// Execute request and check for error
	hc := httputil.NewClient(time.Duration(cmd.Timeout)*time.Second, true, false)

	resp, err := hc.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to execute request")
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		var v apiError

		if err := httputil.BindJSON(resp, &v); err != nil {
			return errors.Wrap(err, "failed to unmarshal error JSON")
		}

		return errors.Errorf("failed to fetch data: %s", v.Message)
	}

	// Only handle result data if receiver is not nil
	if result != nil {
		if err := httputil.BindJSON(resp, &result); err != nil {
			return errors.Wrap(err, "failed to unmarshal JSON")
		}
	}

	return nil
}

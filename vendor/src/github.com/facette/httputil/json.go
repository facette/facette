package httputil

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

// BindJSON binds JSON data received from request or response to an interface.
func BindJSON(v interface{}, out interface{}) error {
	var body io.ReadCloser

	if ct, _ := GetContentType(v); ct != "application/json" {
		return ErrInvalidContentType
	}

	switch v.(type) {
	case *http.Request:
		body = v.(*http.Request).Body
	case *http.Response:
		body = v.(*http.Response).Body
	default:
		return os.ErrInvalid
	}

	data, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, out); err != nil {
		return err
	}

	return nil
}

// WriteJSON marshals an interface to JSON data and writes it on an HTTP response writer.
func WriteJSON(rw http.ResponseWriter, v interface{}, code int) error {
	body, err := json.Marshal(v)
	if err != nil {
		return err
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(code)
	rw.Write(body)
	rw.Write([]byte("\n"))

	return nil
}

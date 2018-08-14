package jsonutil

import (
	"encoding/json"
	"io/ioutil"
)

// MarshalFile writes the JSON encoded version of v to file.
func MarshalFile(file string, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(file, data, 0600)
}

// UnmarshalFile parses the JSON-encoded data from file and stores the result in the value pointed to by v.
func UnmarshalFile(file string, v interface{}) error {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, v)
}
